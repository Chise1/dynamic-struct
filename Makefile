.PHONY: all serverless deps docker docker-cgo clean docs test test-race test-integration fmt lint install deploy-docs build

gotool:
	go fmt ./...
	go vet ./...


all: clean gotool $(TARGET) $(CMD)

deps:
	@go mod tidy

test:
	@go test -coverprofile=coverage.out ./...
html:
	@go tool cover -html=coverage.out