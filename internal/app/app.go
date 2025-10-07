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
	handlerGetCvLink "github.com/AdrianJanczenia/adrianjanczenia.dev_content-service/internal/handler/get_cv_link"
	processGetContent "github.com/AdrianJanczenia/adrianjanczenia.dev_content-service/internal/process/get_content"
	processGetCvLink "github.com/AdrianJanczenia/adrianjanczenia.dev_content-service/internal/process/get_cv_link"
	taskGetCvLink "github.com/AdrianJanczenia/adrianjanczenia.dev_content-service/internal/process/get_cv_link/task"
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
	maxRetries := 15
	retryDelay := 2 * time.Second
	var err error

	var redisClient *serviceRedis.Client
	for i := 0; i < maxRetries; i++ {
		redisClient = serviceRedis.NewClient(cfg.Redis.Addr)
		if err = redisClient.Ping(context.Background()); err == nil {
			log.Println("INFO: successfully connected to Redis")
			break
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

	contentProcess, err := processGetContent.NewProcess(cfg.Content.Path)
	if err != nil {
		return nil, err
	}

	validatePasswordTask := taskGetCvLink.NewValidatePasswordTask(cfg.Cv.Password)
	createTokenTask := taskGetCvLink.NewCreateTokenTask(redisClient, cfg.Cv.TokenTTL)
	cvProcess := processGetCvLink.NewProcess(validatePasswordTask, createTokenTask, redisClient, cfg.Cv.FilePath)

	grpcHandler := handlerGetContent.NewHandler(contentProcess)
	httpHandler := handlerDowloadCv.NewHandler(cvProcess)
	getCvLinkConsumer := handlerGetCvLink.NewConsumer(cvProcess, rabbitBroker)
	rabbitBroker.RegisterConsumer(cfg.RabbitMQ.CVRequestQueue, getCvLinkConsumer.Handle)

	grpcServer := grpc.NewServer()
	contentv1.RegisterContentServiceServer(grpcServer, grpcHandler)

	mux := http.NewServeMux()
	mux.HandleFunc("/download/cv", httpHandler.Handle)

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
