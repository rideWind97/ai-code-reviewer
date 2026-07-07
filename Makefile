.PHONY: test run-server run-worker

test:
	go test ./...

run-server:
	go run ./cmd/aicr-server

run-worker:
	go run ./cmd/aicr-worker
