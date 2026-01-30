.PHONY: test test-race lint update-deps

# Run tests
test:
	@echo "Testing main simba module..."
	@go test -v ./...
	@echo "\nTesting websocket module..."
	@cd websocket && go test -v ./...
	@echo "\nTesting telemetry module..."
	@cd telemetry && go test -v ./...

# Run tests with race detector
test-race:
	@echo "Testing main simba module with race detector..."
	@go test -v -race ./...
	@echo "\nTesting websocket module with race detector..."
	@cd websocket && go test -v -race ./...
	@echo "\nTesting telemetry module with race detector..."
	@cd telemetry && go test -v -race ./...

# Run linters
lint:
	@echo "Running golangci-lint on main module..."
	-@golangci-lint run ./...
	@echo "\nRunning golangci-lint on websocket module..."
	-@cd websocket && golangci-lint run ./...
	@echo "\nRunning golangci-lint on telemetry module..."
	-@cd telemetry && golangci-lint run ./...
	@echo "\nRunning nilaway on main module..."
	@nilaway -test=false ./...
	@echo "\nRunning nilaway on websocket module..."
	@cd websocket && nilaway -test=false ./...
	@echo "\nRunning nilaway on telemetry module..."
	@cd telemetry && nilaway -test=false ./...

# Update dependencies
update-deps:
	@echo "Updating main module dependencies..."
	@go get -u ./...
	@go mod tidy
	@echo "\nUpdating websocket module dependencies..."
	@cd websocket && go get -u ./... && go mod tidy
	@echo "\nUpdating telemetry module dependencies..."
	@cd telemetry && go get -u ./... && go mod tidy
