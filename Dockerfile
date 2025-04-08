FROM golang:1.24-alpine3.21 AS builder

WORKDIR /app
COPY . .

# Build the binary
RUN go mod download
RUN go build -o action-control

# Create a minimal image
FROM alpine:3.21

RUN apk --no-cache add ca-certificates

WORKDIR /app
COPY --from=builder /app/action-control /usr/local/bin/action-control

# Create entrypoint script to handle boolean flags
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]