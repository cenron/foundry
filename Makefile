.PHONY: build run test lint docker-up docker-down migrate-up migrate-down test-all

build:
	go build -o bin/foundry ./cmd/foundry

run:
	air

test:
	go test ./... -v

lint:
	golangci-lint run

docker-up:
	docker compose -f build/docker-compose.yml up -d

docker-down:
	docker compose -f build/docker-compose.yml down

migrate-up:
	go run ./cmd/foundry migrate up

migrate-down:
	go run ./cmd/foundry migrate down

test-all: test lint
	cd web && npm test && npm run lint && npx playwright test
