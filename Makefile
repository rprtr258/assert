.PHONY: test
test:
	gotestsum --format dots-v2 ./...
