# Makefile for Video Translation Simulator

.PHONY: tidy start-server test

# Default configuration values
DELAY ?= 10
ERROR_RATE ?= 20


# Target to tidy up Go modules
tidy:
	go mod tidy

# Target to start the server with configurable delay and error rate
# Usage: make start-server DELAY=20 ERROR_RATE=25
server:
	go run cmd/server/main.go --delay $(DELAY) --error $(ERROR_RATE)

client:
	go run cmd/client/main.go
# Target to run tests for the client package
test:
	go test Video-Translation-Simulator/pkg/client -v

# Help target to display available commands
help:
	@echo "Available commands:"
	@echo "  make tidy                      Tidy up Go modules"
	@echo "  make start-server DELAY=20 ERROR_RATE=25  Start the server with specified delay and error rate"
	@echo "  make test                      Run tests for the client package"
