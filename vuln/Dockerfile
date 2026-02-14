
# Builder stage: use Alpine-based Go image to compile the binary
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git ca-certificates
WORKDIR /src

# Download deps early to leverage layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy sources and build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-s -w" -o /vuln ./cmd


# Final stage: minimal Alpine image
FROM alpine:3.18
RUN apk add --no-cache ca-certificates
WORKDIR /app

# Copy binary and config
COPY --from=builder /vuln /app/vuln
COPY config.yaml /app/config.yaml

# Create a non-root user and run as them
RUN addgroup -S app && adduser -S -G app app && chown -R app:app /app
USER app

ENTRYPOINT ["/app/vuln"]

