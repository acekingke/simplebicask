.PHONY: all build test clean

BINARY_NAME=simplebitcask
MAIN_PATH=./cmd/main.go

all: build

build:
	@echo "Building..."
	@go build -o $(BINARY_NAME) $(MAIN_PATH)

test:
	@echo "Running tests..."
	@go test -v ./...
bench:
	@echo "Running benchmarks..."
	@go test -benchmem -run=^$$ -bench=. github.com/acekingke/simplebitcask/bitcask
clean:
	@echo "Cleaning..."
	@rm -f $(BINARY_NAME)
	@go clean

run: build
	@echo "Running..."
	@./$(BINARY_NAME)

cover:
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@rm coverage.out

fmt:
	@echo "Formatting code..."
	@go fmt ./...

vet:
	@echo "Vetting code..."
	@go vet ./...