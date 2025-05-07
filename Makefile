.PHONY: migrate build-linux

migrate:
	docker-compose up migrate

build-linux:
	GOOS=linux GOARCH=amd64 go build -o ./cmd/gophermart/gophermart ./cmd/gophermart