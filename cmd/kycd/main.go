// Command kycd is the composition root for the KYC service: it wires
// infrastructure adapters to application use cases and starts the server.
// It must contain no business logic.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/fintech/kyc/internal/application/eligibility"
	"github.com/fintech/kyc/internal/infrastructure/memory"
	"github.com/fintech/kyc/internal/interfaces/httpapi"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	repo := memory.NewScreeningRepository()
	if err := memory.SeedDemo(repo, time.Now()); err != nil {
		logger.Error("seed demo data", slog.String("error", err.Error()))
		os.Exit(1)
	}
	svc := eligibility.New(repo, time.Now, logger)

	addr := os.Getenv("KYC_ADDR")
	if addr == "" {
		addr = "localhost:8080"
	}
	srv := httpapi.NewServer(addr, httpapi.NewHandler(svc))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	errc := make(chan error, 1)
	go func() {
		logger.Info("kycd listening", slog.String("addr", addr))
		errc <- srv.ListenAndServe()
	}()

	select {
	case err := <-errc:
		if !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server failed", slog.String("error", err.Error()))
			os.Exit(1)
		}
	case <-ctx.Done():
		logger.Info("shutting down")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			logger.Error("shutdown failed", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}
}
