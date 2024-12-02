# Run tests
test:
	@go test -v ./...

update-deps:
	@go get -u ./...

.PHONY: test, update-deps
