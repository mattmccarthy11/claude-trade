package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	
	"vibetrade-claude/internal/ai_assistant"
)

type AIHandlers struct {
	aiAssistant   *ai_assistant.TradingAssistant
	dataAggregator *ai_assistant.MarketDataAggregator
	userStore     *snaptrade.FileUserStore
	encryptor     *snaptrade.Encryptor
}

type ConnectClaudeRequest struct {
	APIKey         string `json:"apiKey"`
	TurnstileToken string `json:"turnstileToken"`
}

type ClaudeConnectionStatus struct {
	IsConnected bool      `json:"isConnected"`
	ConnectedAt *time.Time `json:"connectedAt,omitempty"`
	LastUsedAt  *time.Time `json:"lastUsedAt,omitempty"`
}

type TradeRecommendationsResponse struct {
	Trades    []ai_assistant.TradeRecommendation `json:"trades"`
	Timestamp time.Time                          `json:"timestamp"`
	Message   string                             `json:"message,omitempty"`
}

func NewAIHandlers(userStore *snaptrade.FileUserStore, encryptor *snaptrade.Encryptor) *AIHandlers {
	return &AIHandlers{
		userStore: userStore,
		encryptor: encryptor,
	}
}

// HandleClaudeConnect handles connecting a Claude API key
func (h *AIHandlers) HandleClaudeConnect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ConnectClaudeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate turnstile token (simplified for now)
	if req.TurnstileToken == "" {
		sendJSONError(w, "Turnstile verification required", http.StatusBadRequest)
		return
	}

	// Get user ID from session
	userID := r.Context().Value("userID").(string)
	if userID == "" {
		sendJSONError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Encrypt the API key
	encryptedKey, err := h.encryptor.Encrypt(req.APIKey)
	if err != nil {
		h.logger.WithError(err).Error("Failed to encrypt API key")
		sendJSONError(w, "Failed to secure API key", http.StatusInternalServerError)
		return
	}

	// Store the encrypted key in user data
	user, err := h.userStore.GetUser(userID)
	if err != nil {
		sendJSONError(w, "User not found", http.StatusNotFound)
		return
	}

	// Add Claude connection to user metadata
	if user.Metadata == nil {
		user.Metadata = make(map[string]interface{})
	}
	
	user.Metadata["claude_api_key"] = encryptedKey
	user.Metadata["claude_connected_at"] = time.Now()
	
	if err := h.userStore.UpdateUser(user); err != nil {
		h.logger.WithError(err).Error("Failed to update user")
		sendJSONError(w, "Failed to save connection", http.StatusInternalServerError)
		return
	}

	// Initialize AI assistant with the API key
	h.aiAssistant = ai_assistant.NewTradingAssistant(req.APIKey)

	sendJSONResponse(w, map[string]interface{}{
		"success": true,
		"message": "Claude API connected successfully",
	})
}

// HandleClaudeStatus returns the connection status
func (h *AIHandlers) HandleClaudeStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.Context().Value("userID").(string)
	user, err := h.userStore.GetUser(userID)
	if err != nil {
		sendJSONResponse(w, ClaudeConnectionStatus{IsConnected: false})
		return
	}

	// Check if Claude is connected
	if apiKey, ok := user.Metadata["claude_api_key"].(string); ok && apiKey != "" {
		status := ClaudeConnectionStatus{
			IsConnected: true,
		}
		
		if connectedAt, ok := user.Metadata["claude_connected_at"].(time.Time); ok {
			status.ConnectedAt = &connectedAt
		}
		
		if lastUsed, ok := user.Metadata["claude_last_used_at"].(time.Time); ok {
			status.LastUsedAt = &lastUsed
		}
		
		sendJSONResponse(w, status)
	} else {
		sendJSONResponse(w, ClaudeConnectionStatus{IsConnected: false})
	}
}

// HandleGetRecommendations generates AI trade recommendations
func (h *AIHandlers) HandleGetRecommendations(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.Context().Value("userID").(string)
	user, err := h.userStore.GetUser(userID)
	if err != nil {
		sendJSONError(w, "User not found", http.StatusNotFound)
		return
	}

	// Get and decrypt API key
	encryptedKey, ok := user.Metadata["claude_api_key"].(string)
	if !ok || encryptedKey == "" {
		sendJSONError(w, "Claude not connected", http.StatusBadRequest)
		return
	}

	apiKey, err := h.encryptor.Decrypt(encryptedKey)
	if err != nil {
		h.logger.WithError(err).Error("Failed to decrypt API key")
		sendJSONError(w, "Failed to access Claude connection", http.StatusInternalServerError)
		return
	}

	// Initialize AI assistant if not already done
	if h.aiAssistant == nil {
		h.aiAssistant = ai_assistant.NewTradingAssistant(apiKey)
	}

	// Get user's portfolio data
	portfolio, err := h.getUserPortfolio(userID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get portfolio")
		portfolio = make(map[string]interface{}) // Use empty portfolio if error
	}

	// Aggregate market data for top symbols
	symbols := []string{"SPY", "QQQ", "AAPL", "MSFT", "NVDA", "TSLA", "AMD", "META"}
	marketData, err := h.dataAggregator.AggregateDataForSymbols(context.Background(), symbols)
	if err != nil {
		h.logger.WithError(err).Error("Failed to aggregate market data")
		sendJSONError(w, "Failed to fetch market data", http.StatusInternalServerError)
		return
	}

	// Get AI recommendations
	recommendations, err := h.aiAssistant.AnalyzeTrades(context.Background(), marketData, portfolio)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get AI recommendations")
		sendJSONError(w, "Failed to generate recommendations", http.StatusInternalServerError)
		return
	}

	// Update last used timestamp
	user.Metadata["claude_last_used_at"] = time.Now()
	h.userStore.UpdateUser(user)

	response := TradeRecommendationsResponse{
		Trades:    recommendations,
		Timestamp: time.Now(),
	}

	if len(recommendations) < 5 {
		response.Message = "Fewer than 5 trades meet the strict criteria. Market conditions may be unfavorable."
	}

	sendJSONResponse(w, response)
}

// HandleAnalyzeRisk analyzes risk for specific positions
func (h *AIHandlers) HandleAnalyzeRisk(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.Context().Value("userID").(string)
	user, err := h.userStore.GetUser(userID)
	if err != nil {
		sendJSONError(w, "User not found", http.StatusNotFound)
		return
	}

	// Check Claude connection
	encryptedKey, ok := user.Metadata["claude_api_key"].(string)
	if !ok || encryptedKey == "" {
		sendJSONError(w, "Claude not connected", http.StatusBadRequest)
		return
	}

	apiKey, err := h.encryptor.Decrypt(encryptedKey)
	if err != nil {
		sendJSONError(w, "Failed to access Claude connection", http.StatusInternalServerError)
		return
	}

	// Parse positions from request
	var req struct {
		Positions []map[string]interface{} `json:"positions"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Initialize AI assistant if needed
	if h.aiAssistant == nil {
		h.aiAssistant = ai_assistant.NewTradingAssistant(apiKey)
	}

	// Get risk analysis from AI
	analysis, err := h.aiAssistant.AnalyzeRisk(context.Background(), req.Positions)
	if err != nil {
		h.logger.WithError(err).Error("Failed to analyze risk")
		sendJSONError(w, "Failed to analyze risk", http.StatusInternalServerError)
		return
	}

	sendJSONResponse(w, map[string]interface{}{
		"analysis":  analysis,
		"timestamp": time.Now(),
	})
}

// HandleClaudeDisconnect removes the Claude connection
func (h *AIHandlers) HandleClaudeDisconnect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.Context().Value("userID").(string)
	user, err := h.userStore.GetUser(userID)
	if err != nil {
		sendJSONError(w, "User not found", http.StatusNotFound)
		return
	}

	// Remove Claude connection data
	delete(user.Metadata, "claude_api_key")
	delete(user.Metadata, "claude_connected_at")
	delete(user.Metadata, "claude_last_used_at")

	if err := h.userStore.UpdateUser(user); err != nil {
		h.logger.WithError(err).Error("Failed to update user")
		sendJSONError(w, "Failed to disconnect", http.StatusInternalServerError)
		return
	}

	// Clear AI assistant
	h.aiAssistant = nil

	sendJSONResponse(w, map[string]interface{}{
		"success": true,
		"message": "Claude disconnected successfully",
	})
}

func (h *AIHandlers) getUserPortfolio(userID string) (map[string]interface{}, error) {
	// This would fetch real portfolio data from your database or broker connection
	// For now, return mock data
	return map[string]interface{}{
		"cash_balance": 100000,
		"positions": []map[string]interface{}{
			{
				"symbol":   "SPY",
				"quantity": 100,
				"cost_basis": 450.00,
			},
		},
		"options": []map[string]interface{}{},
	}, nil
}

func sendJSONResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func sendJSONError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}