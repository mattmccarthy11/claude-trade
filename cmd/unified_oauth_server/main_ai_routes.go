package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
	
	"vibetrade-claude/internal/ai_assistant"
)

// RegisterAIRoutes adds AI-related routes to the server
func (s *Server) RegisterAIRoutes(mux *http.ServeMux) {
	// Initialize AI handlers
	aiHandlers := NewAIHandlers(s.snapTradeUsers, s.consentManager.encryptor)
	
	// Claude Code connection endpoints
	mux.HandleFunc("/api/claude-code/connect", s.authenticateMiddleware(aiHandlers.HandleClaudeConnect))
	mux.HandleFunc("/api/claude-code/status", s.authenticateMiddleware(aiHandlers.HandleClaudeStatus))
	mux.HandleFunc("/api/claude-code/disconnect", s.authenticateMiddleware(aiHandlers.HandleClaudeDisconnect))
	
	// AI trading features
	mux.HandleFunc("/api/claude-code/recommendations", s.authenticateMiddleware(aiHandlers.HandleGetRecommendations))
	mux.HandleFunc("/api/claude-code/analyze-risk", s.authenticateMiddleware(aiHandlers.HandleAnalyzeRisk))
	
	// Educational endpoints
	mux.HandleFunc("/api/claude-code/explain-strategy", s.authenticateMiddleware(aiHandlers.HandleExplainStrategy))
	
	// Claude Code SDK endpoints
	claudeCodeServiceURL := os.Getenv("CLAUDE_CODE_SERVICE_URL")
	if claudeCodeServiceURL == "" {
		claudeCodeServiceURL = "http://localhost:3001"
	}
	claudeCodeHandlers := NewClaudeCodeHandlers(s.logger, claudeCodeServiceURL)
	claudeCodeHandlers.RegisterRoutes(mux)
	
	s.logger.Info("AI routes registered successfully")
}

// HandleExplainStrategy provides AI explanations for trading strategies
func (h *AIHandlers) HandleExplainStrategy(w http.ResponseWriter, r *http.Request) {
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

	// Parse strategy from request
	var req struct {
		Strategy string `json:"strategy"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Initialize AI assistant if needed
	if h.aiAssistant == nil {
		h.aiAssistant = ai_assistant.NewTradingAssistant(apiKey)
	}

	// Get explanation from AI
	explanation, err := h.aiAssistant.ExplainStrategy(r.Context(), req.Strategy)
	if err != nil {
		h.logger.WithError(err).Error("Failed to explain strategy")
		sendJSONError(w, "Failed to generate explanation", http.StatusInternalServerError)
		return
	}

	sendJSONResponse(w, map[string]interface{}{
		"strategy":    req.Strategy,
		"explanation": explanation,
		"timestamp":   time.Now(),
	})
}

// authenticateMiddleware checks for valid session/JWT token
func (s *Server) authenticateMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			sendJSONError(w, "Missing authorization header", http.StatusUnauthorized)
			return
		}

		// Extract bearer token
		const bearerPrefix = "Bearer "
		if !strings.HasPrefix(authHeader, bearerPrefix) {
			sendJSONError(w, "Invalid authorization format", http.StatusUnauthorized)
			return
		}

		token := authHeader[len(bearerPrefix):]
		
		// Validate token and get user ID
		userID, err := s.validateToken(token)
		if err != nil {
			sendJSONError(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Add user ID to request context
		ctx := context.WithValue(r.Context(), "userID", userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func (s *Server) validateToken(token string) (string, error) {
	// This would validate the JWT token and return the user ID
	// For now, return a mock user ID
	if token == "" {
		return "", fmt.Errorf("empty token")
	}
	
	// In production, decode JWT and extract user ID
	return "anonymous-user-123", nil
}