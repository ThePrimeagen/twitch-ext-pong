#!/bin/bash
set -e

# Build the Docker image
docker build -t twitch-pong-server .

# Run the container
docker run -p 42069:42069 twitch-pong-server
