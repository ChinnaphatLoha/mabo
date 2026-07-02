run:
	go run ./cmd/server

build:
	go build -o bin/server ./cmd/server

docker:
	docker compose -f ../docker/docker-compose.yml up --build

fmt:
	go fmt ./...

test:
	go test ./...