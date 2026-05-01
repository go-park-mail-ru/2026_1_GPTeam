FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN mkdir -p /out && go build -o /out/app . && go build -o /out/auth-service ./cmd/auth

FROM alpine:latest
WORKDIR /app
COPY --from=builder /out/app ./app
COPY --from=builder /out/auth-service ./auth-service
RUN chmod +x app auth-service
COPY .env .
CMD ["./app"]
