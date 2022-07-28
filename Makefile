PACKAGES         ?= $(shell go list ./... | grep -v vendor | grep -v gopath)
GOTOOLS          ?= github.com/GeertJohan/fgt\
					golang.org/x/tools/cmd/goimports\
					github.com/kisielk/errcheck \
					golang.org/x/lint/golint \
					github.com/wadey/gocovmerge \
					github.com/golangci/golangci-lint/cmd/golangci-lint@v1.27.0 \
					honnef.co/go/tools/cmd/staticcheck@latest

build:
	go build -v ./cmd/importer
.PHONY: build

lint: tools
	fgt go fmt $(PACKAGES)
	fgt go vet -composites=False $(PACKAGES)
	fgt errcheck -ignore Close $(PACKAGES)
	echo $(PACKAGES) | xargs -L1 fgt golint
	staticcheck $(PACKAGES)
.SILENT: lint
.PHONY: lint

tools:
	go get $(GOTOOLS)
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
	docker run --name $$TESTDB_CONTAINER_NAME -p 0.0.0.0:0:5432 -e POSTGRES_PASSWORD=test -d --rm postgres:12.3 && \
	docker run --name $$TESTREDIS_NAME -p 0.0.0.0:0:6379 -d --rm redis:alpine && \
	export TESTDB_DOCKER_ID=$$(docker ps -a -f name=$$TESTDB_CONTAINER_NAME --format '{{.ID}}') && \
	export TESTDB_PORT=$$(docker inspect -f '{{range $$p, $$conf := .NetworkSettings.Ports}}{{(index $$conf 0).HostPort}}{{end}}' $$TESTDB_DOCKER_ID) && \
	sleep 2 && \
	PGPASSWORD=test psql -U postgres -p $$TESTDB_PORT -h 127.0.0.1 -c "CREATE DATABASE hfwtest" && \
	PGPASSWORD=test psql -U postgres -p $$TESTDB_PORT -h 127.0.0.1 -c "CREATE USER hfwtest WITH PASSWORD 'test';" && \
	PGPASSWORD=test psql -U postgres -p $$TESTDB_PORT -h 127.0.0.1 -c "GRANT ALL PRIVILEGES ON DATABASE hfwtest TO hfwtest;" && \
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
	docker-compose -f docker-compose.test.yml up -d && sleep 4 && \
	export TESTDB_PORT=$$(docker-compose -f docker-compose.test.yml port -- hfw_test_db 5432 | cut -d ':' -f 2) && \
	export TESTREDIS_PORT=$$(docker-compose -f docker-compose.test.yml port -- hfw_test_redis 6279 | cut -d ':' -f 2) && \
	docker run -v $$PWD/testing/migrations:/migrations --network host migrate/migrate -path=/migrations/ -database postgres://hfwtest:test@localhost:$$TESTDB_PORT/hfwtest?sslmode=disable up && \
	export NOTIFICATIONS_TEMPLATES_DIR=$$(pwd)/pkg/notifications/templates && \
	go test -coverprofile=coverage.out $(PACKAGES) ; \
	export TEST_RESULT=$$? ; \
	docker-compose -f docker-compose.test.yml down ; \
	exit $$TEST_RESULT


.PHONY: dctest

check: test lint
.PHONY: check

coverage: tools
	docker-compose -f docker-compose.test.yml up -d && sleep 4 && \
	export TESTDB_PORT=$$(docker-compose -f docker-compose.test.yml port -- hfw_test_db 5432 | cut -d ':' -f 2) && \
	export TESTREDIS_PORT=$$(docker-compose -f docker-compose.test.yml port -- hfw_test_redis 6279 | cut -d ':' -f 2) && \
	docker run -v $$PWD/testing/migrations:/migrations --network host migrate/migrate -path=/migrations/ -database postgres://hfwtest:test@localhost:$$TESTDB_PORT/hfwtest?sslmode=disable up && \
	export NOTIFICATIONS_TEMPLATES_DIR=$$(pwd)/pkg/notifications/templates && \
	mkdir -p coverage && \
	mkdir -p docs/coverage && \
	$(foreach pkg,$(PACKAGES),\
	go test $(pkg) -coverprofile="coverage/$(shell echo $(pkg) | tr "\/" _)" -coverpkg=$(go list ./... | grep -v /assets/ | paste -sd "," -) -covermode=set;)
	gocovmerge coverage/* > coverage/aggregate.coverprofile
	go tool cover -html=coverage/aggregate.coverprofile -o docs/coverage/coverage.html
	@echo ----------------------------------------
	@echo COVERAGE IS AT: $$(go tool cover -func=coverage/aggregate.coverprofile | tail -n 1 | rev | cut -d" " -f1 | rev)
	@echo ----------------------------------------
	rm -rf coverage
.PHONY: coverage

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
