migrateup:
	goose -dir ./migrations postgres "host=localhost user=postgres database=orders_data password=postgres sslmode=disable" up
migratedown:
	goose -dir ./migrations postgres "host=localhost user=postgres database=orders_data password=postgres sslmode=disable" down
docker up:
	docker-compose up -d
docker down:
	docker-compose down