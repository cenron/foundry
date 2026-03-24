.PHONY: build run test lint docker-up docker-down migrate-up migrate-down test-all swagger web-generate-api generate

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

swagger:
	@test -x $$(go env GOPATH)/bin/swag || { echo "swag not found. Install it: go install github.com/swaggo/swag/cmd/swag@latest"; exit 1; }
	$$(go env GOPATH)/bin/swag init -g cmd/foundry/main.go -o api/swagger

web-generate-api:
	cd web && npm run generate:api

generate: swagger web-generate-api

test-all: test lint
	cd web && npm test && npm run lint && npx playwright test
