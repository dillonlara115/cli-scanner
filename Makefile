.PHONY: build test install clean release frontend-build frontend-dev serve

# Build the binary (requires frontend to be built first)
build: frontend-build
	go build -o bin/barracuda .

# Run tests
test:
	go test ./... -v

# Run tests with coverage
test-coverage:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out

# Install to $GOPATH/bin (requires frontend to be built first)
install: frontend-build
	go install .

# Install alias to ~/.zshrc
install-alias:
	@echo "Adding alias to ~/.zshrc..."
	@if ! grep -q "alias barracuda=" ~/.zshrc 2>/dev/null; then \
		echo "alias barracuda=\"$(shell pwd)/bin/barracuda\"" >> ~/.zshrc; \
		echo "✓ Alias added to ~/.zshrc"; \
		echo "Run 'source ~/.zshrc' or restart your terminal to use it"; \
	else \
		echo "⚠️  Alias already exists in ~/.zshrc"; \
	fi

# Add bin directory to PATH in ~/.zshrc
install-path:
	@echo "Adding bin directory to PATH in ~/.zshrc..."
	@if ! grep -q "$(shell pwd)/bin" ~/.zshrc 2>/dev/null; then \
		echo 'export PATH="$(shell pwd)/bin:$$PATH"' >> ~/.zshrc; \
		echo "✓ PATH updated in ~/.zshrc"; \
		echo "Run 'source ~/.zshrc' or restart your terminal to use it"; \
	else \
		echo "⚠️  PATH already includes bin directory"; \
	fi

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f coverage.out

# Build for multiple platforms (requires frontend to be built first)
release: frontend-build
	GOOS=linux GOARCH=amd64 go build -o bin/barracuda-linux-amd64 .
	GOOS=linux GOARCH=arm64 go build -o bin/barracuda-linux-arm64 .
	GOOS=darwin GOARCH=amd64 go build -o bin/barracuda-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build -o bin/barracuda-darwin-arm64 .
	GOOS=windows GOARCH=amd64 go build -o bin/barracuda-windows-amd64.exe .

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	golangci-lint run

# Run benchmarks
bench:
	go test -bench=. -benchmem ./...

# Build frontend
frontend-build:
	cd web && npm install && npm run build

# Run frontend in dev mode
frontend-dev:
	cd web && npm install && npm run dev

# Serve results (requires built frontend)
serve:
	go run . serve --results results.json --graph graph.json

