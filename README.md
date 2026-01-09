# Adrian Janczenia - Content Service

The **Content Service** serves as the primary data provider (Source of Truth) and secure document distributor within the portfolio microservices ecosystem. It handles localized content delivery, CV access authorization, and secure file streaming.

## Service Role

This service is the core data manager of the system. Its primary responsibilities include:

- **Content Delivery (gRPC)**: Serving localized text and metadata for the frontend.
- **Asynchronous Processing (MQ)**: Acting as a RabbitMQ worker to handle CV download token requests using an RPC pattern.
- **Token Management (Redis)**: Generating and validating high-entropy, short-lived tokens for secure file access.
- **Captcha Verification**: Checking the Captcha solution state in Redis before issuing CV tokens.

## Architecture and Security

The service is built with a "security-first" approach regarding document issuance.

### Multi-step CV Token Issuance Flow:
1. **Captcha Verification**: Checks Redis to ensure the Captcha session for the given ID is marked as solved.
2. **Password Validation**: Verifies the access password.
3. **Session Invalidation**: Immediately deletes the Captcha data from Redis after successful verification (one-time use session).
4. **Token Generation**: Creates a 32-character alphanumeric token (a-z, A-Z, 0-9) stored in Redis with a specific TTL.

### Layered Pattern: Handler -> Process -> Task
1. Handler: Entry point for gRPC calls or RabbitMQ messages.
2. Process: Orchestrates the business logic (Captcha Verify -> Password -> Session Delete -> Create Token).
3. Task / Service: Performs atomic infrastructure operations on Redis or the file system.

## Technical Specification

- Go: 1.23+
- Redis: Used for volatile storage of Captcha states and download tokens.
- RabbitMQ: Asynchronous communication for CV requests.
- Token Format: 32-character random alphanumeric string.

## Environment Configuration

| Variable | Description |
|----------|-------------|
| REDIS_URL | Connection string for the Redis instance |
| RABBITMQ_URL | Connection string for the RabbitMQ broker |
| CV_FILE_PATH | Absolute path to the CV PDF files in the container |

## Development and Deployment

### Build Optimized Docker Image
docker build -t content-service .

### Execute Unit Tests
go test -v ./...

---
Adrian Janczenia