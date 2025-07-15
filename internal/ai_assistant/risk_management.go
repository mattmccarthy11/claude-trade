package ai_assistant

import (
	"fmt"
	"math"
)

// RiskManager enforces safety rules for AI-generated trades
type RiskManager struct {
	maxPortfolioRiskPercent  float64
	maxPositionSizePercent   float64
	maxDailyLossPercent      float64
	minProbabilityOfProfit   float64
	maxConcentrationPercent  float64
	requireManualApproval    bool
}

// RiskLimits defines user-configurable risk parameters
type RiskLimits struct {
	MaxPortfolioRisk  float64 `json:"max_portfolio_risk"`  // Max % of portfolio at risk
	MaxPositionSize   float64 `json:"max_position_size"`   // Max % per position
	MaxDailyLoss      float64 `json:"max_daily_loss"`      // Max daily loss %
	MinPOP            float64 `json:"min_pop"`             // Minimum probability of profit
	MaxConcentration  float64 `json:"max_concentration"`   // Max % in single symbol
}

// TradeValidation contains the result of risk validation
type TradeValidation struct {
	IsValid      bool     `json:"is_valid"`
	Violations   []string `json:"violations"`
	RiskScore    float64  `json:"risk_score"`
	RequiresApproval bool `json:"requires_approval"`
}

// PortfolioRiskMetrics contains overall portfolio risk measurements
type PortfolioRiskMetrics struct {
	TotalDelta      float64            `json:"total_delta"`
	TotalGamma      float64            `json:"total_gamma"`
	TotalVega       float64            `json:"total_vega"`
	TotalTheta      float64            `json:"total_theta"`
	ValueAtRisk     float64            `json:"value_at_risk"`
	MaxDrawdown     float64            `json:"max_drawdown"`
	Concentrations  map[string]float64 `json:"concentrations"`
	CorrelationRisk float64            `json:"correlation_risk"`
}

func NewRiskManager() *RiskManager {
	return &RiskManager{
		maxPortfolioRiskPercent:  2.0,   // 2% max portfolio risk
		maxPositionSizePercent:   0.5,   // 0.5% max per position
		maxDailyLossPercent:      1.0,   // 1% max daily loss
		minProbabilityOfProfit:   0.65,  // 65% minimum POP
		maxConcentrationPercent:  10.0,  // 10% max in single symbol
		requireManualApproval:    true,  // Always require manual approval
	}
}

// ValidateTrade checks if a trade recommendation meets risk criteria
func (rm *RiskManager) ValidateTrade(trade *TradeRecommendation, portfolio map[string]interface{}) *TradeValidation {
	validation := &TradeValidation{
		IsValid:          true,
		Violations:       []string{},
		RequiresApproval: rm.requireManualApproval,
	}

	// Get portfolio value
	portfolioValue, ok := portfolio["total_value"].(float64)
	if !ok || portfolioValue <= 0 {
		portfolioValue = 100000 // Default to 100k if not provided
	}

	// Check position size
	positionRisk := math.Abs(trade.MaxLoss)
	positionRiskPercent := (positionRisk / portfolioValue) * 100
	
	if positionRiskPercent > rm.maxPositionSizePercent {
		validation.IsValid = false
		validation.Violations = append(validation.Violations, 
			fmt.Sprintf("Position risk %.2f%% exceeds limit %.2f%%", 
				positionRiskPercent, rm.maxPositionSizePercent))
	}

	// Check probability of profit
	if trade.POP < rm.minProbabilityOfProfit {
		validation.IsValid = false
		validation.Violations = append(validation.Violations,
			fmt.Sprintf("POP %.2f%% below minimum %.2f%%",
				trade.POP*100, rm.minProbabilityOfProfit*100))
	}

	// Check risk/reward ratio
	riskRewardRatio := trade.MaxProfit / math.Abs(trade.MaxLoss)
	if riskRewardRatio < 0.33 { // Minimum 1:3 risk/reward
		validation.IsValid = false
		validation.Violations = append(validation.Violations,
			fmt.Sprintf("Risk/reward ratio %.2f below minimum 0.33", riskRewardRatio))
	}

	// Calculate risk score (0-100, lower is better)
	validation.RiskScore = rm.calculateRiskScore(trade, positionRiskPercent)

	return validation
}

// ValidatePortfolio checks overall portfolio risk
func (rm *RiskManager) ValidatePortfolio(trades []TradeRecommendation, portfolio map[string]interface{}) *TradeValidation {
	validation := &TradeValidation{
		IsValid:    true,
		Violations: []string{},
	}

	portfolioValue, ok := portfolio["total_value"].(float64)
	if !ok || portfolioValue <= 0 {
		portfolioValue = 100000
	}

	// Calculate total risk
	totalRisk := 0.0
	symbolRisk := make(map[string]float64)
	
	for _, trade := range trades {
		risk := math.Abs(trade.MaxLoss)
		totalRisk += risk
		symbolRisk[trade.Ticker] += risk
	}

	// Check total portfolio risk
	totalRiskPercent := (totalRisk / portfolioValue) * 100
	if totalRiskPercent > rm.maxPortfolioRiskPercent {
		validation.IsValid = false
		validation.Violations = append(validation.Violations,
			fmt.Sprintf("Total portfolio risk %.2f%% exceeds limit %.2f%%",
				totalRiskPercent, rm.maxPortfolioRiskPercent))
	}

	// Check concentration risk
	for symbol, risk := range symbolRisk {
		concentrationPercent := (risk / portfolioValue) * 100
		if concentrationPercent > rm.maxConcentrationPercent {
			validation.IsValid = false
			validation.Violations = append(validation.Violations,
				fmt.Sprintf("Concentration in %s (%.2f%%) exceeds limit %.2f%%",
					symbol, concentrationPercent, rm.maxConcentrationPercent))
		}
	}

	return validation
}

// CalculatePortfolioMetrics computes comprehensive risk metrics
func (rm *RiskManager) CalculatePortfolioMetrics(positions []map[string]interface{}) *PortfolioRiskMetrics {
	metrics := &PortfolioRiskMetrics{
		Concentrations: make(map[string]float64),
	}

	totalValue := 0.0
	symbolValues := make(map[string]float64)

	// Aggregate position values and Greeks
	for _, pos := range positions {
		if symbol, ok := pos["symbol"].(string); ok {
			if value, ok := pos["market_value"].(float64); ok {
				totalValue += value
				symbolValues[symbol] += value
			}

			// Sum up Greeks if available
			if delta, ok := pos["delta"].(float64); ok {
				metrics.TotalDelta += delta
			}
			if gamma, ok := pos["gamma"].(float64); ok {
				metrics.TotalGamma += gamma
			}
			if vega, ok := pos["vega"].(float64); ok {
				metrics.TotalVega += vega
			}
			if theta, ok := pos["theta"].(float64); ok {
				metrics.TotalTheta += theta
			}
		}
	}

	// Calculate concentrations
	if totalValue > 0 {
		for symbol, value := range symbolValues {
			metrics.Concentrations[symbol] = (value / totalValue) * 100
		}
	}

	// Calculate Value at Risk (simplified)
	// In production, use historical data and proper VaR calculation
	metrics.ValueAtRisk = totalValue * 0.02 // 2% VaR estimate

	// Estimate max drawdown based on position characteristics
	metrics.MaxDrawdown = rm.estimateMaxDrawdown(positions)

	// Calculate correlation risk (simplified)
	metrics.CorrelationRisk = rm.calculateCorrelationRisk(symbolValues)

	return metrics
}

func (rm *RiskManager) calculateRiskScore(trade *TradeRecommendation, positionRiskPercent float64) float64 {
	score := 0.0

	// POP component (0-30 points, higher POP = lower risk)
	popScore := (1 - trade.POP) * 30
	score += popScore

	// Position size component (0-30 points)
	sizeScore := (positionRiskPercent / rm.maxPositionSizePercent) * 30
	score += math.Min(sizeScore, 30)

	// Risk/reward component (0-20 points)
	rrRatio := trade.MaxProfit / math.Abs(trade.MaxLoss)
	rrScore := (1 / (rrRatio + 1)) * 20
	score += rrScore

	// Strategy risk component (0-20 points)
	strategyScore := rm.getStrategyRiskScore(trade.Strategy)
	score += strategyScore

	return math.Min(score, 100)
}

func (rm *RiskManager) getStrategyRiskScore(strategy string) float64 {
	// Assign risk scores to different strategies
	strategyRiskMap := map[string]float64{
		"covered call":       5.0,
		"cash secured put":   7.0,
		"credit spread":      10.0,
		"iron condor":        12.0,
		"butterfly":          15.0,
		"naked option":       20.0,
	}

	if score, ok := strategyRiskMap[strategy]; ok {
		return score
	}
	return 15.0 // Default medium-high risk
}

func (rm *RiskManager) estimateMaxDrawdown(positions []map[string]interface{}) float64 {
	// Simplified max drawdown estimation
	// In production, use Monte Carlo simulation or historical analysis
	totalRisk := 0.0
	
	for _, pos := range positions {
		if maxLoss, ok := pos["max_loss"].(float64); ok {
			totalRisk += math.Abs(maxLoss)
		}
	}
	
	return totalRisk * 1.5 // 1.5x multiplier for correlated moves
}

func (rm *RiskManager) calculateCorrelationRisk(symbolValues map[string]float64) float64 {
	// Simplified correlation risk based on sector concentration
	// In production, use actual correlation matrices
	
	techSymbols := []string{"AAPL", "MSFT", "NVDA", "AMD", "META", "GOOGL"}
	techConcentration := 0.0
	totalValue := 0.0
	
	for symbol, value := range symbolValues {
		totalValue += value
		for _, techSymbol := range techSymbols {
			if symbol == techSymbol {
				techConcentration += value
				break
			}
		}
	}
	
	if totalValue > 0 {
		techPercent := techConcentration / totalValue
		// Higher concentration = higher correlation risk
		return math.Min(techPercent * 100, 100)
	}
	
	return 0
}

// GetRiskDisclaimer returns appropriate risk disclaimers
func GetRiskDisclaimer() string {
	return `IMPORTANT RISK DISCLOSURE:

Options trading involves substantial risk and is not suitable for all investors. You may lose all or more than your initial investment. 

AI-GENERATED RECOMMENDATIONS:
- Are for educational purposes only
- Are NOT personalized financial advice
- Should NOT be acted upon without professional consultation
- May contain errors or outdated information
- Do not consider your personal financial situation

BEFORE TRADING:
- Understand all risks involved
- Consult with a licensed financial advisor
- Review the Characteristics and Risks of Standardized Options
- Only risk capital you can afford to lose

Past performance is not indicative of future results. The AI system's recommendations are based on historical data and mathematical models which may not predict future market movements accurately.

By using this AI trading assistant, you acknowledge that you understand these risks and that VibeTrade and its AI systems are not responsible for any trading losses you may incur.`
}