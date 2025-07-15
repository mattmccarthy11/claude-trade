package ai_assistant

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	
	"github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
	"github.com/alpacahq/alpaca-trade-api-go/v3/marketdata"
)

type MarketDataAggregator struct {
	alpacaClient *alpaca.Client
	marketData   *marketdata.Client
}

type AggregatedMarketData struct {
	Timestamp   time.Time                      `json:"timestamp"`
	Quotes      map[string]*Quote              `json:"quotes"`
	Options     map[string][]*OptionChain      `json:"options"`
	Technicals  map[string]*TechnicalIndicators `json:"technicals"`
	Fundamentals map[string]*Fundamentals       `json:"fundamentals"`
	MarketStats *MarketStatistics              `json:"market_stats"`
}

type Quote struct {
	Symbol    string    `json:"symbol"`
	Price     float64   `json:"price"`
	Bid       float64   `json:"bid"`
	Ask       float64   `json:"ask"`
	Volume    int64     `json:"volume"`
	Timestamp time.Time `json:"timestamp"`
}

type OptionChain struct {
	Symbol     string    `json:"symbol"`
	Strike     float64   `json:"strike"`
	Expiration string    `json:"expiration"`
	Type       string    `json:"type"` // "call" or "put"
	Bid        float64   `json:"bid"`
	Ask        float64   `json:"ask"`
	Last       float64   `json:"last"`
	Volume     int64     `json:"volume"`
	OpenInt    int64     `json:"open_interest"`
	IV         float64   `json:"implied_volatility"`
	Greeks     *Greeks   `json:"greeks"`
}

type Greeks struct {
	Delta float64 `json:"delta"`
	Gamma float64 `json:"gamma"`
	Theta float64 `json:"theta"`
	Vega  float64 `json:"vega"`
	Rho   float64 `json:"rho"`
}

type TechnicalIndicators struct {
	SMA50      float64 `json:"sma_50"`
	SMA200     float64 `json:"sma_200"`
	RSI        float64 `json:"rsi"`
	MACD       float64 `json:"macd"`
	MACDSignal float64 `json:"macd_signal"`
	ATR        float64 `json:"atr"`
	BollingerUpper float64 `json:"bollinger_upper"`
	BollingerLower float64 `json:"bollinger_lower"`
}

type Fundamentals struct {
	PE          float64 `json:"pe_ratio"`
	EPS         float64 `json:"eps"`
	Revenue     float64 `json:"revenue"`
	NetIncome   float64 `json:"net_income"`
	MarketCap   float64 `json:"market_cap"`
	DebtToEquity float64 `json:"debt_to_equity"`
}

type MarketStatistics struct {
	VIX         float64 `json:"vix"`
	SPYChange   float64 `json:"spy_change"`
	QQQChange   float64 `json:"qqq_change"`
	TenYearYield float64 `json:"ten_year_yield"`
	DollarIndex float64 `json:"dollar_index"`
}

func NewMarketDataAggregator(alpacaClient *alpaca.Client) *MarketDataAggregator {
	return &MarketDataAggregator{
		alpacaClient: alpacaClient,
		marketData:   marketdata.NewClient(marketdata.ClientOpts{}),
	}
}

func (mda *MarketDataAggregator) AggregateDataForSymbols(ctx context.Context, symbols []string) (*AggregatedMarketData, error) {
	aggregated := &AggregatedMarketData{
		Timestamp:    time.Now(),
		Quotes:       make(map[string]*Quote),
		Options:      make(map[string][]*OptionChain),
		Technicals:   make(map[string]*TechnicalIndicators),
		Fundamentals: make(map[string]*Fundamentals),
	}

	// Fetch quotes for all symbols
	for _, symbol := range symbols {
		quote, err := mda.fetchQuote(ctx, symbol)
		if err != nil {
			fmt.Printf("Error fetching quote for %s: %v\n", symbol, err)
			continue
		}
		aggregated.Quotes[symbol] = quote
	}

	// Fetch options chains
	for _, symbol := range symbols {
		chains, err := mda.fetchOptionChains(ctx, symbol)
		if err != nil {
			fmt.Printf("Error fetching options for %s: %v\n", symbol, err)
			continue
		}
		aggregated.Options[symbol] = chains
	}

	// Calculate technical indicators
	for _, symbol := range symbols {
		technicals, err := mda.calculateTechnicals(ctx, symbol)
		if err != nil {
			fmt.Printf("Error calculating technicals for %s: %v\n", symbol, err)
			continue
		}
		aggregated.Technicals[symbol] = technicals
	}

	// Fetch market statistics
	marketStats, err := mda.fetchMarketStats(ctx)
	if err != nil {
		fmt.Printf("Error fetching market stats: %v\n", err)
	} else {
		aggregated.MarketStats = marketStats
	}

	return aggregated, nil
}

func (mda *MarketDataAggregator) fetchQuote(ctx context.Context, symbol string) (*Quote, error) {
	// Fetch latest quote from Alpaca
	latestQuote, err := mda.marketData.GetLatestQuote(symbol, marketdata.GetLatestQuoteRequest{})
	if err != nil {
		return nil, err
	}

	return &Quote{
		Symbol:    symbol,
		Price:     (latestQuote.BidPrice + latestQuote.AskPrice) / 2,
		Bid:       latestQuote.BidPrice,
		Ask:       latestQuote.AskPrice,
		Volume:    int64(latestQuote.BidSize + latestQuote.AskSize),
		Timestamp: latestQuote.Timestamp,
	}, nil
}

func (mda *MarketDataAggregator) fetchOptionChains(ctx context.Context, symbol string) ([]*OptionChain, error) {
	// This is a placeholder - in production, you would fetch from a real options data provider
	// For now, return mock data
	chains := []*OptionChain{
		{
			Symbol:     symbol,
			Strike:     100.0,
			Expiration: time.Now().AddDate(0, 0, 7).Format("2006-01-02"),
			Type:       "call",
			Bid:        2.50,
			Ask:        2.60,
			Last:       2.55,
			Volume:     1000,
			OpenInt:    5000,
			IV:         0.25,
			Greeks: &Greeks{
				Delta: 0.50,
				Gamma: 0.02,
				Theta: -0.05,
				Vega:  0.10,
				Rho:   0.01,
			},
		},
	}
	
	return chains, nil
}

func (mda *MarketDataAggregator) calculateTechnicals(ctx context.Context, symbol string) (*TechnicalIndicators, error) {
	// Fetch historical bars for technical calculations
	bars, err := mda.marketData.GetBars(symbol, marketdata.GetBarsRequest{
		TimeFrame: marketdata.OneDay,
		Start:     time.Now().AddDate(0, -6, 0),
		End:       time.Now(),
	})
	if err != nil {
		return nil, err
	}

	// Calculate simple moving averages and other indicators
	// This is simplified - in production you'd use a proper technical analysis library
	tech := &TechnicalIndicators{}
	
	if len(bars) > 50 {
		// Calculate 50-day SMA
		sum := 0.0
		for i := len(bars) - 50; i < len(bars); i++ {
			sum += bars[i].Close
		}
		tech.SMA50 = sum / 50
	}
	
	// RSI calculation placeholder
	tech.RSI = 50.0 // Neutral RSI
	
	return tech, nil
}

func (mda *MarketDataAggregator) fetchMarketStats(ctx context.Context) (*MarketStatistics, error) {
	// Fetch market indicators
	stats := &MarketStatistics{}
	
	// Fetch VIX quote
	vixQuote, err := mda.fetchQuote(ctx, "VIX")
	if err == nil {
		stats.VIX = vixQuote.Price
	}
	
	// Fetch SPY and QQQ changes
	spyQuote, err := mda.fetchQuote(ctx, "SPY")
	if err == nil {
		// Calculate daily change percentage
		stats.SPYChange = 0.0 // Placeholder
	}
	
	qqqQuote, err := mda.fetchQuote(ctx, "QQQ")
	if err == nil {
		// Calculate daily change percentage
		stats.QQQChange = 0.0 // Placeholder
	}
	
	return stats, nil
}

func (mda *MarketDataAggregator) FormatForAI(data *AggregatedMarketData) (string, error) {
	// Format the aggregated data in a way that's optimal for AI analysis
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", err
	}
	
	return string(jsonData), nil
}