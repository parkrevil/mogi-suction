.PHONY: help build-server build-client run-server run-client watch-server watch-client test clean deps

# 기본 타겟
help:
	@echo "Available commands:"
	@echo "  make deps                  - Install dependencies"
	@echo "  make build-server          - Build server"
	@echo "  make build-client          - Build client"
	@echo "  make build                 - Build all"
	@echo "  make run-server            - Run server"
	@echo "  make run-client            - Run client"
	@echo "  make watch-server          - Run server with hot reload"
	@echo "  make watch-client          - Run client with hot reload"
	@echo "  make test                  - Run tests"
	@echo "  make clean                 - Clean build artifacts"

# 의존성 설치
deps:
	@echo "Installing dependencies..."
	go mod tidy
	cd apps/server && go mod tidy
	cd apps/client && go mod tidy

# 서버 빌드
build-server:
	@echo "Building server..."
	cd apps/server && go build -o ../../bin/server .

# 클라이언트 빌드
build-client:
	@echo "Building client..."
	cd apps/client && go build -o ../../bin/client .

# 전체 빌드
build: build-server build-client
	@echo "All binaries built successfully!"

# 서버 실행
run-server:
	@echo "Starting server..."
	cd apps/server && go run main.go

# 클라이언트 실행
run-client:
	@echo "Starting client..."
	cd apps/client && go run main.go

# 서버 watch 모드 (hot reload)
watch-server:
	@echo "Starting server with hot reload..."
	cd apps/server && $(HOME)/go/bin/air

# 클라이언트 watch 모드 (sudo 권한으로 hot reload)
watch-client:
	@echo "Starting client with hot reload..."
	cd apps/client && sudo env PATH=$(HOME)/go/bin:$$PATH air

# 테스트 실행
test:
	@echo "Running tests..."
	go test ./...
	cd apps/server && go test ./...
	cd apps/client && go test ./...

# 정리
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -rf apps/server/tmp/
	rm -rf apps/client/tmp/
	go clean
	cd apps/server && go clean
	cd apps/client && go clean 