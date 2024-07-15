create:
	migrate create -ext sql -dir database/migration/ -seq create_book_schema

up:
	migrate -path=database/migration -database="mysql://root:@/bookstore?parseTime=true" up
migration_up:
	migrate -path database/migration/ -database "mysql://root:@/bookstore?parseTime=true" -verbose up

migration_down:
	migrate -path database/migration/ -database "mysql://root:@/bookstore?parseTime=true" -verbose down

migration_fix:
	migrate -path database/migration/ -database "mysql://root:@/bookstore?parseTime=true" force VERSION

