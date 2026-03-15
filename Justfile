# Default recipe: build the binary
default: build

# Build binary into bin/
build:
    go build -o bin/helm-charts-tui .

# Build with version info embedded
build-release version="dev":
    go build -trimpath -ldflags "-s -w -X main.version={{version}} -X main.commit=$(git rev-parse --short HEAD 2>/dev/null || echo none) -X main.date=$(date -u +%Y-%m-%dT%H:%M:%SZ)" -o bin/helm-charts-tui .

# Run all tests with race detector
test:
    go test -race ./...

# Run tests with verbose output
test-verbose:
    go test -race -v ./...

# Run tests with coverage report
test-coverage:
    go test -race -coverprofile=coverage.out ./...
    go tool cover -func=coverage.out
    rm -f coverage.out

# Run go vet
vet:
    go vet ./...

# Run golangci-lint (install: https://golangci-lint.run/welcome/install/)
lint:
    golangci-lint run ./...

# Tidy module dependencies
tidy:
    go mod tidy

# Format code
fmt:
    gofmt -w .

# Clean build artifacts
clean:
    rm -rf bin/

# Run all checks (tidy, vet, test, build)
check: tidy fmt vet test build

# Run the application
run: build
    ./bin/helm-charts-tui
