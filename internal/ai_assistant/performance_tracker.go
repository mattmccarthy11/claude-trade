package ai_assistant

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// PerformanceTracker tracks AI recommendation performance
type PerformanceTracker struct {
	mu          sync.RWMutex
	dataDir     string
	currentFile string
}

// RecommendationRecord stores a single AI recommendation
type RecommendationRecord struct {
	ID            string                `json:"id"`
	Timestamp     time.Time             `json:"timestamp"`
	Recommendation TradeRecommendation   `json:"recommendation"`
	Executed      bool                  `json:"executed"`
	ExecutionTime *time.Time            `json:"execution_time,omitempty"`
	ExitTime      *time.Time            `json:"exit_time,omitempty"`
	ActualProfit  *float64              `json:"actual_profit,omitempty"`
	Status        string                `json:"status"` // "pending", "executed", "closed", "expired"
}

// PerformanceMetrics contains aggregated performance statistics
type PerformanceMetrics struct {
	TotalRecommendations int                    `json:"total_recommendations"`
	ExecutedTrades       int                    `json:"executed_trades"`
	WinningTrades        int                    `json:"winning_trades"`
	LosingTrades         int                    `json:"losing_trades"`
	WinRate              float64                `json:"win_rate"`
	AverageReturn        float64                `json:"average_return"`
	TotalReturn          float64                `json:"total_return"`
	SharpeRatio          float64                `json:"sharpe_ratio"`
	MaxDrawdown          float64                `json:"max_drawdown"`
	ByStrategy           map[string]*StrategyMetrics `json:"by_strategy"`
	ByTimeframe          map[string]*TimeframeMetrics `json:"by_timeframe"`
}

// StrategyMetrics tracks performance by strategy type
type StrategyMetrics struct {
	Count        int     `json:"count"`
	WinRate      float64 `json:"win_rate"`
	AvgReturn    float64 `json:"avg_return"`
	TotalReturn  float64 `json:"total_return"`
}

// TimeframeMetrics tracks performance over time periods
type TimeframeMetrics struct {
	Period       string    `json:"period"`
	StartDate    time.Time `json:"start_date"`
	EndDate      time.Time `json:"end_date"`
	TradeCount   int       `json:"trade_count"`
	WinRate      float64   `json:"win_rate"`
	Return       float64   `json:"return"`
}

func NewPerformanceTracker(dataDir string) (*PerformanceTracker, error) {
	// Create data directory if it doesn't exist
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	// Create current month's file
	currentFile := filepath.Join(dataDir, fmt.Sprintf("ai_performance_%s.json", time.Now().Format("2006_01")))

	return &PerformanceTracker{
		dataDir:     dataDir,
		currentFile: currentFile,
	}, nil
}

// RecordRecommendation saves a new AI recommendation
func (pt *PerformanceTracker) RecordRecommendation(rec TradeRecommendation) (*RecommendationRecord, error) {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	record := &RecommendationRecord{
		ID:             generateRecordID(),
		Timestamp:      time.Now(),
		Recommendation: rec,
		Executed:       false,
		Status:         "pending",
	}

	// Load existing records
	records, err := pt.loadRecords()
	if err != nil {
		return nil, err
	}

	// Add new record
	records = append(records, record)

	// Save updated records
	if err := pt.saveRecords(records); err != nil {
		return nil, err
	}

	return record, nil
}

// UpdateExecution marks a recommendation as executed
func (pt *PerformanceTracker) UpdateExecution(recordID string, executed bool) error {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	records, err := pt.loadRecords()
	if err != nil {
		return err
	}

	for i, rec := range records {
		if rec.ID == recordID {
			records[i].Executed = executed
			if executed {
				now := time.Now()
				records[i].ExecutionTime = &now
				records[i].Status = "executed"
			}
			break
		}
	}

	return pt.saveRecords(records)
}

// UpdateTradeResult updates the outcome of a closed trade
func (pt *PerformanceTracker) UpdateTradeResult(recordID string, profit float64) error {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	records, err := pt.loadRecords()
	if err != nil {
		return err
	}

	for i, rec := range records {
		if rec.ID == recordID {
			now := time.Now()
			records[i].ExitTime = &now
			records[i].ActualProfit = &profit
			records[i].Status = "closed"
			break
		}
	}

	return pt.saveRecords(records)
}

// GetMetrics calculates performance metrics for a time period
func (pt *PerformanceTracker) GetMetrics(startDate, endDate time.Time) (*PerformanceMetrics, error) {
	pt.mu.RLock()
	defer pt.mu.RUnlock()

	records, err := pt.loadRecordsForPeriod(startDate, endDate)
	if err != nil {
		return nil, err
	}

	metrics := &PerformanceMetrics{
		TotalRecommendations: len(records),
		ByStrategy:          make(map[string]*StrategyMetrics),
		ByTimeframe:         make(map[string]*TimeframeMetrics),
	}

	returns := []float64{}
	
	for _, rec := range records {
		// Count executed trades
		if rec.Executed {
			metrics.ExecutedTrades++
			
			// Count wins/losses
			if rec.ActualProfit != nil {
				if *rec.ActualProfit > 0 {
					metrics.WinningTrades++
				} else if *rec.ActualProfit < 0 {
					metrics.LosingTrades++
				}
				
				metrics.TotalReturn += *rec.ActualProfit
				returns = append(returns, *rec.ActualProfit)
				
				// Track by strategy
				strategy := rec.Recommendation.Strategy
				if _, ok := metrics.ByStrategy[strategy]; !ok {
					metrics.ByStrategy[strategy] = &StrategyMetrics{}
				}
				metrics.ByStrategy[strategy].Count++
				metrics.ByStrategy[strategy].TotalReturn += *rec.ActualProfit
			}
		}
	}

	// Calculate aggregate metrics
	if metrics.ExecutedTrades > 0 {
		metrics.WinRate = float64(metrics.WinningTrades) / float64(metrics.ExecutedTrades) * 100
		metrics.AverageReturn = metrics.TotalReturn / float64(metrics.ExecutedTrades)
	}

	// Calculate Sharpe ratio (simplified)
	if len(returns) > 1 {
		metrics.SharpeRatio = calculateSharpeRatio(returns)
	}

	// Calculate max drawdown
	metrics.MaxDrawdown = calculateMaxDrawdown(returns)

	// Calculate strategy-specific metrics
	for strategy, stratMetrics := range metrics.ByStrategy {
		if stratMetrics.Count > 0 {
			stratMetrics.AvgReturn = stratMetrics.TotalReturn / float64(stratMetrics.Count)
			// Calculate win rate for this strategy
			wins := 0
			total := 0
			for _, rec := range records {
				if rec.Recommendation.Strategy == strategy && rec.Executed && rec.ActualProfit != nil {
					total++
					if *rec.ActualProfit > 0 {
						wins++
					}
				}
			}
			if total > 0 {
				stratMetrics.WinRate = float64(wins) / float64(total) * 100
			}
		}
	}

	// Add timeframe metrics
	pt.addTimeframeMetrics(metrics, records)

	return metrics, nil
}

// GetRecommendationHistory retrieves recent recommendations
func (pt *PerformanceTracker) GetRecommendationHistory(limit int) ([]*RecommendationRecord, error) {
	pt.mu.RLock()
	defer pt.mu.RUnlock()

	records, err := pt.loadRecords()
	if err != nil {
		return nil, err
	}

	// Sort by timestamp (newest first)
	// Return limited number of records
	if limit > 0 && limit < len(records) {
		return records[len(records)-limit:], nil
	}

	return records, nil
}

func (pt *PerformanceTracker) loadRecords() ([]*RecommendationRecord, error) {
	var records []*RecommendationRecord
	
	data, err := os.ReadFile(pt.currentFile)
	if err != nil {
		if os.IsNotExist(err) {
			return records, nil // Empty records if file doesn't exist
		}
		return nil, err
	}

	if err := json.Unmarshal(data, &records); err != nil {
		return nil, err
	}

	return records, nil
}

func (pt *PerformanceTracker) loadRecordsForPeriod(startDate, endDate time.Time) ([]*RecommendationRecord, error) {
	var allRecords []*RecommendationRecord

	// Load records from multiple monthly files if needed
	current := startDate
	for current.Before(endDate) || current.Equal(endDate) {
		monthFile := filepath.Join(pt.dataDir, fmt.Sprintf("ai_performance_%s.json", current.Format("2006_01")))
		
		if data, err := os.ReadFile(monthFile); err == nil {
			var records []*RecommendationRecord
			if err := json.Unmarshal(data, &records); err == nil {
				// Filter records within date range
				for _, rec := range records {
					if rec.Timestamp.After(startDate) && rec.Timestamp.Before(endDate) {
						allRecords = append(allRecords, rec)
					}
				}
			}
		}
		
		// Move to next month
		current = current.AddDate(0, 1, 0)
	}

	return allRecords, nil
}

func (pt *PerformanceTracker) saveRecords(records []*RecommendationRecord) error {
	data, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(pt.currentFile, data, 0644)
}

func (pt *PerformanceTracker) addTimeframeMetrics(metrics *PerformanceMetrics, records []*RecommendationRecord) {
	// Daily metrics
	dailyMetrics := pt.calculateTimeframeMetrics(records, "daily", 1)
	metrics.ByTimeframe["daily"] = dailyMetrics

	// Weekly metrics
	weeklyMetrics := pt.calculateTimeframeMetrics(records, "weekly", 7)
	metrics.ByTimeframe["weekly"] = weeklyMetrics

	// Monthly metrics
	monthlyMetrics := pt.calculateTimeframeMetrics(records, "monthly", 30)
	metrics.ByTimeframe["monthly"] = monthlyMetrics
}

func (pt *PerformanceTracker) calculateTimeframeMetrics(records []*RecommendationRecord, period string, days int) *TimeframeMetrics {
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -days)

	metrics := &TimeframeMetrics{
		Period:    period,
		StartDate: startDate,
		EndDate:   endDate,
	}

	wins := 0
	totalReturn := 0.0

	for _, rec := range records {
		if rec.Timestamp.After(startDate) && rec.Timestamp.Before(endDate) && rec.Executed && rec.ActualProfit != nil {
			metrics.TradeCount++
			totalReturn += *rec.ActualProfit
			if *rec.ActualProfit > 0 {
				wins++
			}
		}
	}

	if metrics.TradeCount > 0 {
		metrics.WinRate = float64(wins) / float64(metrics.TradeCount) * 100
		metrics.Return = totalReturn
	}

	return metrics
}

func calculateSharpeRatio(returns []float64) float64 {
	if len(returns) < 2 {
		return 0
	}

	// Calculate average return
	sum := 0.0
	for _, r := range returns {
		sum += r
	}
	avgReturn := sum / float64(len(returns))

	// Calculate standard deviation
	variance := 0.0
	for _, r := range returns {
		variance += (r - avgReturn) * (r - avgReturn)
	}
	stdDev := math.Sqrt(variance / float64(len(returns)-1))

	if stdDev == 0 {
		return 0
	}

	// Assume risk-free rate of 5% annually, scaled to daily
	riskFreeRate := 0.05 / 252

	return (avgReturn - riskFreeRate) / stdDev * math.Sqrt(252) // Annualized
}

func calculateMaxDrawdown(returns []float64) float64 {
	if len(returns) == 0 {
		return 0
	}

	cumulative := 0.0
	peak := 0.0
	maxDrawdown := 0.0

	for _, r := range returns {
		cumulative += r
		if cumulative > peak {
			peak = cumulative
		}
		drawdown := (peak - cumulative) / peak
		if drawdown > maxDrawdown {
			maxDrawdown = drawdown
		}
	}

	return maxDrawdown * 100 // Return as percentage
}

func generateRecordID() string {
	return fmt.Sprintf("rec_%d_%d", time.Now().Unix(), time.Now().Nanosecond())
}