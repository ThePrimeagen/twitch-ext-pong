# Start from golang base image
FROM golang:1.21-alpine

# Add git for go mod download
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy server files
COPY server/ ./server/
COPY src/ ./src/

# Set up Go workspace
WORKDIR /app/server

# Download all dependencies
RUN go mod download

# Build the application
RUN go build -o main .

# Expose port 42069 🦍
EXPOSE 42069

# Command to run the executable
CMD ["./main"]
