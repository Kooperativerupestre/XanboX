include .env
export

migrate-up:
	migrate -path migrations -database "$(DATABASE_URL)" up
	migrate -path migrations -database "$(TEST_DATABASE_URL)" up

migrate-down:
	migrate -path migrations -database "$(DATABASE_URL)" down 1
	migrate -path migrations -database "$(TEST_DATABASE_URL)" down 1

migrate-version:
	migrate -path migrations -database "$(DATABASE_URL)" version
	migrate -path migrations -database "$(TEST_DATABASE_URL)" version

run:
	go run ./cmd/server