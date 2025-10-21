# ---- Builder ----
FROM golang:1.25.1-alpine AS builder

WORKDIR /app

# Copy module files and download deps
COPY apps/api/go.mod apps/api/go.sum ./
RUN go mod download

# Copy API source code
COPY apps/api/ ./

# Build the application
# CGO_ENABLED=0: build a statically linked binary
# -o /api-server: specify the output file name
# ./cmd/api: specify the main package to build
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /api-server ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /reconcile ./cmd/reconcile

# ---- Runner ----
FROM alpine:latest

# Install ca-certificates for HTTPS and curl for migrate download
RUN apk --no-cache add ca-certificates curl

# Download and install golang-migrate
RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.0/migrate.linux-amd64.tar.gz | tar xvz \
    && mv migrate /usr/local/bin/migrate \
    && chmod +x /usr/local/bin/migrate

# Set working directory
WORKDIR /app

# Copy the compiled binary from the builder stage
COPY --from=builder /api-server /app/api-server
COPY --from=builder /reconcile /app/reconcile

# Copy migration files (context is root, so infra/ is accessible)
COPY infra/migrations /app/migrations

# Expose the port the application runs on
EXPOSE 8080

# Set the entrypoint for the container
ENTRYPOINT ["/app/api-server"]
