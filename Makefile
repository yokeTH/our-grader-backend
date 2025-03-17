build:
	go build -o bin/server ./cmd/server/main.go

run:
	go run ./api/cmd/server/main.go

watch:
	air

migrate:
	go run api/cmd/migrate/main.go

clean:
	rm -rf ./bin

deps:
	go mod tidy
