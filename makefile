.PHONY: test tidy vet lint
GOFLAGS=-tags=unit

tidy:
	go mod tidy

vet:
	go vet ./...

lint:
	golangci-lint run

test:
	go test -v $(GOFLAGS) ./...

ci: tidy vet lint test
