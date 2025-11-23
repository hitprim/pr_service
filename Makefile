.DEFAULT_GOAL := docker-up

build:
	go build -o pr_service ./cmd/app

run: build
	./pr_service

docker-up:
	docker-compose up --build

clean:
	if exist pr_service del pr_service


