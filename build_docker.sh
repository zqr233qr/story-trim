#!/bin/bash
set -e

# 请修改为你的 Docker Hub 用户名
DOCKER_USER="kirydocker"
IMAGE_NAME="story-trim"
TAG="latest"

FULL_IMAGE_NAME="$DOCKER_USER/$IMAGE_NAME:$TAG"

echo "🐳 Building Docker image: $FULL_IMAGE_NAME..."
docker build -t $FULL_IMAGE_NAME .

echo "🚀 Pushing to Docker Hub..."
echo "Note: Make sure you have run 'docker login' first."
docker push $FULL_IMAGE_NAME

echo "✅ Done!"
