package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

// ClaudeCodeHandlers handles Claude Code related endpoints
type ClaudeCodeHandlers struct {
	logger      *logrus.Logger
	serviceURL  string
	httpClient  *http.Client
}

// NewClaudeCodeHandlers creates a new instance
func NewClaudeCodeHandlers(logger *logrus.Logger, serviceURL string) *ClaudeCodeHandlers {
	if serviceURL == "" {
		serviceURL = "http://localhost:3001"
	}

	return &ClaudeCodeHandlers{
		logger:     logger,
		serviceURL: serviceURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// RegisterRoutes registers all Claude Code routes
func (h *ClaudeCodeHandlers) RegisterRoutes(mux *http.ServeMux) {
	// Chat endpoint for simple HTTP-based chat
	mux.HandleFunc("/api/claude-code/chat", h.authMiddleware(h.handleChat))
	
	// Status endpoint
	mux.HandleFunc("/api/claude-code/status", h.authMiddleware(h.handleStatus))
	
	// WebSocket proxy (future implementation)
	mux.HandleFunc("/api/claude-code/ws", h.authMiddleware(h.handleWebSocket))
}

// authMiddleware validates authentication
func (h *ClaudeCodeHandlers) authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// For now, just check if user is authenticated via cookie
		// In production, implement proper auth check
		userID := r.Header.Get("X-User-ID")
		if userID == "" {
			// Try to get from cookie or session
			userID = "default-user" // Placeholder
		}
		
		ctx := context.WithValue(r.Context(), "userID", userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// ChatRequest represents a chat message request
type ChatRequest struct {
	Message   string `json:"message"`
	SessionID string `json:"sessionId"`
}

// ChatResponse represents a chat message response
type ChatResponse struct {
	Response  string `json:"response"`
	SessionID string `json:"sessionId"`
}

// handleChat processes chat messages
func (h *ClaudeCodeHandlers) handleChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	userID := r.Context().Value("userID").(string)
	
	// Forward request to Claude Code service
	serviceReq := map[string]interface{}{
		"type":      "chat",
		"content":   req.Message,
		"sessionId": req.SessionID,
		"userId":    userID,
	}

	reqBody, err := json.Marshal(serviceReq)
	if err != nil {
		h.logger.WithError(err).Error("Failed to marshal request")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Call Claude Code service
	resp, err := h.httpClient.Post(
		h.serviceURL+"/api/chat",
		"application/json",
		bytes.NewReader(reqBody),
	)
	if err != nil {
		h.logger.WithError(err).Error("Failed to call Claude Code service")
		
		// For now, return a mock response if service is not available
		mockResponse := ChatResponse{
			Response:  "I'm currently connecting to the Claude service. Please try again in a moment.",
			SessionID: req.SessionID,
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
		return
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		h.logger.WithError(err).Error("Failed to read response")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Forward response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	w.Write(body)
}

// handleStatus returns the Claude Code connection status
func (h *ClaudeCodeHandlers) handleStatus(w http.ResponseWriter, r *http.Request) {
	status := map[string]interface{}{
		"isConnected": true, // For now, always return true
		"service":     "claude-code",
		"timestamp":   time.Now().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// handleWebSocket will handle WebSocket connections (future implementation)
func (h *ClaudeCodeHandlers) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement WebSocket proxy to Claude Code service
	http.Error(w, "WebSocket support coming soon", http.StatusNotImplemented)
}