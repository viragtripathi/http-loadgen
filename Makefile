# Output binary name
APP_NAME = http-loadgen

# Default build
build:
	go build -o $(APP_NAME) cmd/main.go

build-linux:
	GOOS=linux GOARCH=amd64 go build -o $(APP_NAME)-linux cmd/main.go

build-mac:
	GOOS=darwin GOARCH=arm64 go build -o $(APP_NAME)-mac cmd/main.go

build-windows:
	GOOS=windows GOARCH=amd64 go build -o $(APP_NAME).exe cmd/main.go

build-all: build build-linux build-mac build-windows

clean:
	rm -f $(APP_NAME) $(APP_NAME)-linux $(APP_NAME)-mac $(APP_NAME).exe

# Tests
test:
	go test ./...

test-ci:
	go test -short ./...

test-integration:
	go test -v -tags=integration ./...

# Benchmark
bench:
	./scripts/run.sh --benchmark

# Docker build
docker-build:
	docker build -t $(APP_NAME):latest .

docker-buildx:
	docker buildx build --platform linux/amd64,linux/arm64 -t $(APP_NAME):multiarch .
