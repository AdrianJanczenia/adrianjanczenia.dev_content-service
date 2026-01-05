# Adrian Janczenia - Content Service

The **Content Service** serves as the primary data provider (Source of Truth) and secure document distributor within the portfolio microservices ecosystem. It handles localized content delivery, CV access authorization, and secure file streaming.

## Service Role

This service is the core "worker" and data manager of the system. Its primary responsibilities include:

- **Content Delivery (gRPC)**: Serving localized text and metadata for the frontend via high-performance gRPC procedures.
- **Asynchronous Processing (MQ)**: Acting as a RabbitMQ worker to handle CV download token requests using an RPC pattern.
- **Token Management (Redis)**: Generating, storing, and validating short-lived, one-time-use tokens for secure file access.
- **Secure File Serving**: Managing and streaming PDF files directly from the private file system with strict access control.

## Architecture and Resilience

The service is built with high availability and reliability in mind, following strict architectural standards.

### Layered Pattern: Handler -> Process -> Task
1. Handler: Acts as the entry point for gRPC calls or RabbitMQ messages. It unmarshals data and delegates work.
2. Process: Orchestrates the business logic, such as validating credentials before calling the token generator.
3. Task / Service: Performs atomic infrastructure operations like interacting with Redis or the file system.

### Advanced Features
- **Graceful Shutdown**: Implements a sophisticated shutdown mechanism that allows MQ consumers to finish processing "in-flight" messages before the service terminates.
- **Infrastructure Retry Strategy**: Features a robust startup loop that waits for Redis and RabbitMQ to become ready, ensuring the service doesn't fail during orchestrator cold starts.
- **Context Propagation**: Full context.Context integration across all layers for reliable timeout management and resource cleanup.

## Technical Specification

- Go: 1.23+ (utilizing the latest concurrency patterns).
- Redis: Used as a high-speed, volatile storage for session and download tokens.
- RabbitMQ: Used for decoupled, asynchronous communication and task distribution.
- gRPC: Provides the interface for synchronous content retrieval.
- Docker: Optimized multi-stage builds on Alpine Linux, ensuring a minimal security footprint.

## Environment Configuration

The service follows a "fail-fast" configuration approach, validating all necessary infrastructure links upon startup.

| Variable | Description |
|----------|-------------|
| APP_ENV | Runtime environment (local/production) |
| REDIS_URL | Connection string for the Redis instance |
| RABBITMQ_URL | Connection string for the RabbitMQ broker |
| CV_PASSWORD | The master password required to request a CV token |

## Development and Deployment

### Build Optimized Docker Image
docker build -t content-service .

### Execute Unit Tests
go test -v ./...

## Data Flow: CV Request
1. Gateway Service publishes a JSON request to the 'cv_requests' queue.
2. Content Service Consumer receives the message and triggers the Handler.
3. The Process validates the password and language.
4. The Task generates a UUID, saves it to Redis with a TTL, and returns it.
5. The Broker publishes the response back to the 'reply_to' queue specified in the message.

---
Adrian Janczenia