start:
	docker-compose up -d

stop:
	docker-compose down

remove:
	docker-compose down -v --rmi all

logs_db:
	docker-compose logs postgres

ifndef NAME
override NAME=init
endif

migrate_create:
	migrate create -ext sql -dir db/migrations $(NAME)

.PHNOY: start stop remove logs_db migrate_create
