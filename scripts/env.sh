#!/bin/bash

export POSTGRESQL_URL='postgres://hfw:dev@localhost:5432/hfw?sslmode=disable'

# password for `postgres` user is the same
export POSTGRESQL_PASSWORD=dev
export POSTGRESQL_USER=hfw

# mailer config: output to console
export MAILER_PREFERRED=console
export MAILER_LOGS=true

# db config:
export DB_SQL_MASTER_NAME=hfw
export DB_SQL_MASTER_HOST='127.0.0.1'
export DB_SQL_MASTER_PORT=5432
export DB_SQL_MASTER_USER=hfw
export DB_SQL_MASTER_PASS=dev

alias mup='migrate -database ${POSTGRESQL_URL} -path pkg/usecases/users/migrations up'
alias mdown='migrate -database ${POSTGRESQL_URL} -path pkg/usecases/users/migrations down'
