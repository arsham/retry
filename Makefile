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
	@go get -u github.com/cespare/reflex
	@go get github.com/golangci/golangci-lint/cmd/golangci-lint@v1.28.3
	@go get -u github.com/git-chglog/git-chglog/cmd/git-chglog
	@go mod tidy


.PHONY: changelog
changelog: ## Update the changelog.
	@git-chglog > CHANGELOG.md
	@echo "Changelog has been updated."


.PHONY: changelog_release
changelog_release: ## Update the changelog with a release tag.
	@git-chglog --next-tag $(tag) > CHANGELOG.md
	@echo "Changelog has been updated."


.PHONY: clean
clean: ## Clean test caches and tidy up modules.
	@go clean -testcache
	@go mod tidy
