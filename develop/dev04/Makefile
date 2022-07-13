# shadow linter: https://pkg.go.dev/golang.org/x/tools/go/analysis/passes/shadow/cmd/shadow
# go install golang.org/x/tools/go/analysis/passes/shadow/cmd/shadow@latest
.PHONY: lint
lint:
	@echo 'staticcheck:'
	@staticcheck ./...
	@echo ' ok'
	@echo 'go vet:'
	@go vet ./...
	@echo ' ok'
	@echo 'shadow:'
	@shadow ./...
	@echo ' ok'
