package server

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"
)

// Server encapsulates the HTTP server and its dependencies
type Server struct {
	server *http.Server
	mux    *http.ServeMux
	config *Config
}

func New(cfg *Config, mux *http.ServeMux) *Server {
	srv := &Server{
		mux:    mux,
		config: cfg,
		server: &http.Server{
			Addr:         cfg.Address(),
			Handler:      mux,
			ReadTimeout:  cfg.ReadTimeout,
			WriteTimeout: cfg.WriteTimeout,
		},
	}

	srv.registerRoutes()

	return srv
}

func (s *Server) Start() error {
	return s.server.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

// registerRoutes sets up all the routes for the server
func (s *Server) registerRoutes() {
	// health endpoint
	s.mux.HandleFunc("GET /health", s.handleHealth)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	response := HealthResponse{
		Status: "healthy",
		Time:   time.Now().Format(time.RFC3339),
	}

	WriteJSONResponse(w, http.StatusOK, response)
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status string `json:"status"`
	Time   string `json:"time"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// WriteJSONResponse is a helper function to write JSON responses
func WriteJSONResponse(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		slog.Error("error encoding response", "error", err)
	}
}
