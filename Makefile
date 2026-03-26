.PHONY: kill run

kill:
	@-lsof -ti :48310 | xargs kill -9 2>/dev/null; true

run: kill
	@cd api && go run .
