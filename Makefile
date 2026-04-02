.PHONY: build lint fmt fix install-hooks

# Build the application
build:
	go build -o rig cmd/rig/main.go

# Run all linters
lint:
	golangci-lint run ./...

# Check formatting
fmt:
	@unformatted=$$(gofmt -l .); \
	if [ -n "$$unformatted" ]; then \
		echo "Unformatted files:"; \
		echo "$$unformatted"; \
		exit 1; \
	fi
	@echo "All files formatted correctly."

# Auto-fix formatting and imports
fix:
	gofmt -w .
	goimports -w -local github.com/mrwhyte/rig .

# Install git hooks
install-hooks:
	cp scripts/pre-commit .git/hooks/pre-commit
	chmod +x .git/hooks/pre-commit
	@echo "Git hooks installed."
