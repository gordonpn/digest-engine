// Package app wires adapters and manages the application lifecycle.
package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/gordonpn/digest-engine/adapters/httpserver"
	"github.com/gordonpn/digest-engine/internal/config"
)

func Run(ctx context.Context, cfg config.Config, logger *slog.Logger) error {
	listener, err := net.Listen("tcp", net.JoinHostPort("", strconv.Itoa(cfg.Port)))
	if err != nil {
		return errors.New("HTTP listener could not start")
	}
	server := &http.Server{
		Handler:           httpserver.NewHandler(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}
	logger.Info("HTTP server listening", "port", cfg.Port)
	return serve(ctx, server, listener, cfg.ShutdownTimeout)
}

// serve owns listener and drains requests using a fresh shutdown context, because
// the application context has already been cancelled when shutdown begins.
func serve(ctx context.Context, server *http.Server, listener net.Listener, timeout time.Duration) error {
	defer listener.Close()
	result := make(chan error, 1)
	go func() { result <- server.Serve(listener) }()
	select {
	case err := <-result:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return errors.New("HTTP server failed")
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		err := server.Shutdown(shutdownCtx)
		if err != nil {
			_ = server.Close()
			<-result
			return fmt.Errorf("HTTP shutdown failed: %w", err)
		}
		<-result
		return nil
	}
}
