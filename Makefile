build:
	@docker-compose up --build

run:
	@docker-compose up

test:
	@go test ./... -v