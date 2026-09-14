.PHONY: test
test:
	go tool gotestsum --format dots-v2 ./...

.PHONY: lint
lint:
	golangci-lint run

.PHONY: fmt
fmt:
	golangci-lint fmt
