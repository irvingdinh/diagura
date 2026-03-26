.PHONY: install kill run lint

install:
	@go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.11.4
	@cd api && go mod download

kill:
	@-lsof -ti :48310 | xargs kill -9 2>/dev/null; true

run: kill
	@cd api && go run .

lint:
	@cd api && golangci-lint run --fix ./...
