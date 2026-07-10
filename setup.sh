#!/bin/bash

set -e

CONTAINER_NAME="cubyzListDB"
VOLUME_NAME="cubyzListDB"
SCRIPTS_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPTS_DIR" && pwd)"

# Generate a strong random password if not already set
if [ -z "${mysqlpass}" ]; then
    mysqlpass=$(openssl rand -hex 32)
fi

echo "Production MySQL setup"
echo "Generated root password: $mysqlpass"
echo "Save this password. It will be needed to connect and to run Ironite."

if ! docker info >/dev/null 2>&1; then
    echo "Docker is not running or not installed."
    exit 1
fi

# Remove existing container and volume for clean production setup
if [ "$(docker ps -a -q -f name=^/${CONTAINER_NAME}$)" ]; then
    echo "Existing container found. Removing it for clean production setup..."
    docker stop "$CONTAINER_NAME" >/dev/null 2>&1 || true
    docker rm "$CONTAINER_NAME" >/dev/null 2>&1 || true
fi

if [ "$(docker volume ls -q -f name=^${VOLUME_NAME}$)" ]; then
    echo "Existing volume found. Removing it for clean production setup..."
    docker volume rm "$VOLUME_NAME" >/dev/null 2>&1 || true
fi

echo "Creating volume..."
docker volume create "$VOLUME_NAME"

echo "Creating and starting MySQL container..."
docker run -d \
    --name "$CONTAINER_NAME" \
    -e MYSQL_ROOT_PASSWORD="$mysqlpass" \
    -v "$VOLUME_NAME:/var/lib/mysql" \
    -p 3306:3306 \
    mysql:8.0


echo "Database setup complete."
echo "Run ./scripts/connect_db.sh and paste in the"
echo "contents of database_setup.sql into the MySQL console"
echo ""
echo "Run Ironite with:"
echo "  ./ironite-linux-amd64 -dbpass \"$mysqlpass\""
echo ""
echo "To stop the database:"
echo "./scripts/stop_db.sh"
