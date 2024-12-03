# Run tests
test:
	@go test -v ./...

lint:
	@staticcheck ./...

update-deps:
	@go get -u ./...

.PHONY: test, update-deps
