package main

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"zenmind-kanban-server/internal/config"
	"zenmind-kanban-server/internal/kanban"
	"zenmind-kanban-server/internal/realtime"
	"zenmind-kanban-server/internal/store"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	cfg := config.Load()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	sqliteStore, err := store.Open(ctx, cfg.DatabasePath)
	if err != nil {
		logger.Error("failed to open store", "error", err)
		os.Exit(1)
	}
	defer sqliteStore.Close()

	service := kanban.NewService(sqliteStore)
	hub := realtime.NewHub(cfg, service, sqliteStore, logger)

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	})
	mux.HandleFunc("/api/snapshot", func(w http.ResponseWriter, r *http.Request) {
		if cfg.Token != "" && r.Header.Get("Authorization") != "Bearer "+cfg.Token {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		result, err := service.Snapshot(r.Context(), kanban.DefaultBoardID, r.URL.Query().Get("projectId"), hub.DesktopStatus())
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "message": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, result)
	})
	mux.HandleFunc("/api/issues", handleIssues(cfg, service))
	mux.HandleFunc("/ws", hub.ServeWS)

	server := &http.Server{
		Addr:              cfg.Addr,
		Handler:           withCORS(cfg, mux),
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		logger.Info("kanban server listening", "addr", cfg.Addr, "db", cfg.DatabasePath)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server failed", "error", err)
			stop()
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("server shutdown failed", "error", err)
	}
}

func withCORS(cfg config.Config, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if isAllowedOrigin(cfg, origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			if origin == "" {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			}
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
			w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func handleIssues(cfg config.Config, service *kanban.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if cfg.Token != "" && r.Header.Get("Authorization") != "Bearer "+cfg.Token {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		start := time.Now()
		result, err := service.ListProjectIssues(r.Context(), kanban.DefaultBoardID, r.URL.Query().Get("projectId"))
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "message": err.Error()})
			return
		}
		elapsed := time.Since(start)
		w.Header().Set("Server-Timing", "issues;dur="+strconv.FormatFloat(elapsed.Seconds()*1000, 'f', 2, 64))
		w.Header().Set("X-Issue-Count", strconv.Itoa(result.Count))
		writeJSON(w, http.StatusOK, result)
	}
}

func isAllowedOrigin(cfg config.Config, origin string) bool {
	for _, allowed := range cfg.AllowedOrigins {
		if allowed == "*" || allowed == origin {
			return true
		}
	}
	return false
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
