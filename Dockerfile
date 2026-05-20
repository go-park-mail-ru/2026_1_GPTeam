FROM golang:1.26-alpine AS builder

WORKDIR /app

RUN apk add --no-cache protobuf make
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.0
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.3.0
ENV PATH="$PATH:$(go env GOPATH)/bin"
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN make proto
RUN mkdir -p /out && go build -o /out/app . && go build -o /out/auth-service ./cmd/auth && go build -o /out/ai-service ./cmd/ai && go build -o /out/fileserver ./cmd/fileserver

FROM alpine:latest
RUN apk add --no-cache curl
WORKDIR /app
COPY --from=builder /out/app ./app
COPY --from=builder /out/auth-service ./auth-service
COPY --from=builder /out/ai-service ./ai-service
COPY --from=builder /out/fileserver ./fileserver
RUN chmod +x app auth-service ai-service fileserver
CMD ["./app"]
