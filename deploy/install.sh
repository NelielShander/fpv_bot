#!/bin/bash

set -e

PROJECT_DIR="/opt/tg-bot"

echo "=== Installing Docker ==="

if ! command -v docker >/dev/null 2>&1; then
    curl -fsSL https://get.docker.com | sh
fi

echo "=== Creating project directory ==="

mkdir -p "$PROJECT_DIR"

cd "$PROJECT_DIR"

echo "=== Starting services ==="

docker compose up -d --build

echo
echo "=== Containers ==="

docker compose ps

echo
echo "=== Installation completed ==="
echo
echo "Logs:"
echo "  cd $PROJECT_DIR"
echo "  docker compose logs -f bot"