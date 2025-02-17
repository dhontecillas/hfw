PACKAGES         ?= $(shell go list ./... | grep -v vendor | grep -v gopath | tr '\n' ' ')
GOTOOLS          ?= \
					golang.org/x/tools/cmd/goimports \
					golang.org/x/tools/cmd/cover \
					github.com/kisielk/errcheck \
					golang.org/x/lint/golint \
					github.com/wadey/gocovmerge \
					github.com/golangci/golangci-lint/cmd/golangci-lint@v1.27.0 \
					honnef.co/go/tools/cmd/staticcheck@latest \
					github.com/mattn/goveralls


build:
	go build -v ./cmd/collectmigrations
.PHONY: build

lint: tools
	# go fmt $(PACKAGES)
	go vet -composites=False $(PACKAGES)
	# errcheck -ignore Close $(PACKAGES)
	# echo $(PACKAGES) | xargs -L1 fgt golint
	staticcheck $(PACKAGES)
	golangci-lint run ./...
	govulncheck
.SILENT: lint
.PHONY: lint

seccheck: lint
	gosec ./...
.SILENT: lint
.PHONY: lint

tools:
	# go install $(GOTOOLS)
	go install honnef.co/go/tools/cmd/staticcheck@latest
	go install github.com/securego/gosec/v2/cmd/gosec@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/vuln/cmd/govulncheck@latest
	go install github.com/mattn/goveralls@latest
.SILENT: tools
.PHONY: tools

bench:
	go test -run=XXX -bench=. $(PACKAGES)
.PHONY: bench


test: DBNAME=testdb-$(shell date +%s)
test: REDISNAME=testredis-$(shell date +%s)
test:
	export TESTREDIS_NAME=$(REDISNAME) && \
	export TESTDB_CONTAINER_NAME=$(DBNAME) && \
	export TESTDB_NAME="hfwtest" && \
	docker run --name $$TESTDB_CONTAINER_NAME -p 0.0.0.0:0:5432 -e POSTGRES_PASSWORD=test -d --rm postgres:latest && \
	docker run --name $$TESTREDIS_NAME -p 0.0.0.0:0:6379 -d --rm redis:alpine && \
	export TESTDB_DOCKER_ID=$$(docker ps -a -f name=$$TESTDB_CONTAINER_NAME --format '{{.ID}}') && \
	export TESTDB_PORT=$$(docker inspect -f '{{range $$p, $$conf := .NetworkSettings.Ports}}{{(index $$conf 0).HostPort}}{{end}}' $$TESTDB_DOCKER_ID) && \
	sleep 2 && \
	PGPASSWORD=test psql -U postgres -p $$TESTDB_PORT -h 127.0.0.1 -c "CREATE DATABASE hfwtest;" && \
	PGPASSWORD=test psql -U postgres -p $$TESTDB_PORT -h 127.0.0.1 -c "CREATE USER hfwtest WITH PASSWORD 'test';" && \
	PGPASSWORD=test psql -U postgres -p $$TESTDB_PORT -h 127.0.0.1 -c "GRANT ALL PRIVILEGES ON DATABASE hfwtest TO hfwtest;" && \
	PGPASSWORD=test psql -U postgres -p $$TESTDB_PORT -h 127.0.0.1 -c "GRANT ALL ON SCHEMA public TO hfwtest;" hfwtest && \
	docker run -v $$PWD/testing/migrations:/migrations --network host migrate/migrate -path=/migrations/ -database postgres://hfwtest:test@localhost:$$TESTDB_PORT/hfwtest?sslmode=disable up && \
	export TESTDB_DOCKER_ID=$$(docker ps -a -f name=$$TESTDB_CONTAINER_NAME --format '{{.ID}}') && \
	export TESTREDIS_DOCKER_ID=$$(docker ps -a -f name=$$TESTREDIS_NAME --format '{{.ID}}') && \
	export TESTREDIS_PORT=$$(docker inspect -f '{{range $$p, $$conf := .NetworkSettings.Ports}}{{(index $$conf 0).HostPort}}{{end}}' $$TESTREDIS_DOCKER_ID) && \
	export NOTIFICATIONS_TEMPLATES_DIR=$$(pwd)/pkg/notifications/templates && \
    echo "TEST DB PORT: $$TESTDB_PORT" && \
	go test -coverprofile=coverage.out $(PACKAGES) ; \
	export TEST_RESULT=$$? ; \
	echo "go tool cover -html=coverage.out"; \
	docker stop $$TESTDB_CONTAINER_NAME ; \
	echo "STOPED DB $$TESTDB_CONTAINER_NAME" ; \
	docker stop $$TESTREDIS_NAME ; \
	echo "STOPED REDIS $$TESTREDIS_NAME" ; \
	exit $$TEST_RESULT

.PHONY: test

dctest:
	docker compose -f docker-compose.test.yml up -d && sleep 4 && \
	export TESTDB_PORT=$$(docker compose -f docker-compose.test.yml port -- hfw_test_db 5432 | cut -d ':' -f 2) && \
	export TESTREDIS_PORT=$$(docker compose -f docker-compose.test.yml port -- hfw_test_redis 6279 | cut -d ':' -f 2) && \
	docker run -v $$PWD/testing/migrations:/migrations --network host migrate/migrate -path=/migrations/ -database postgres://hfwtest:test@localhost:$$TESTDB_PORT/hfwtest?sslmode=disable up && \
	export NOTIFICATIONS_TEMPLATES_DIR=$$(pwd)/pkg/notifications/templates && \
	go test -coverprofile=coverage.out $(PACKAGES) ; \
	export TEST_RESULT=$$? ; \
	docker compose -f docker-compose.test.yml down ; \
	exit $$TEST_RESULT


.PHONY: dctest

check: test lint
.PHONY: check

coverage: tools
	docker compose -f docker-compose.test.yml down; sleep 4; \
	docker compose -f docker-compose.test.yml up -d && sleep 4 && \
	export TESTDB_PORT=$$(docker compose -f docker-compose.test.yml port -- hfw_test_db 5432 | cut -d ':' -f 2) && \
	export TESTREDIS_PORT=$$(docker compose -f docker-compose.test.yml port -- hfw_test_redis 6279 | cut -d ':' -f 2) && \
	docker run -v $$PWD/testing/migrations:/migrations --network host migrate/migrate -path=/migrations/ -database postgres://hfwtest:test@localhost:$$TESTDB_PORT/hfwtest?sslmode=disable up && \
	export NOTIFICATIONS_TEMPLATES_DIR=$$(pwd)/pkg/notifications/templates && \
	mkdir -p docs/coverage && \
	go test -coverprofile=coverage.cov $(PACKAGES) && \
	go tool cover -func=coverage.cov -o coverage.out
	go tool cover -html=coverage.cov -o docs/coverage/coverage.html
	@echo ----------------------------------------
	@echo COVERAGE IS AT: $$(go tool cover -func=coverage.cov | tail -n 1 | rev | cut -d" " -f1 | rev)
	@echo ----------------------------------------
.PHONY: coverage

cicoverage: coverage
	COVERALLS_TOKEN=$$COVERALLS_TOKEN goveralls -coverprofile=coverage.cov

viewcoverage: coverage
	xdg-open docs/coverage/coverage.html
.PHONY: viewcoverage

obs_example:
	go build ./examples/obs_example
.PHONY: obs_example

web_example:
	go build ./examples/web_example
.PHONY: web_example

examples: web_example obs_example
.PHONY: examples
