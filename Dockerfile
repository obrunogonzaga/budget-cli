# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o financli cmd/main.go

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates
WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/financli .

# Set environment variables
ENV MONGODB_URI=mongodb://localhost:27017
ENV MONGODB_DATABASE=financli

# Expose port (if needed for future web interface)
EXPOSE 8080

# Run the application
CMD ["./financli"]