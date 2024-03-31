all:
	docker compose up -d --build

down:
	docker compose down

up:
	docker compose up

build:
	docker compose build
