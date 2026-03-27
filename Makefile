.PHONY: install kill kill-ui run run-ui lint lint-ui

install:
	@go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.11.4
	@cd api && go mod download

kill:
	@-lsof -ti :48310 | xargs kill -9 2>/dev/null; true

kill-ui:
	@-lsof -ti :48305 | xargs kill -9 2>/dev/null; true

run: kill
	@cd api && go run .

run-ui: kill-ui
	@cd admin && bun dev

lint:
	@cd api && golangci-lint run --fix ./...

lint-ui:
	@cd admin && bun run lint:fix
