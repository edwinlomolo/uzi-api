# init:
# 	go run github.com/99designs/gqlgen init --verbose
# 
# generate:
# 	go run github.com/99designs/gqlgen --verbose
# 
# migrate-db:
# 	sqlc generate
# 
server:
	go run .
test:
	go test -v ./...
