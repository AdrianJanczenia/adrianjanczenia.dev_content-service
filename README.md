# Content service

This Go service supplies Polish and English portfolio content over gRPC, handles CV token requests from RabbitMQ, and streams CV files over HTTP. Redis holds the short-lived CAPTCHA and token state shared with the CAPTCHA service.

## Interfaces

| Interface | Purpose |
| --- | --- |
| gRPC `ContentService.Handle` | Read localized content from `content/pl.json` or `content/en.json`. |
| RabbitMQ CV request queue | Validate the language, solved CAPTCHA state, and password; reply with a short-lived download token or an error slug. |
| `GET /download/cv?token=...&lang=...` | Redeem a token and stream the configured PDF. |

The gRPC contract is in `api/proto/v1/content.proto`. File paths, server ports, token lifetime, and RabbitMQ topology are configured in `config/<environment>/config.yml`. Keep the localized JSON files aligned when editing site content.

HTTP and RabbitMQ error replies carry distinct, machine-readable slugs for invalid input, unsupported language, CAPTCHA state, and CV failures. The front end translates these slugs from the paired Polish and English content files.

## CV authorization

The token process delegates request validation, constant-time password comparison, token generation, and issuing to focused tasks. The token is 32 characters generated with `crypto/rand`. A Redis Lua script checks the solved CAPTCHA, decrements attempts for a wrong password without extending the challenge TTL, and atomically records a successful issue. The original CAPTCHA state is removed after issue. A result mapping lets a successful retry for the same CAPTCHA ID and language receive the same still-valid token; a different language is rejected.

The download process opens and checks the selected regular file before redeeming the token; these steps have separate tasks. A second Lua script validates the token's language and deletes both token and result mapping atomically. A token can be redeemed only once, so an interrupted transfer cannot be resumed with that token. PDF responses carry `Cache-Control: no-store` and `Referrer-Policy: no-referrer`.

The RabbitMQ worker uses manual acknowledgements and a prefetch of one message per consumer. It publishes replies in confirm mode. If publishing a reply fails, it logs the failure and acknowledges the original request; a client can retry the same CAPTCHA ID while the result mapping is valid. Publisher confirms do not guarantee that a client received the reply.

Redis scripts live in `internal/service/redis/scripts/` and are embedded in the Go binary. They receive all accessed keys explicitly. The adapter uses a single Redis endpoint (`redis.NewClient`); the current multi-key layout is not designed for Redis Cluster hash-slot constraints.

## Configuration

The module requires Go 1.25.1 (see `go.mod`). `APP_ENV=production` selects `config/production/config.yml`; otherwise the local file is used. Supply `CV_PASSWORD`, `REDIS_URL`, and `RABBITMQ_URL` through environment variables; their YAML fields are intentionally empty. Startup validates required settings, checks Redis with a bounded ping, declares RabbitMQ topology, and starts the HTTP, gRPC, and RabbitMQ consumers.

CV files belong in `private/` and are selected by language through the configured paths. The directory is excluded from the Docker build context; Compose mounts it at runtime. Do not include real CVs or credentials in tests or documentation.

From the repository root, start Redis and RabbitMQ, provide the required configuration and CV files, and run `go run .`. The `Dockerfile` is available for container builds.

## Verification

```sh
go test -race ./...
```

Opt-in integration tests use real Redis and RabbitMQ to check token transitions, concurrent issuance, and request/reply delivery:

```sh
go test -race -tags=integration ./internal/service/redis ./internal/service/rabbitmq
```

Set `REDIS_TEST_URL` and, for the RabbitMQ test, `RABBIT_TEST_URL` to isolated test services. Tests whose variables are unset skip; enabled tests create temporary keys and broker topology.
