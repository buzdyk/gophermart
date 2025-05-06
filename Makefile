.PHONY: migrate build-linux

migrate:
	docker compose exec gophermart migrate -path /app/migrations -database "postgres://postgres:secret@postgres:5432/gophermart?sslmode=disable" up

build-linux:
	GOOS=linux GOARCH=amd64 go build -o ./cmd/gophermart/gophermart ./cmd/gophermart