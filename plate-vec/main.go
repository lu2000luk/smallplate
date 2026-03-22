package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
)

const Version = "0.1.0"

func main() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	if err := godotenv.Load(); err != nil {
		Warn("Error loading .env file (ignore if vars are set in the env)")
	}

	deps, err := buildDependencies()
	if err != nil {
		Error("Failed to build dependencies:", err)
		Log("Variables configuration guide:")
		Log("SERVICE_ID: The unique identifier for this service")
		Log("SERVICE_KEY: The secret key to connect to the manager")
		Log("MANAGER_URL: The URL of the manager service (e.g. manager.example.com), no protocol, no path")
		Log("CHROMA_URL: The internal URL of the Chroma API")
		return
	}
	defer func() {
		if err := deps.Close(); err != nil {
			Warn("Dependency cleanup error:", err)
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go deps.Manager.Run(ctx)

	server := newHTTPServer(deps)
	errCh := make(chan error, 1)
	go func() {
		Log("HTTP server listening on", deps.Config.HTTPAddr)
		if serveErr := server.ListenAndServe(); serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
			errCh <- serveErr
		}
	}()

	select {
	case <-interrupt:
		Log("Interrupt received, shutting down")
	case serveErr := <-errCh:
		Error("HTTP server failed:", serveErr)
	}

	cancel()
	shutdownHTTPServer(server, deps.Config.ShutdownTimeout)
	Log("Shutdown complete")
}
