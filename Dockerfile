# Builder stage
FROM golang:1.23-alpine AS builder

WORKDIR /src

# Download dependencies first (better layer caching)
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app ./cmd/app

# Runtime stage - minimal alpine
FROM alpine:3.19

# Install ca-certificates and wget for healthcheck
RUN apk --no-cache add ca-certificates wget

# Copy the binary
COPY --from=builder /app /app

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/healthz || exit 1

# Run the app
ENTRYPOINT ["/app"]
