package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"healthcheckbot/internal/checks"
	"healthcheckbot/internal/config"
	"healthcheckbot/internal/httpapi"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	store, err := checks.NewStore(cfg.DataFile)
	if err != nil {
		log.Fatalf("failed to initialize store: %v", err)
	}

	runner := checks.NewRunner(store, cfg.RequestUA, cfg.AlertWebhookURL, cfg.HTTPTimeoutSec, cfg.WebhookTimeoutSec)
	handler := httpapi.NewHandler(store, runner, cfg.APIKey)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go runner.Start(ctx)

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           handler.Routes(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("server listening on :%s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("server stopped with error: %v", err)
			stop()
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
		os.Exit(1)
	}
}
