# Load variables from the .env file
include app.env
export

start:
	docker-compose up -d

stop:
	docker-compose down

remove:
	docker-compose down -v

destroy:
	docker-compose down -v --rmi all

logs_db:
	docker-compose logs postgres

logs_sqlc:
	docker-compose logs sqlc

ifndef NAME
override NAME=init
endif

migrate_create:
	migrate create -ext sql -dir db/migrations $(NAME)

migrate_up:
	migrate -path db/migrations -database $(DATABASE_URL) up

migrate_down:
	migrate -path db/migrations -database $(DATABASE_URL) down

build:
	go build -v ./...

test:
	go test -v -cover ./...

server:
	go run main.go

mock:
	mockgen -source="db/sqlc/store.go" -destination="db/mock/store.go" -package mockdb

.PHNOY: start stop remove destroy logs_db logs_sqlc migrate_create migrate_up migrate_down build test server mock
