PLUGIN_ID    := davidmillerpf9-gaia-adql-datasource
DIST_DIR     := dist
BACKEND_BIN  := $(DIST_DIR)/gpx_gaia_adql

# Build the backend for both linux/amd64 and linux/arm64 (multi-arch Docker support)
LINUX_AMD64  := $(BACKEND_BIN)_linux_amd64
LINUX_ARM64  := $(BACKEND_BIN)_linux_arm64

.PHONY: all build build-backend build-frontend build-backend-linux start stop restart logs clean help

all: build

## build: Build frontend and backend (linux binaries for Docker)
build: build-frontend build-backend-linux
	@echo "Build complete → $(DIST_DIR)/"

## build-frontend: Compile TypeScript/React frontend via webpack
build-frontend: node_modules
	@echo "Building frontend..."
	npm run build

node_modules: package.json
	npm install
	@touch node_modules

## build-backend: Build backend binary for the current OS (for running outside Docker)
build-backend:
	@echo "Building backend ($(shell go env GOOS)/$(shell go env GOARCH))..."
	@mkdir -p $(DIST_DIR)
	go build -o $(BACKEND_BIN) ./pkg/main.go

## build-backend-linux: Cross-compile backend for linux/amd64 and linux/arm64 (required for Docker)
build-backend-linux:
	@echo "Cross-compiling backend for linux/amd64..."
	@mkdir -p $(DIST_DIR)
	GOOS=linux GOARCH=amd64 go build -o $(LINUX_AMD64) ./pkg/main.go
	@echo "Cross-compiling backend for linux/arm64..."
	GOOS=linux GOARCH=arm64 go build -o $(LINUX_ARM64) ./pkg/main.go
	@echo "Backend binaries: $(LINUX_AMD64), $(LINUX_ARM64)"

## start: Start Grafana in Docker (builds first)
start: build
	@echo "Starting Grafana at http://localhost:3000 (admin/admin)..."
	docker compose up -d
	@echo "Grafana is starting — run 'make logs' to follow output"

## stop: Stop and remove the Grafana container
stop:
	docker compose down

## restart: Rebuild plugin and restart Grafana
restart: build stop start

## logs: Follow Grafana container logs
logs:
	docker compose logs -f grafana

## dev: Watch frontend for changes and restart Grafana on rebuild
dev: build-backend-linux
	@echo "Starting Grafana..."
	docker compose up -d
	@echo "Watching frontend for changes (Ctrl-C to stop)..."
	npm run dev

## test: Run all tests (Go + Jest)
test: test-backend test-frontend

## test-backend: Run Go tests
test-backend:
	go test ./pkg/... -v

## test-frontend: Run Jest tests
test-frontend:
	npm test

## sign: Sign the plugin for local dev (requires GRAFANA_API_KEY env var from grafana.com/auth/api-keys)
sign: build-frontend
	GRAFANA_API_KEY=$(GRAFANA_API_KEY) npx --yes @grafana/sign-plugin@latest --rootUrls http://localhost:3000
	@echo "Plugin signed — MANIFEST.txt written to dist/. Restart with 'make restart'."

## clean: Remove build artifacts
clean:
	rm -rf $(DIST_DIR)
	docker compose down -v 2>/dev/null || true

## help: Show this help
help:
	@grep -E '^## ' Makefile | sed 's/## /  /'
