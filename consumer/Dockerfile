# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install required tools
RUN apk add --no-cache git

# Copy source code and dependencies
COPY temporal-replacement/consumer/ ./consumer/
COPY machinery/v2/ ./machinery/v2/

# Create a fresh go.mod file
WORKDIR /app/consumer
RUN rm -f go.mod go.sum
RUN go mod init temporal-replacement/consumer
RUN go mod edit -require github.com/RichardKnop/machinery/v2@v2.0.13
RUN go mod edit -replace github.com/RichardKnop/machinery/v2=/app/machinery/v2

# Generate go.sum file and download dependencies
RUN go mod tidy
RUN go mod download

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o consumer .

# Runtime stage
FROM alpine:3.17

WORKDIR /app
COPY --from=builder /app/consumer/consumer .
RUN apk --no-cache add ca-certificates

ENTRYPOINT ["/app/consumer"]