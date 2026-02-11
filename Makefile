# Detect OS
ifeq ($(OS),Windows_NT)
    # Windows-specific settings
    RM = del /Q /F
    RM_DIR = rmdir /S /Q
    SLASH = \\
    FIX_PATH = $(subst /,\,$(1))
    EXE = .exe
else
    # Unix/Linux/macOS settings
    RM = rm -f
    RM_DIR = rm -rf
    SLASH = /
    FIX_PATH = $(1)
    EXE =
endif

.PHONY: all clean proto build run-api run-core

PROTO_DIR := internal/proto
OUT_DIR := .

clean:
	@echo Cleaning...
	$(RM) $(call FIX_PATH,$(PROTO_DIR)/*.pb.go) 2>NUL || exit 0
	@if exist bin $(RM_DIR) bin 2>NUL || exit 0

proto:
	protoc --go_out=$(OUT_DIR) --go_opt=paths=source_relative \
	--go-grpc_out=$(OUT_DIR) --go-grpc_opt=paths=source_relative \
	$(PROTO_DIR)/*.proto

build: proto
	go build -o bin/api$(EXE) cmd/api/main.go
	go build -o bin/core$(EXE) cmd/core/main.go

run-api:
	go run cmd/api/main.go

run-core:
	go run cmd/core/main.go

test:
	@echo "Running tests..."
	go test -v -coverpkg=./... -coverprofile=coverage.out ./...
	@echo "Filtering generated files..."
	cat coverage.out | grep -v ".pb.go" | grep -v "mock_" > coverage_clean.out
	@echo "Generating HTML report..."
	go tool cover -html=coverage_clean.out

all: clean build test