# An official Go runtime
FROM golang:1.19

ENV DBDRIVER=$DBDRIVER
ENV DATABASE_URI=$DATABASE_URI
ENV POSTAL_DATABASE_URI=$POSTAL_DATABASE_URI
ENV ACCESS_KEY=$ACCESS_KEY
ENV SECRET_ACCESS_KEY=$SECRET_ACCESS_KEY
ENV S3_CARETAKER_BUCKET=$S3_CARETAKER_BUCKET
# Server
ENV SERVERPORT=$SERVERPORT
# Server env
ENV SERVERENV=$SERVERENV
# Jwt
ENV JWTEXPIRE=$JWTEXPIRE
# generated by - (pwgen -s -1 64)
ENV JWTSECRET=$JWTSECRET

# Create an application directory
RUN mkdir -p go/src/app

# Set working directory to /go/src/app
WORKDIR go/src/app

# Copy current dir contents into the container at go/src/app
COPY . .

# Install dependencies
RUN go mod download

# Build the Go app
RUN CGO_ENABLED=0 GOOS=linux go build -o uzi

# Expose port 4000
EXPOSE 4000

# Run the command be default when container starts
CMD ["./uzi"]
