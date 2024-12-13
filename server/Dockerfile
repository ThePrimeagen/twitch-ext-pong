# Start from golang base image
FROM golang:1.18-alpine

# Add git for go mod download
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN go build -o main .

# Expose port 42069 🦍
EXPOSE 42069

# Command to run the executable
CMD ["./main"]
