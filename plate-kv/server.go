package main

import (
	"context"
	"net/http"
	"time"

	"plain/kv/internal/plate"
	"plain/kv/routes"
)

func newHTTPServer(deps *plate.Dependencies) *http.Server {
	mux := http.NewServeMux()
	routes.Register(mux, deps)
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		plate.WriteOK(w, http.StatusOK, map[string]any{"status": "ok", "service": plate.ServiceType})
	})

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		plate.WriteOK(w, http.StatusOK, map[string]any{"status": "ok", "service": plate.ServiceType})
	})

	return &http.Server{
		Addr:              deps.Config.HTTPAddr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
}

func shutdownHTTPServer(server *http.Server, timeout time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		plate.Warn("HTTP server shutdown error:", err)
	}
}
