package ai_assistant

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type TradingAssistant struct {
	claudeClient *ClaudeClient
	prompts      *PromptTemplates
}

type TradeRecommendation struct {
	Ticker    string  `json:"ticker"`
	Strategy  string  `json:"strategy"`
	Legs      string  `json:"legs"`
	Thesis    string  `json:"thesis"`
	POP       float64 `json:"pop"`       // Probability of Profit
	MaxLoss   float64 `json:"max_loss"`
	MaxProfit float64 `json:"max_profit"`
	Score     float64 `json:"score"`
}

type MarketData struct {
	Timestamp time.Time              `json:"timestamp"`
	Quotes    map[string]interface{} `json:"quotes"`
	Options   map[string]interface{} `json:"options"`
	IV        map[string]float64     `json:"iv"`
	Greeks    map[string]interface{} `json:"greeks"`
}

func NewTradingAssistant(apiKey string) *TradingAssistant {
	return &TradingAssistant{
		claudeClient: NewClaudeClient(apiKey),
		prompts:      NewPromptTemplates(),
	}
}

func (ta *TradingAssistant) AnalyzeTrades(ctx context.Context, marketData *MarketData, portfolio map[string]interface{}) ([]TradeRecommendation, error) {
	// Format market data for the prompt
	dataJSON, err := json.MarshalIndent(marketData, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("error formatting market data: %w", err)
	}

	portfolioJSON, err := json.MarshalIndent(portfolio, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("error formatting portfolio: %w", err)
	}

	// Build the user message with current data
	userMessage := fmt.Sprintf(
		"Current timestamp: %s\n\nPortfolio Data:\n%s\n\nMarket Data:\n%s\n\nPlease analyze and provide exactly 5 trade recommendations.",
		marketData.Timestamp.Format(time.RFC3339),
		string(portfolioJSON),
		string(dataJSON),
	)

	// Get recommendations from Claude
	response, err := ta.claudeClient.SendMessage(ctx, ta.prompts.TradingSystemPrompt, userMessage)
	if err != nil {
		return nil, fmt.Errorf("error getting AI recommendations: %w", err)
	}

	// Parse the response
	recommendations, err := ta.parseRecommendations(response)
	if err != nil {
		return nil, fmt.Errorf("error parsing recommendations: %w", err)
	}

	return recommendations, nil
}

func (ta *TradingAssistant) parseRecommendations(response string) ([]TradeRecommendation, error) {
	var recommendations []TradeRecommendation

	// Look for JSON in the response
	startIdx := strings.Index(response, "[")
	endIdx := strings.LastIndex(response, "]")
	
	if startIdx != -1 && endIdx != -1 && endIdx > startIdx {
		jsonStr := response[startIdx : endIdx+1]
		if err := json.Unmarshal([]byte(jsonStr), &recommendations); err != nil {
			// If JSON parsing fails, try to parse table format
			return ta.parseTableFormat(response)
		}
		return recommendations, nil
	}

	// Fall back to table parsing
	return ta.parseTableFormat(response)
}

func (ta *TradingAssistant) parseTableFormat(response string) ([]TradeRecommendation, error) {
	var recommendations []TradeRecommendation
	
	lines := strings.Split(response, "\n")
	inTable := false
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		// Skip empty lines and headers
		if line == "" || strings.Contains(line, "Ticker") || strings.Contains(line, "---") {
			if strings.Contains(line, "Ticker") {
				inTable = true
			}
			continue
		}
		
		if !inTable {
			continue
		}
		
		// Parse table row (assuming pipe-separated values)
		parts := strings.Split(line, "|")
		if len(parts) >= 5 {
			rec := TradeRecommendation{
				Ticker:   strings.TrimSpace(parts[0]),
				Strategy: strings.TrimSpace(parts[1]),
				Legs:     strings.TrimSpace(parts[2]),
				Thesis:   strings.TrimSpace(parts[3]),
			}
			
			// Try to parse POP
			if len(parts) > 4 {
				popStr := strings.TrimSpace(strings.TrimSuffix(parts[4], "%"))
				var pop float64
				fmt.Sscanf(popStr, "%f", &pop)
				rec.POP = pop / 100.0
			}
			
			recommendations = append(recommendations, rec)
		}
	}
	
	if len(recommendations) == 0 {
		return nil, fmt.Errorf("no recommendations found in response")
	}
	
	return recommendations, nil
}

func (ta *TradingAssistant) ExplainStrategy(ctx context.Context, strategy string) (string, error) {
	prompt := fmt.Sprintf("Explain the following options trading strategy in simple terms: %s\n\nInclude risk/reward profile and when to use it.", strategy)
	
	return ta.claudeClient.SendMessage(ctx, ta.prompts.EducationalPrompt, prompt)
}

func (ta *TradingAssistant) AnalyzeRisk(ctx context.Context, positions []map[string]interface{}) (string, error) {
	positionsJSON, err := json.MarshalIndent(positions, "", "  ")
	if err != nil {
		return "", fmt.Errorf("error formatting positions: %w", err)
	}
	
	prompt := fmt.Sprintf("Analyze the risk profile of these positions:\n\n%s\n\nProvide a comprehensive risk assessment.", string(positionsJSON))
	
	return ta.claudeClient.SendMessage(ctx, ta.prompts.RiskAnalysisPrompt, prompt)
}