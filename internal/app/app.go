package app

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"google.golang.org/grpc"

	"github.com/AdrianJanczenia/adrianjanczenia.dev_content-service/api/proto/v1"
	handlerDownloadCV "github.com/AdrianJanczenia/adrianjanczenia.dev_content-service/internal/handler/download_cv"
	handlerGetContent "github.com/AdrianJanczenia/adrianjanczenia.dev_content-service/internal/handler/get_content"
	handlerGetCvToken "github.com/AdrianJanczenia/adrianjanczenia.dev_content-service/internal/handler/get_cv_token"
	processDownloadCv "github.com/AdrianJanczenia/adrianjanczenia.dev_content-service/internal/process/download_cv"
	taskDownloadCv "github.com/AdrianJanczenia/adrianjanczenia.dev_content-service/internal/process/download_cv/task"
	processGetContent "github.com/AdrianJanczenia/adrianjanczenia.dev_content-service/internal/process/get_content"
	processGetCvToken "github.com/AdrianJanczenia/adrianjanczenia.dev_content-service/internal/process/get_cv_token"
	taskGetCvToken "github.com/AdrianJanczenia/adrianjanczenia.dev_content-service/internal/process/get_cv_token/task"
	"github.com/AdrianJanczenia/adrianjanczenia.dev_content-service/internal/registry"
	serviceRabbitmq "github.com/AdrianJanczenia/adrianjanczenia.dev_content-service/internal/service/rabbitmq"
	serviceRedis "github.com/AdrianJanczenia/adrianjanczenia.dev_content-service/internal/service/redis"
)

type App struct {
	grpcServer   *grpc.Server
	httpServer   *http.Server
	rabbitBroker *serviceRabbitmq.Broker
	redisClient  *serviceRedis.Client
	grpcPort     string
}

func Build(cfg *registry.Config) (*App, error) {
	if cfg == nil || cfg.Infrastructure.Retry.MaxAttempts < 1 || cfg.Infrastructure.Retry.DelaySeconds < 0 || cfg.Cv.TokenTTL <= 0 || cfg.Cv.Password == "" || len(cfg.Cv.Files) == 0 || cfg.Server.HTTPPort == "" || cfg.Server.GRPCPort == "" || cfg.Redis.URL == "" || cfg.RabbitMQ.URL == "" || cfg.RabbitMQ.Topology.Queues["cv_requests"].Name == "" {
		return nil, fmt.Errorf("invalid content-service configuration")
	}
	maxRetries := cfg.Infrastructure.Retry.MaxAttempts
	retryDelay := cfg.Infrastructure.Retry.DelaySeconds
	var err error

	var redisClient *serviceRedis.Client
	for i := range maxRetries {
		redisClient, err = serviceRedis.NewClient(cfg.Redis.URL)
		if err == nil {
			pingCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			err = redisClient.Ping(pingCtx)
			cancel()
			if err == nil {
				log.Println("INFO: successfully connected to Redis")
				break
			}
			_ = redisClient.Close()
		}
		if i+1 < maxRetries {
			log.Printf("INFO: could not connect to Redis, retrying in %v... (%d/%d)", retryDelay, i+1, maxRetries)
			time.Sleep(retryDelay)
		}
	}
	if err != nil {
		return nil, err
	}
	cleanupRedis := true
	defer func() {
		if cleanupRedis {
			_ = redisClient.Close()
		}
	}()

	var rabbitBroker *serviceRabbitmq.Broker
	for i := range maxRetries {
		if rabbitBroker, err = serviceRabbitmq.NewBroker(cfg.RabbitMQ.URL); err == nil {
			log.Println("INFO: successfully connected to RabbitMQ")
			break
		}
		if i+1 < maxRetries {
			log.Printf("INFO: could not connect to RabbitMQ, retrying in %v... (%d/%d)", retryDelay, i+1, maxRetries)
			time.Sleep(retryDelay)
		}
	}
	if err != nil {
		return nil, err
	}
	cleanupRabbit := true
	defer func() {
		if cleanupRabbit {
			_ = rabbitBroker.Shutdown()
		}
	}()

	if err := rabbitBroker.DeclareTopology(cfg.RabbitMQ.Topology); err != nil {
		return nil, err
	}

	getContentProcess, err := processGetContent.NewProcess(cfg.Content.Files, cfg.Content.DefaultLang)
	if err != nil {
		return nil, err
	}

	getCvTokenProcess := processGetCvToken.NewProcess(
		taskGetCvToken.NewValidateRequestTask(cfg.Cv.Files),
		taskGetCvToken.NewComparePasswordTask(cfg.Cv.Password),
		taskGetCvToken.NewIssueTokenTask(redisClient, taskGetCvToken.NewCreateTokenTask(), cfg.Cv.TokenTTL),
	)
	downloadCvProcess := processDownloadCv.NewProcess(
		taskDownloadCv.NewOpenFileTask(cfg.Cv.Files),
		taskDownloadCv.NewRedeemTokenTask(redisClient),
	)

	getContentHandler := handlerGetContent.NewHandler(getContentProcess)
	downloadCvHandler := handlerDownloadCV.NewHandler(downloadCvProcess)
	getCvTokenHandler := handlerGetCvToken.NewHandler(getCvTokenProcess)

	consumerCount := cfg.RabbitMQ.Consumers.DefaultCount
	if consumerCount <= 0 {
		consumerCount = 1
	}

	rabbitBroker.RegisterConsumer(cfg.RabbitMQ.Topology.Queues["cv_requests"].Name, consumerCount, getCvTokenHandler.Handle)

	grpcServer := grpc.NewServer()
	contentv1.RegisterContentServiceServer(grpcServer, getContentHandler)

	mux := http.NewServeMux()
	mux.HandleFunc("/download/cv", downloadCvHandler.Handle)

	httpServer := &http.Server{
		Addr: ":" + cfg.Server.HTTPPort,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, 16<<10)
			r.URL.Path = strings.TrimSuffix(r.URL.Path, "/")
			mux.ServeHTTP(w, r)
		}),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    32 << 10,
	}
	cleanupRedis, cleanupRabbit = false, false

	return &App{
		grpcServer:   grpcServer,
		httpServer:   httpServer,
		rabbitBroker: rabbitBroker,
		redisClient:  redisClient,
		grpcPort:     cfg.Server.GRPCPort,
	}, nil
}

func (a *App) RunGRPC() error {
	lis, err := net.Listen("tcp", ":"+a.grpcPort)
	if err != nil {
		return err
	}
	log.Printf("INFO: gRPC server listening on %s", lis.Addr().String())
	return a.grpcServer.Serve(lis)
}

func (a *App) RunHTTP() error {
	log.Printf("INFO: HTTP server listening on %s", a.httpServer.Addr)
	return a.httpServer.ListenAndServe()
}

func (a *App) RunRabbitMQConsumers(ctx context.Context) error {
	log.Println("INFO: starting RabbitMQ consumers")
	return a.rabbitBroker.Start(ctx)
}

func (a *App) Shutdown(ctx context.Context) {
	log.Println("INFO: shutting down servers...")
	stopped := make(chan struct{})
	go func() {
		a.grpcServer.GracefulStop()
		close(stopped)
	}()
	select {
	case <-stopped:
	case <-ctx.Done():
		a.grpcServer.Stop()
	}
	_ = a.httpServer.Shutdown(ctx)
	_ = a.rabbitBroker.Shutdown()
	_ = a.redisClient.Close()
}
