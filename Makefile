# Tidy go packages
tidy:
	go mod tidy
# Generate graphql resolvers and models
gql:
	go run github.com/99designs/gqlgen generate --verbose
# Sqlc
sqlc:
	sqlc generate
# Test
test:
	go test -v ./...
# Run server
server:
	go run cmd/main.go
