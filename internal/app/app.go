package app

import (
	"context"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"google.golang.org/grpc"

	"github.com/AdrianJanczenia/adrianjanczenia.dev_content-service/api/proto/v1"
	handlerDowloadCv "github.com/AdrianJanczenia/adrianjanczenia.dev_content-service/internal/handler/download_cv"
	handlerGetContent "github.com/AdrianJanczenia/adrianjanczenia.dev_content-service/internal/handler/get_content"
	handlerGetCvToken "github.com/AdrianJanczenia/adrianjanczenia.dev_content-service/internal/handler/get_cv_token"
	processDownloadCv "github.com/AdrianJanczenia/adrianjanczenia.dev_content-service/internal/process/download_cv"
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
}

func Build(cfg *registry.Config) (*App, error) {
	maxRetries := 20
	retryDelay := 2 * time.Second
	var err error

	var redisClient *serviceRedis.Client
	for i := 0; i < maxRetries; i++ {
		redisClient, err = serviceRedis.NewClient(cfg.Redis.URL)
		if err == nil {
			if err = redisClient.Ping(context.Background()); err == nil {
				log.Println("INFO: successfully connected to Redis")
				break
			}
		}
		log.Printf("INFO: could not connect to Redis, retrying in %v... (%d/%d)", retryDelay, i+1, maxRetries)
		time.Sleep(retryDelay)
	}
	if err != nil {
		return nil, err
	}

	var rabbitBroker *serviceRabbitmq.Broker
	for i := 0; i < maxRetries; i++ {
		if rabbitBroker, err = serviceRabbitmq.NewBroker(cfg.RabbitMQ.URL); err == nil {
			log.Println("INFO: successfully connected to RabbitMQ")
			break
		}
		log.Printf("INFO: could not connect to RabbitMQ, retrying in %v... (%d/%d)", retryDelay, i+1, maxRetries)
		time.Sleep(retryDelay)
	}
	if err != nil {
		return nil, err
	}

	if err := rabbitBroker.DeclareTopology(cfg.RabbitMQ.Topology); err != nil {
		return nil, err
	}

	getContentProcess, err := processGetContent.NewProcess(cfg.Content.Files, cfg.Content.DefaultLang)
	if err != nil {
		return nil, err
	}

	validatePasswordTask := taskGetCvToken.NewValidatePasswordTask(cfg.Cv.Password)
	createTokenTask := taskGetCvToken.NewCreateTokenTask(redisClient, cfg.Cv.TokenTTL)
	getCvTokenProcess := processGetCvToken.NewProcess(validatePasswordTask, createTokenTask, cfg.Cv.Files)

	downloadCvProcess := processDownloadCv.NewProcess(redisClient, cfg.Cv.Files)

	// handlers
	getContentHandler := handlerGetContent.NewHandler(getContentProcess)
	downloadCvHandler := handlerDowloadCv.NewHandler(downloadCvProcess)
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
			r.URL.Path = strings.TrimSuffix(r.URL.Path, "/")
			mux.ServeHTTP(w, r)
		}),
	}

	return &App{
		grpcServer:   grpcServer,
		httpServer:   httpServer,
		rabbitBroker: rabbitBroker,
	}, nil
}

func (a *App) RunGRPC() error {
	lis, err := net.Listen("tcp", ":"+registry.Cfg.Server.GRPCPort)
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

func (a *App) RunRabbitMQConsumers() error {
	log.Println("INFO: starting RabbitMQ consumers")
	return a.rabbitBroker.Start()
}

func (a *App) Shutdown(ctx context.Context) {
	log.Println("INFO: shutting down servers...")
	a.grpcServer.GracefulStop()
	_ = a.httpServer.Shutdown(ctx)
	_ = a.rabbitBroker.Shutdown()
}
