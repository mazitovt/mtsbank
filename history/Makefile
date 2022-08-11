migration.up:
	migrate -path migrations/ -database postgres://user:password@localhost:5432/dbname?sslmode=disable up
migration.down:
	migrate -path migrations/ -database postgres://user:password@localhost:5432/dbname?sslmode=disable down