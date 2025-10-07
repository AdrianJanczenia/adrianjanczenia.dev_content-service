FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /app/content-service ./main.go

FROM gcr.io/distroless/base-debian11
WORKDIR /app
COPY --from=builder /app/config ./config
COPY --from=builder /app/content ./content
COPY --from=builder /app/content-service .
USER nonroot:nonroot
CMD ["/app/content-service"]