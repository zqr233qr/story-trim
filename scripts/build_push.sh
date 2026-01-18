#!/bin/bash
set -e

# ================= Configuration =================
# Docker Hub Namespace
REGISTRY="kirydocker"
IMAGE_NAME="story-trim"
TAG=$(date +%Y%m%d-%H%M)
# =================================================

FULL_IMAGE_NAME="$REGISTRY/$IMAGE_NAME:$TAG"
LATEST_IMAGE_NAME="$REGISTRY/$IMAGE_NAME:latest"

echo "[1/3] Building Docker image..."
# Use BuildKit for faster builds
DOCKER_BUILDKIT=1 docker build -t $FULL_IMAGE_NAME -t $LATEST_IMAGE_NAME .

echo "[2/3] Pushing images to registry..."
echo "      Pushing $TAG..."
docker push $FULL_IMAGE_NAME
echo "      Pushing latest..."
docker push $LATEST_IMAGE_NAME

echo "[3/3] Done!"
echo "------------------------------------------------"
echo "Image: $FULL_IMAGE_NAME"
echo "Run on server:"
echo "docker run -d -p 8080:8080 -v \$(pwd)/config.yaml:/app/config.yaml -v \$(pwd)/data:/app/data $LATEST_IMAGE_NAME"
