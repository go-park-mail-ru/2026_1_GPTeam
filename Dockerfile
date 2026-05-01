FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN mkdir -p /out && go build -o /out/app . && go build -o /out/auth-service ./cmd/auth && go build -o /out/ai-service ./cmd/ai

FROM alpine:latest
WORKDIR /app
COPY --from=builder /out/app ./app
COPY --from=builder /out/auth-service ./auth-service
COPY --from=builder /out/ai-service ./ai-service
RUN chmod +x app auth-service ai-service
COPY .env .
CMD ["./app"]
