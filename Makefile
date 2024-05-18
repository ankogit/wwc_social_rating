.PHONY:

build:
	go build -o ./.bin/bot cmd/main.go

run: build
	./.bin/bot

buildAndRun:
	make build && make run

build-image:
	docker build -t template_bot:0.1 .

start-container:
	docker run --env-file .env -p 80:80 template_bot:0.1

build-compose:
	docker-compose

run-compose:
	docker-compose up

build-run-compose:
	docker-compose up -d --build