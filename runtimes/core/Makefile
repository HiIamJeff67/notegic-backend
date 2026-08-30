.PHONY: test test-race test-e2e view-databases view-enums psql migrate seed clear-db remigrate-db

test:
	go test ./...

test-race:
	go test -race ./...

test-e2e:
	go test ./test/e2e/...

view-databases:
	docker compose -f ../../docker-compose.yaml exec -T notegic-core go -C runtimes/core run ./commands viewDatabases

view-enums:
	docker compose -f ../../docker-compose.yaml exec -T notegic-core go -C runtimes/core run ./commands viewAllEnums

psql:
	docker exec -it notegic-db psql -U jeff -d notegic-db

migrate:
	docker compose -f ../../docker-compose.yaml -f ../../docker-compose.init.yaml run --rm --no-deps notegic-core-database-init
	docker compose -f ../../docker-compose.yaml -f ../../docker-compose.init.yaml run --rm --no-deps notegic-durable-job-database-init

seed:
	docker compose -f ../../docker-compose.yaml -f ../../docker-compose.init.yaml run --rm --no-deps notegic-core-database-init

clear-db:
	docker exec -i notegic-db psql -U jeff -d notegic-db -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"

remigrate-db: clear-db migrate
