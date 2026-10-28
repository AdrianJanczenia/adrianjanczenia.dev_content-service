package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/AdrianJanczenia/adrianjanczenia.dev_content-service/internal/app"
	"github.com/AdrianJanczenia/adrianjanczenia.dev_content-service/internal/registry"
)

func main() {
	cfg, err := registry.LoadConfig()
	if err != nil {
		log.Fatalf("FATAL: could not load configuration: %v", err)
	}
	application, err := app.Build(cfg)
	if err != nil {
		log.Fatalf("FATAL: could not build application: %v", err)
	}

	serverErrors := make(chan error, 3)
	consumerCtx, cancelConsumers := context.WithCancel(context.Background())
	go func() {
		if err := application.RunGRPC(); err != nil {
			serverErrors <- err
		}
	}()

	go func() {
		if err := application.RunHTTP(); !errors.Is(err, http.ErrServerClosed) {
			serverErrors <- err
		}
	}()

	go func() {
		if err := application.RunRabbitMQConsumers(consumerCtx); err != nil {
			serverErrors <- err
		}
	}()

	shutdownChannel := make(chan os.Signal, 1)
	signal.Notify(shutdownChannel, syscall.SIGINT, syscall.SIGTERM)
	failed := false
	select {
	case sig := <-shutdownChannel:
		log.Printf("INFO: received signal %s. Shutting down...", sig)
	case err := <-serverErrors:
		log.Printf("ERROR: service failed: %v", err)
		failed = true
	}
	cancelConsumers()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	application.Shutdown(ctx)
	log.Println("INFO: application shutdown complete.")
	if failed {
		os.Exit(1)
	}
}
