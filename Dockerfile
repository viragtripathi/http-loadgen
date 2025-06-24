# Stage 1: build the statically linked binary for Linux
FROM golang:1.24-alpine as builder

RUN apk add --no-cache git

WORKDIR /app
COPY . .

RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o http-loadgen ./cmd/main.go

# Stage 2: use a minimal final image
FROM alpine:3.19

WORKDIR /app
COPY --from=builder /app/http-loadgen /app/http-loadgen
COPY config /app/config
COPY api /app/api

ENTRYPOINT ["/app/http-loadgen"]
