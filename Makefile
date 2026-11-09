.PHONY: build start

build:
	docker build -t stdout.me:latest .

start: build
	docker run --rm -p 8080:8080 --env-file .env stdout.me:latest
