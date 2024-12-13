#!/bin/bash

# 🦍 STRONK DOCKER SCRIPT 🦍

# Stop any existing container
docker stop twitch-pong-server 2>/dev/null || true
docker rm twitch-pong-server 2>/dev/null || true

# Build fresh image
echo "🦍 BUILDING STRONK DOCKER IMAGE 🦍"
docker build -t twitch-pong-server .

# Run container
echo "🦍 RUNNING STRONK SERVER ON PORT 42069 🦍"
docker run --name twitch-pong-server -p 42069:42069 twitch-pong-server
