# Overview

## Prerequisites installation

- Dbeaver Universal database GUI tool: [Download](https://dbeaver.io/download/)
- Make:
  - [Download](https://gnuwin32.sourceforge.net/packages/make.htm) or `choco install make`
  - `make --version`
- `golang-migrate`:
  - [Download](https://github.com/golang-migrate/migrate/tree/master/cmd/migrate)
  - `migrate --version`
- `sqlc` (if not to be used as docker)
  - [Download](https://docs.sqlc.dev/en/latest/overview/install.html)
  - `sqlc version`
- `gomock`:
  - [Install] `go install go.uber.org/mock/mockgen@latest`
  - `mockgen --version`

## Run docker containers locally

To run container, please follow below steps:

- Download docker

  - [Windows](https://docs.docker.com/desktop/install/windows-install/)
  - [Ubuntu](https://docs.docker.com/engine/install/ubuntu/#install-using-the-repository)

- Run docker containers

```bash
make start
```

- To stop docker containers

```bash
make stop
```

- To remove docker containers

```bash
make remove
```

- To see containers logs

```bash
make logs_db
make logs_sqlc
```

- To connect to Postgres database manually

```bash
 docker exec -it postgres bash
 psql -U admin -d postgresdb
```

- To see all tables in database

```bash
\dt;
```

## Migration

- To create migration

```bash
make migrate_create
make migrate_create NAME=init
```

- To apply migration

```bash
make migrate_up
make migrate_down
```

- Code generation

```bash
Code from db/queries is automatically generated using sqlc docker container which has command `generate`.
```

## Server

- To start server

```bash
make server
```

## Test

- To run unit tests

```bash
make test
```
