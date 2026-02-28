GO_BIN=/opt/homebrew/bin/go
BINARY_NAME=go-gin-boilerplate
BINARY_PATH=./bin/$(BINARY_NAME)
MAIN_PATH=./cmd/api/main.go

.PHONY: all run build test tidy clean help swag db.push db.migrate

all: build

## run: Run the application
run:
	$(GO_BIN) run $(MAIN_PATH)

## db.push: Auto-migrate schema based on GORM models (like Prisma db push)
db.push:
	$(GO_BIN) run scripts/db_push.go

## db.migrate: Create the database and necessary initial setup 
db.migrate:
	$(GO_BIN) run scripts/create_db.go

## build: Build the application binary
build:
	$(GO_BIN) build -o $(BINARY_PATH) $(MAIN_PATH)

## test: Run all tests
test:
	$(GO_BIN) test -v ./...

## test-cover: Run tests with coverage
test-cover:
	$(GO_BIN) test -coverprofile=coverage.out ./...
	$(GO_BIN) tool cover -html=coverage.out

## tidy: Tidy go modules
tidy:
	$(GO_BIN) mod tidy

## swag: Generate swagger documentation
swag:
	$(GO_BIN) run github.com/swaggo/swag/cmd/swag init -g $(MAIN_PATH)

## vet: Run go vet
vet:
	$(GO_BIN) vet ./...

## clean: Clean build artifacts
clean:
	rm -rf ./bin
	rm -f coverage.out

## help: Show this help message
help:
	@echo "Usage: make [target]"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'
