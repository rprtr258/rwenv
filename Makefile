RWENV := go run main.go

.PHONY: _help
_help:
	@echo "TODO: implement"

.PHONY: run-dotenv
run-dotenv: # test run on .env file
	${RWENV} -vie .env ${RWENV}

.PHONY: run-rwenv-dotenv
run-rwenv-dotenv: # test run rwenv on rwenv on .env file
	${RWENV} -vi ${RWENV} -e .env -vi ${RWENV}

.PHONY: fmt
fmt: # format go sources
	go fmt
	gofumpt -l -w *.go

.PHONY: lint
lint: # run linter
	golangci-lint run ./...

.PHONY: todo
todo: # show list of all todos left in code
	@rg 'TODO' --glob '**/*.go' || echo 'All done!'

.PHONY: build
build: # build executable
	@go build -o rwenv main.go

.PHONY: bump
bump: # bump dependencies
	@go get -u ./...
	@go mod tidy
