package vibetrade

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
)

// Client is the HTTP client for connecting to the vibetrade backend API
type Client struct {
	baseURL    string
	httpClient *http.Client
	logger     *logrus.Logger
	userID     string
}

// Config holds the configuration for the vibetrade client
type Config struct {
	BaseURL string
	UserID  string
	Timeout time.Duration
}

// NewClient creates a new vibetrade API client
func NewClient(config *Config, logger *logrus.Logger) *Client {
	timeout := config.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	if logger == nil {
		logger = logrus.New()
	}

	return &Client{
		baseURL:    config.BaseURL,
		httpClient: &http.Client{Timeout: timeout},
		logger:     logger,
		userID:     config.UserID,
	}
}

// OptionChain represents an options chain for a symbol
type OptionChain struct {
	Symbol      string         `json:"symbol"`
	Expirations []string       `json:"expirations"`
	Strikes     []OptionStrike `json:"strikes"`
}

// OptionStrike represents a strike price with call and put information
type OptionStrike struct {
	Strike      decimal.Decimal `json:"strike"`
	CallBid     decimal.Decimal `json:"call_bid"`
	CallAsk     decimal.Decimal `json:"call_ask"`
	CallDelta   float64         `json:"call_delta"`
	CallVolume  int64           `json:"call_volume"`
	CallSymbol  string          `json:"call_symbol"`
	PutBid      decimal.Decimal `json:"put_bid"`
	PutAsk      decimal.Decimal `json:"put_ask"`
	PutDelta    float64         `json:"put_delta"`
	PutVolume   int64           `json:"put_volume"`
	PutSymbol   string          `json:"put_symbol"`
}

// OptionQuote represents a quote for an option contract
type OptionQuote struct {
	Symbol    string          `json:"symbol"`
	Bid       decimal.Decimal `json:"bid"`
	Ask       decimal.Decimal `json:"ask"`
	Last      decimal.Decimal `json:"last"`
	Volume    int64           `json:"volume"`
	OpenInt   int64           `json:"open_interest"`
	Delta     float64         `json:"delta"`
	Gamma     float64         `json:"gamma"`
	Theta     float64         `json:"theta"`
	Vega      float64         `json:"vega"`
	IV        float64         `json:"implied_volatility"`
	UpdatedAt time.Time       `json:"updated_at"`
}

// OptionPosition represents an options position
type OptionPosition struct {
	Symbol         string          `json:"symbol"`
	Quantity       decimal.Decimal `json:"quantity"`
	AveragePrice   decimal.Decimal `json:"average_price"`
	MarketValue    decimal.Decimal `json:"market_value"`
	UnrealizedPnL  decimal.Decimal `json:"unrealized_pnl"`
	Side           string          `json:"side"`
	AssetType      string          `json:"asset_type"`
	LastUpdated    time.Time       `json:"last_updated"`
}

// GetOptionsChain retrieves the options chain for a symbol
func (c *Client) GetOptionsChain(ctx context.Context, symbol string, daysToExpiry int) (*OptionChain, error) {
	params := url.Values{
		"symbol": {symbol},
	}
	if daysToExpiry > 0 {
		params.Set("days_to_expiry", fmt.Sprintf("%d", daysToExpiry))
	}

	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/api/options/chains?"+params.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-User-ID", c.userID)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: status %d", resp.StatusCode)
	}

	var chain OptionChain
	if err := json.NewDecoder(resp.Body).Decode(&chain); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &chain, nil
}

// GetOptionsQuotes retrieves quotes for multiple option symbols
func (c *Client) GetOptionsQuotes(ctx context.Context, symbols []string) (map[string]*OptionQuote, error) {
	params := url.Values{}
	for _, symbol := range symbols {
		params.Add("symbols", symbol)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/api/options/quotes?"+params.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-User-ID", c.userID)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: status %d", resp.StatusCode)
	}

	var quotes map[string]*OptionQuote
	if err := json.NewDecoder(resp.Body).Decode(&quotes); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return quotes, nil
}

// GetExpirations retrieves available expiration dates for a symbol
func (c *Client) GetExpirations(ctx context.Context, symbol string) ([]string, error) {
	params := url.Values{
		"symbol": {symbol},
	}

	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/api/options/expirations?"+params.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-User-ID", c.userID)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: status %d", resp.StatusCode)
	}

	var result struct {
		Symbol      string   `json:"symbol"`
		Expirations []string `json:"expirations"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Expirations, nil
}

// GetOptionsPositions retrieves options positions
func (c *Client) GetOptionsPositions(ctx context.Context, accountID string) ([]OptionPosition, error) {
	params := url.Values{
		"account_id": {accountID},
	}

	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/api/options/positions?"+params.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-User-ID", c.userID)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: status %d", resp.StatusCode)
	}

	var positions []OptionPosition
	if err := json.NewDecoder(resp.Body).Decode(&positions); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return positions, nil
}

// HealthCheck verifies the API is accessible
func (c *Client) HealthCheck(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/api/status", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API health check failed: status %d", resp.StatusCode)
	}

	return nil
}