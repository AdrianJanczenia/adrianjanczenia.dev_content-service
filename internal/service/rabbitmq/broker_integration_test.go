//go:build integration

package rabbitmq

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	handlerGetCVToken "github.com/AdrianJanczenia/adrianjanczenia.dev_content-service/internal/handler/get_cv_token"
	processGetCVToken "github.com/AdrianJanczenia/adrianjanczenia.dev_content-service/internal/process/get_cv_token"
	taskGetCVToken "github.com/AdrianJanczenia/adrianjanczenia.dev_content-service/internal/process/get_cv_token/task"
	"github.com/AdrianJanczenia/adrianjanczenia.dev_content-service/internal/registry"
	serviceRedis "github.com/AdrianJanczenia/adrianjanczenia.dev_content-service/internal/service/redis"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/rabbitmq/amqp091-go"
)

func TestCVTokenRPCWithRabbitMQAndRedis(t *testing.T) {
	rabbitURL, redisURL := os.Getenv("RABBIT_TEST_URL"), os.Getenv("REDIS_TEST_URL")

	if rabbitURL == "" || redisURL == "" {
		t.Skip("RABBIT_TEST_URL and REDIS_TEST_URL are required")
	}

	ctx := context.Background()
	redisOptions, err := redis.ParseURL(redisURL)
	if err != nil {
		t.Fatal(err)
	}

	store := redis.NewClient(redisOptions)
	defer store.Close()

	id := uuid.NewString()
	captchaKey := "captcha:" + id
	resultKey := "cv:result:" + id
	if err := store.Set(ctx, captchaKey, `{"value":"ABC","triesLeft":3,"solved":true}`, 30*time.Second).Err(); err != nil {
		t.Fatal(err)
	}

	var issuedToken string
	defer func() {
		store.Del(ctx, captchaKey, resultKey, "cv:token:"+issuedToken)
	}()

	tokenStore, err := serviceRedis.NewClient(redisURL)
	if err != nil {
		t.Fatal(err)
	}
	defer tokenStore.Close()

	broker, err := NewBroker(rabbitURL)
	if err != nil {
		t.Fatal(err)
	}
	defer broker.Shutdown()

	exchangeName := "audit.cv." + id
	queueName := "audit.cv." + id

	const routingKey = "cv.request"

	topology := registry.RabbitMQTopologyConfig{
		Exchanges: []registry.ExchangeConfig{{Name: exchangeName, Type: "direct"}},
		Queues: map[string]registry.QueueConfig{
			"requests": {Name: queueName},
		},
		Bindings: []registry.BindingConfig{{Exchange: exchangeName, QueueKey: "requests", RoutingKey: routingKey}},
	}
	if err := broker.DeclareTopology(topology); err != nil {
		t.Fatal(err)
	}
	clientConnection, err := amqp091.Dial(rabbitURL)
	if err != nil {
		t.Fatal(err)
	}
	defer clientConnection.Close()

	channel, err := clientConnection.Channel()
	if err != nil {
		t.Fatal(err)
	}
	defer channel.Close()
	defer func() {
		channel.QueueDelete(queueName, false, false, false)
		channel.ExchangeDelete(exchangeName, false, false)
	}()

	replyQueue, err := channel.QueueDeclare("", false, true, true, false, nil)
	if err != nil {
		t.Fatal(err)
	}

	replies, err := channel.Consume(replyQueue.Name, "", true, true, false, false, nil)
	if err != nil {
		t.Fatal(err)
	}

	process := processGetCVToken.NewProcess(
		taskGetCVToken.NewValidateRequestTask(map[string]string{"pl": "unused"}),
		taskGetCVToken.NewComparePasswordTask("test-only-password"),
		taskGetCVToken.NewIssueTokenTask(tokenStore, taskGetCVToken.NewCreateTokenTask(), 30*time.Second),
	)
	broker.RegisterConsumer(queueName, 1, handlerGetCVToken.NewHandler(process).Handle)
	workerCtx, cancelWorker := context.WithCancel(ctx)
	workerDone := make(chan error, 1)
	go func() { workerDone <- broker.Start(workerCtx) }()
	defer func() {
		cancelWorker()
		select {
		case err := <-workerDone:
			if err != nil {
				t.Errorf("worker stopped with error: %v", err)
			}
		case <-time.After(5 * time.Second):
			t.Error("worker did not stop")
		}
	}()

	request, err := json.Marshal(map[string]string{"password": "test-only-password", "lang": "pl", "captchaId": id})
	if err != nil {
		t.Fatal(err)
	}

	for attempt := 0; attempt < 2; attempt++ {
		correlationID := uuid.NewString()
		if err := channel.PublishWithContext(ctx, exchangeName, routingKey, true, false, amqp091.Publishing{
			ContentType: "application/json", CorrelationId: correlationID, ReplyTo: replyQueue.Name, Body: request,
		}); err != nil {
			t.Fatal(err)
		}
		select {
		case response := <-replies:
			if response.CorrelationId != correlationID {
				t.Fatalf("reply correlation ID = %q, expected %q", response.CorrelationId, correlationID)
			}
			var payload struct {
				Token string `json:"token"`
				Error string `json:"error"`
			}
			if err := json.Unmarshal(response.Body, &payload); err != nil || payload.Token == "" || payload.Error != "" {
				t.Fatalf("invalid reply: %+v err=%v", payload, err)
			}
			if attempt == 0 {
				issuedToken = payload.Token
			} else if payload.Token != issuedToken {
				t.Fatalf("retry issued another token: %q", payload.Token)
			}
		case <-time.After(5 * time.Second):
			t.Fatal("CV token reply timed out")
		}
	}
}
