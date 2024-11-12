build:
	go build -o srv ./main.go
.PHONY: build

docker_up:
	mkdir -p ./compose/tmp
	cd ./compose && docker-compose -p obs_example_hfw -f docker-compose.yml up -d 
.PHONY: docker_up

docker_stop:
	cd ./compose && docker-compose -p obs_example_hfw -f docker-compose.yml stop -d 
.PHONY: docker_stop

docker_down:
	cd ./compose && docker-compose -p obs_example_hfw -f docker-compose.yml down --remove-orphans -v
.PHONY: docker_down

run:
	OBSEXAMPLE_LOGFILE=./compose/tmp/example_log.txt ./srv | tee log_frontend.log
.PHONY: run 

dbg:
	dlv debug ./main.go --wd .
.PHONY: dbg
