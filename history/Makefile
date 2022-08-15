migration.up:
	migrate -path migrations/ -database postgres://user:password@localhost:5432/dbname?sslmode=disable up
migration.down:
	migrate -path migrations/ -database postgres://user:password@localhost:5432/dbname?sslmode=disable down

# docker network for containers
net = all-network-1

lint:
	golangci-lint run

test:
	go test ./...

gen:
	oapi-codegen -config api/http/v1/config.yaml api/http/v1/swagger.yaml > ./internal/api/http/v1/service.gen.go

build:
	go mod download && CGO_ENABLED=0 GOOS=linux go build -o ./.bin/app ./cmd/main.go

start: build
	SERVICE_NETWORK=$(net) docker compose up -d --remove-orphans --build

stop:
	SERVICE_NETWORK=$(net) docker compose down