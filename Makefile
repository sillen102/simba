.PHONY: test test-race lint update-deps

# Run tests
test:
	@go test -v ./...

# Run tests with race detector
test-race:
	@go test -v -race ./...

# Run linters
lint:
	@golangci-lint run ./...

# Update dependencies
update-deps:
	@go get -u ./...
