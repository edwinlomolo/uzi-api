# Tidy go packages
tidy:
	go mod tidy
# Generate graphql resolvers and models
gql:
	go run github.com/99designs/gqlgen generate --verbose
# migrate-db:
# 	sqlc generate
# 
server:
	go run cmd/main.go
test:
	go test -v ./...
