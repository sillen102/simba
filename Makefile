.PHONY: test, lint, update-deps

# Run tests
test:
	@go test -v ./...

# Run linters
lint:
	@staticcheck ./...
	@nilaway ./...

# Update dependencies
update-deps:
	@go get -u ./...
