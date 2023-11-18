start:
	docker-compose up -d

stop:
	docker-compose down

remove:
	docker-compose down -v --rmi all

logs_db:
	docker-compose logs postgres

.PHNOY: start stop remove logs_db