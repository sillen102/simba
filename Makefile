.PHONY: test test-race lint update-deps

# Run tests
test:
	@go test -v ./...

# Run tests with race detector
test-race:
	@go test -v -race ./...

# Run linters
lint:
	@echo "Running golangci-lint"
	-@golangci-lint run ./...
	@echo "Running nilaway"
	@nilaway -test=false ./...

# Update dependencies
update-deps:
	@go get -u ./...
