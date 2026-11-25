package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/rishabhkum-Dev-1799/go-payments/internal/platform/postgres"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	err := godotenv.Load()
	if err != nil {
		logger.Error("Error loading .env file")
		os.Exit(1)
	}
	dsn := os.Getenv("WALLET_DATABASE_URL")
	if dsn == "" {
		logger.Error("Unable to start the wallet service as WALLET_DATABASE_URL was not found in the env file")
		os.Exit(1)
	}
	pool, err := postgres.NewPool(context.Background(), dsn)
	if err != nil {
		logger.Error("connection to postgres failed", "err", err)
		os.Exit(1)
	}
	logger.Info("connection to wallet db is successfully", "max Connections", pool.Config().MaxConns)
	// Intializing the mux http handler
	mux := http.NewServeMux()
	// Using the http server mux as the handler function ...r
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		logger.Info("heath check", "remote", r.RemoteAddr)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"Ok"}` + "\n"))
	})
	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		w.Header().Set("Content-Type", "application/json")
		if err := pool.Ping(ctx); err != nil {
			logger.Error("initialize:unable to ping the database", "err", err)
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"status":"unavailable"}` + "\n"))
		}
		w.Write([]byte(`{"status":"ready"}` + "\n"))
	})

	// Now creating the server for the server handler ...r
	srv := &http.Server{
		Addr:              ":8081",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	// Now here writing the code for the graceful shutdown for the application and handling ...r
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		logger.Info("wallet service listening", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server startup failed", "err", err)
			os.Exit(1)
		}
	}()
	// Till here the program will be frezzed till we get the done signal which can be either from program termination or the interupption
	<-ctx.Done()
	logger.Info("shutting down gracefully")
	// Creating the shutdown ctx With Timeout of the 10 second that within the 10 second the server should be stopped else the cancel will be called ...r
	shutDownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutDownCtx); err != nil {
		logger.Error("shutdown failed", "err", err)
	}
	logger.Info("Server Stopped")
}
