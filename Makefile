config=config.yaml

test:
	./scripts/validate-license.sh
	go fmt ./cmd ./pkg/... ./internal/...
	go vet ./cmd ./pkg/... ./internal/...
	go mod tidy
	go test -race ./cmd ./pkg/...
	go run github.com/golangci/golangci-lint/cmd/golangci-lint@latest run -v
run:
	go run -race ./cmd -config=$(config) -logLevel=DEBUG $(args)
test-release:
	go run github.com/goreleaser/goreleaser@latest release --snapshot --skip-publish --rm-dist