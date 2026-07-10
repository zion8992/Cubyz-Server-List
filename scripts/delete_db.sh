#!/bin/bash

echo "DELETING DATABASE..."

CONTAINER_NAME="cubyzListDB"
VOLUME_NAME="cubyzListDB"

# Stop container if running
if [ "$(docker ps -q -f name=^/${CONTAINER_NAME}$)" ]; then
    echo "Stopping container..."
    docker stop $CONTAINER_NAME
fi

# Remove container if exists
if [ "$(docker ps -a -q -f name=^/${CONTAINER_NAME}$)" ]; then
    echo "Removing container..."
    docker rm $CONTAINER_NAME
fi

# Remove volume if exists
if [ "$(docker volume ls -q -f name=^${VOLUME_NAME}$)" ]; then
    echo "Removing volume..."
    docker volume rm $VOLUME_NAME
fi

echo "Your Ironite database just got deleted :)"