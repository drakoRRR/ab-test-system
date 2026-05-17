DB_DSN ?= postgres://abtest:abtest@localhost:5432/abtest?sslmode=disable

# ==================== COMPOSE ====================
.PHONY: compose-up
compose-up:
	docker compose up -d

.PHONY: compose-down
compose-down:
	docker compose down

# ==================== MIGRATIONS ====================
.PHONY: migrate-up
migrate-up:
	go run cmd/migrate/main.go -dsn "$(DB_DSN)" -direction up

.PHONY: migrate-down
migrate-down:
	go run cmd/migrate/main.go -dsn "$(DB_DSN)" -direction down

.PHONY: migrate-down-1
migrate-down-1:
	go run cmd/migrate/main.go -dsn "$(DB_DSN)" -direction down -steps 1

# ==================== API ====================
.PHONY: bundle-api
bundle-api:
	@echo "Bundling control-plane OpenAPI spec..."
	redocly join --config api/redocly.yaml --without-x-tag-groups -o api/_build/openapi.yml
	@echo "Bundling SDK OpenAPI spec..."
	redocly bundle api/sdk.yml -o api/_build/sdk.yml

.PHONY: generate-api
generate-api: bundle-api
	@echo "Generating control-plane API code..."
	oapi-codegen --config=internal/handler/http/api/codegen/config.yaml api/_build/openapi.yml
	@echo "Generating SDK API code..."
	oapi-codegen --config=internal/handler/http/api/codegen/sdk/config.yaml api/_build/sdk.yml

# ==================== MOCKS ====================
.PHONY: generate-mocks
generate-mocks:
	mockery --config mockery.yaml

# ==================== DEPENDENCIES ====================
.PHONY: deps
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

# ==================== BUILD ====================
.PHONY: build
build:
	@echo "Building application..."
	go build -o bin/server cmd/server/main.go

# ==================== RUN ====================
.PHONY: run
run:
	go run cmd/server/main.go -config=./config/config.example.yaml

# ==================== LINT ====================
.PHONY: fmt
fmt:
	golines -w --max-len=120 --base-formatter=gofumpt .
	goimports -l -w -local github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend .

.PHONY: lint
lint: fmt
	golangci-lint run --fix

# ==================== TEST ====================
.PHONY: test
test:
	go test -v -race -coverprofile=coverage.out ./...

.PHONY: test-coverage
test-coverage: test
	go tool cover -html=coverage.out

# ==================== CLEAN ====================
.PHONY: clean
clean:
	@echo "Cleaning generated files..."
	rm -rf api/_build/
	rm -rf bin/
	rm -f coverage.out

# ==================== HELP ====================
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  compose-up      - Start Docker Compose services"
	@echo "  compose-down    - Stop Docker Compose services"
	@echo "  migrate-up      - Run all pending migrations"
	@echo "  migrate-down    - Roll back all migrations"
	@echo "  migrate-down-1  - Roll back one migration step"
	@echo "  bundle-api      - Bundle OpenAPI specs into single file"
	@echo "  generate-api    - Generate Go code from OpenAPI spec"
	@echo "  generate-mocks  - Generate mocks for testing"
	@echo "  deps            - Download and tidy dependencies"
	@echo "  build           - Build the application"
	@echo "  run             - Run the application"
	@echo "  fmt             - Format code (gofumpt + goimports)"
	@echo "  lint            - Format + run linter with auto-fix"
	@echo "  test            - Run tests"
	@echo "  test-coverage   - Run tests with coverage report"
	@echo "  clean           - Clean generated files"
