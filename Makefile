.PHONY: dev gen test build

# Live reload on :7331 (proxies the app on :3000).
dev:
	go tool air

# Regenerate templ components and GORM CLI query code.
gen:
	go tool templ generate
	go generate ./internal/db

test: gen
	go vet ./...
	go test ./...

build: gen
	go build -o bin/server ./cmd/server
