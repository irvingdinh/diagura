.PHONY: install kill run lint

install:
	@go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.11.4
	@cd api && go mod download

kill:
	@-lsof -ti :48310 | xargs kill -15 2>/dev/null; true
	@-lsof -ti :48305 | xargs kill -15 2>/dev/null; true

run: kill
	@cd api && LOG_FORMAT=text go run . & \
		cd admin && bun dev & \
		wait

lint:
	@cd admin && bun run lint:fix
	@cd api && golangci-lint run --fix ./...
