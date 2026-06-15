.PHONY: test lint build run coverage

test:
	cd backend && go test ./... -v -coverprofile=coverage.out
	cd frontend && npm run test -- --run

lint:
	cd backend && go vet ./...
	cd frontend && npm run lint

build:
	cd backend && go build -o bin/app ./cmd/server
	cd frontend && npm run build

run:
	cd backend && go run ./cmd/server

coverage:
	cd backend && go test ./... -coverprofile=coverage.out -covermode=atomic && go tool cover -func=coverage.out
	cd frontend && npm run test -- --run --coverage
