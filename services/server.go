package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/albertoboccolini/dsw/models"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type ServerHandler struct {
	configuration *Configuration
	Server        *Server
}

func NewServerHandler(configuration *Configuration, port int) *ServerHandler {
	router := chi.NewRouter()

	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.RequestID)

	server := &Server{
		router: router,
		httpServer: &http.Server{
			Addr:         fmt.Sprintf("0.0.0.0:%d", port),
			Handler:      router,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
		executor: NewExecutor(),
	}

	serverHandler := &ServerHandler{
		configuration: configuration,
		Server:        server,
	}

	server.router.Get("/actions", serverHandler.handleListActions)
	server.router.Post("/execute/{actionName}", serverHandler.handleExecuteAction)

	return serverHandler
}

type Server struct {
	router     chi.Router
	httpServer *http.Server
	executor   *Executor
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type ActionsResponse struct {
	Actions map[string]models.Action `json:"actions"`
}

func (serverHandler *ServerHandler) handleListActions(responseWriter http.ResponseWriter, request *http.Request) {
	response := ActionsResponse{
		Actions: serverHandler.configuration.Actions,
	}

	responseWriter.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(responseWriter).Encode(response); err != nil {
		slog.Error("failed to encode response", "error", err)
		serverHandler.Server.respondError(responseWriter, "internal server error", http.StatusInternalServerError)
		return
	}
}

func (serverHandler *ServerHandler) handleExecuteAction(responseWriter http.ResponseWriter, request *http.Request) {
	actionName := chi.URLParam(request, "actionName")

	action, exists := serverHandler.configuration.GetAction(actionName)
	if !exists {
		serverHandler.Server.respondError(responseWriter, fmt.Sprintf("action not found: %s", actionName), http.StatusNotFound)
		return
	}

	slog.Info("executing action", "name", actionName, "command", action.Command)

	result := serverHandler.Server.executor.Execute(action)

	slog.Info("action completed",
		"name", actionName,
		"success", result.Success,
		"duration_ms", result.DurationMs)

	responseWriter.Header().Set("Content-Type", "application/json")

	statusCode := http.StatusOK
	if !result.Success {
		statusCode = http.StatusInternalServerError
	}
	responseWriter.WriteHeader(statusCode)

	if err := json.NewEncoder(responseWriter).Encode(result); err != nil {
		slog.Error("failed to encode result", "error", err)
	}
}

func (server *Server) respondError(responseWriter http.ResponseWriter, message string, statusCode int) {
	responseWriter.Header().Set("Content-Type", "application/json")
	responseWriter.WriteHeader(statusCode)
	json.NewEncoder(responseWriter).Encode(ErrorResponse{Error: message})
}

func (server *Server) Start() error {
	go func() {
		slog.Info("starting server", "addr", server.httpServer.Addr)

		if err := server.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	server.waitForShutdown()
	return nil
}

func (server *Server) waitForShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	<-quit
	slog.Info("shutting down server")

	context, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.httpServer.Shutdown(context); err != nil {
		slog.Error("server shutdown error", "error", err)
	}

	slog.Info("server stopped")
}
