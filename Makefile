help: ## Show help messages.
	@grep -E '^[0-9a-zA-Z_-]+:(.*?## .*)?$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

run="."
dir="./..."
short="-short"
flags=""
timeout=40s

.PHONY: unittest
unittest: ## Run unit tests in watch mode. You can set: [run, timeout, short, dir, flags]. Example: make unittest flags="-race".
	@echo "running tests on $(run). waiting for changes..."
	@-zsh -c "go test -trimpath --timeout=$(timeout) $(short) $(dir) -run $(run) $(flags); repeat 100 printf '#'; echo"
	@reflex -d none -r "(\.go$$)|(go.mod)" -- zsh -c "go test -trimpath --timeout=$(timeout) $(short) $(dir) -run $(run) $(flags); repeat 100 printf '#'"

.PHONY: dependencies
dependencies: ## Install dependencies requried for development operations.
	@go install github.com/cespare/reflex@latest
	@go install github.com/git-chglog/git-chglog/cmd/git-chglog@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.44.0
	@go get -u ./...
	@go mod tidy

.PHONY: ci_tests
ci_tests: ## Run tests for CI.
	go fmt ./...
	go vet ./...
	golangci-lint run ./...
	go test -trimpath --timeout=10m -failfast -v -race -covermode=atomic -coverprofile=coverage.out ./...

.PHONY: clean
clean: ## Clean test caches and tidy up modules.
	@go clean -testcache
	@go mod tidy

.PHONY: coverage
coverage: ## Show the test coverage on browser.
	go test -covermode=count -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out | tail -n 1
	go tool cover -html=coverage.out
