.PHONY: build clean test deploy-lambda package-lambda install deps help

# Binary name
BINARY_NAME=moonshot

# Build directory
BUILD_DIR=build

# Lambda deployment
LAMBDA_NAME=moonshot-dca-bot
LAMBDA_REGION=us-east-2
LAMBDA_RUNTIME=provided.al2
LAMBDA_HANDLER=bootstrap

# Go files
GO_FILES=$(shell find . -name "*.go" -type f)

# Default target
all: build

# Build the Lambda function
build: deps
	@echo "Building Moonshot DCA Bot for Lambda..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME) cmd/moonshot/main.go
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod tidy
	go mod download

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)
	go clean

# Run tests
test: deps
	@echo "Running tests..."
	go test -v ./...

# Package Lambda function
package-lambda: build
	@echo "Packaging Lambda function..."
	@cd $(BUILD_DIR) && zip -r $(BINARY_NAME).zip $(BINARY_NAME)
	@echo "Lambda package created: $(BUILD_DIR)/$(BINARY_NAME).zip"

# Deploy to AWS Lambda
deploy-lambda: package-lambda
	@echo "Deploying to AWS Lambda..."
	@if ! command -v aws &> /dev/null; then \
		echo "❌ AWS CLI not found. Please install it first."; \
		echo "   Visit: https://aws.amazon.com/cli/"; \
		exit 1; \
	fi
	@echo "Creating/updating Lambda function..."
	aws lambda create-function \
		--function-name $(LAMBDA_NAME) \
		--runtime $(LAMBDA_RUNTIME) \
		--handler $(LAMBDA_HANDLER) \
		--zip-file fileb://$(BUILD_DIR)/$(BINARY_NAME).zip \
		--role arn:aws:iam::$(shell aws sts get-caller-identity --query Account --output text):role/lambda-execution-role \
		--timeout 300 \
		--memory-size 512 \
		--region $(LAMBDA_REGION) \
		--environment Variables='{COINBASE_API_KEY="$(COINBASE_API_KEY)",COINBASE_API_SECRET="$(COINBASE_API_SECRET)",COINBASE_PASSPHRASE="$(COINBASE_PASSPHRASE)"}' \
		2>/dev/null || \
	aws lambda update-function-code \
		--function-name $(LAMBDA_NAME) \
		--zip-file fileb://$(BUILD_DIR)/$(BINARY_NAME).zip \
		--region $(LAMBDA_REGION)
	@echo "✅ Lambda function deployed successfully!"

# Test Lambda function locally
test-lambda: build
	@echo "Testing Lambda function locally..."
	@echo "You can test the function with:"
	@echo "  echo '{}' | ./$(BUILD_DIR)/$(BINARY_NAME)"

# Install to system (optional)
install: build
	@echo "Installing Moonshot to system..."
	sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	@echo "Installation complete. Run 'moonshot --help' for usage."

# Development mode with hot reload (requires air)
dev: deps
	@if ! command -v air &> /dev/null; then \
		echo "Installing air for hot reload..."; \
		go install github.com/cosmtrek/air@latest; \
	fi
	@echo "Starting development mode with hot reload..."
	air

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Lint code
lint: deps
	@echo "Linting code..."
	@if ! command -v golangci-lint &> /dev/null; then \
		echo "Installing golangci-lint..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	fi
	golangci-lint run

# Security check
security: deps
	@echo "Running security checks..."
	go vet ./...
	@if ! command -v gosec &> /dev/null; then \
		echo "Installing gosec..."; \
		go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest; \
	fi
	gosec ./...

# Generate documentation
docs:
	@echo "Generating documentation..."
	@if ! command -v godoc &> /dev/null; then \
		echo "Installing godoc..."; \
		go install golang.org/x/tools/cmd/godoc@latest; \
	fi
	@echo "Starting documentation server at http://localhost:6060"
	@echo "Press Ctrl+C to stop"
	godoc -http=:6060

# Check for updates
update: deps
	@echo "Checking for dependency updates..."
	go list -u -m all

# Show help
help:
	@echo "Moonshot DCA Bot - Available commands:"
	@echo ""
	@echo "  build          - Build the Lambda function"
	@echo "  clean          - Clean build artifacts"
	@echo "  test           - Run tests"
	@echo "  package-lambda - Package Lambda function for deployment"
	@echo "  deploy-lambda  - Deploy to AWS Lambda"
	@echo "  test-lambda    - Test Lambda function locally"
	@echo "  install        - Install to system"
	@echo "  dev            - Development mode with hot reload"
	@echo "  fmt            - Format code"
	@echo "  lint           - Lint code"
	@echo "  security       - Run security checks"
	@echo "  docs           - Generate documentation"
	@echo "  update         - Check for dependency updates"
	@echo "  help           - Show this help message"
	@echo ""
	@echo "Examples:"
	@echo "  make build           # Build the Lambda function"
	@echo "  make package-lambda  # Package for deployment"
	@echo "  make deploy-lambda   # Deploy to AWS Lambda"
	@echo ""
	@echo "Environment Variables:"
	@echo "  COINBASE_API_KEY     - Your Coinbase API key"
	@echo "  COINBASE_API_SECRET  - Your Coinbase API secret"
	@echo "  COINBASE_PASSPHRASE  - Your Coinbase passphrase"
	@echo "  LAMBDA_REGION        - AWS region (default: us-east-1)"
	@echo "  LAMBDA_NAME          - Lambda function name (default: moonshot-dca-bot)"
