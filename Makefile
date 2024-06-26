# Tidy go packages
tidy:
	go mod tidy
# Generate graphql resolvers and models
graphql:
	go run github.com/99designs/gqlgen generate --verbose
# Sqlc
sqlc:
	sqlc generate
# Test
test:
	go test -v ./...
# Run server
server:
	go run server.go
