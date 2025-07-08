postgres:
	docker run --name postgres17 --network bank-network -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:17-alpine

createdb:
	docker exec -it postgres17 createdb --username=root --owner=root simple_bank

dropdb:
	docker exec -it postgres17 dropdb simple_bank

migrateup:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable" -verbose up

migratedown:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable" -verbose down

migrateup1:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable" -verbose up 1

migratedown1:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable" -verbose down 1

createtestdb:
	docker exec -it postgres17 createdb --username=root --owner=root simple_bank_test
	
droptestdb:
	docker exec -it postgres17 dropdb simple_bank_test

migratetestup:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/simple_bank_test?sslmode=disable" -verbose up

migratetestdown:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/simple_bank_test?sslmode=disable" -verbose down

migratetestup1:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/simple_bank_test?sslmode=disable" -verbose up 1

migratetestdown1:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/simple_bank_test?sslmode=disable" -verbose down 1

sqlc:
	sqlc generate

test:
	go test -v -cover ./...

server:
	go run main.go

mock:
	mockgen -package mockdb -destination db/mock/store.go github.com/Drolfothesgnir/simplebank/db/sqlc Store

.PHONY: postgres createdb dropdb migrateup sqlc test createtestdb droptestdb migratetestup migratetestdown server mock migrateup1 migratedown1 migratetestup1 migratetestdown1