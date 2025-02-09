# Use an official Golang runtime image
FROM golang:1.21

# Set working directory inside the container
WORKDIR /app

# Copy Go files
COPY . .

# Download dependencies
RUN go mod tidy

# Build the Go application
RUN go build -o main .

# Expose port 1123 for Prometheus metrics
EXPOSE 1123

# Run the app
CMD ["./main"]
