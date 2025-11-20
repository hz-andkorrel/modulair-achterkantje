FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY broker/go.mod broker/go.sum ./
RUN go mod download

# Copy source code
COPY broker/ ./

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o broker .

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/broker .

# Create data directory
RUN mkdir -p /data

EXPOSE 8081

CMD ["./broker"]
