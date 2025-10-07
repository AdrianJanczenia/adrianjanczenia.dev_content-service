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
	registry.Cfg = cfg

	application, err := app.Build(registry.Cfg)
	if err != nil {
		log.Fatalf("FATAL: could not build application: %v", err)
	}

	go func() {
		if err := application.RunGRPC(); err != nil {
			log.Printf("ERROR: gRPC server failed: %v", err)
		}
	}()

	go func() {
		if err := application.RunHTTP(); !errors.Is(err, http.ErrServerClosed) {
			log.Printf("ERROR: HTTP server failed: %v", err)
		}
	}()

	go func() {
		if err := application.RunRabbitMQConsumers(); err != nil {
			log.Printf("ERROR: RabbitMQ consumer group failed: %v", err)
		}
	}()

	shutdownChannel := make(chan os.Signal, 1)
	signal.Notify(shutdownChannel, syscall.SIGINT, syscall.SIGTERM)
	sig := <-shutdownChannel
	log.Printf("INFO: received signal %s. Shutting down...", sig)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	application.Shutdown(ctx)
	log.Println("INFO: application shutdown complete.")
}
