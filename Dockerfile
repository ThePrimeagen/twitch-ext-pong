# Start from golang base image
FROM golang:1.21-alpine

# Add git for go mod download
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy server files
COPY server/ ./server/
COPY src/ ./src/

# Build the application
RUN cd server && go mod download && go build -o main && cd ..

# Expose port 42069 🦍
EXPOSE 42069

# Command to run the executable
CMD ["./server/main"]
