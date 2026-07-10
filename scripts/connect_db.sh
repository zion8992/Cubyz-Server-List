#!/bin/bash

CONTAINER_NAME="cubyzListDB"
MYSQL_ROOT_PASSWORD="H0EeLfLnO,xDEVELOPERSx4c!#%"

if [ -z "$(printenv mysqlpass)" ]; then
    mysqlpass="$MYSQL_ROOT_PASSWORD"
fi

if ! docker info >/dev/null 2>&1; then
    echo "Docker is not running or not installed."
    exit 1
fi

if [ ! "$(docker ps -q -f name=^/${CONTAINER_NAME}$)" ]; then
    echo "Container '$CONTAINER_NAME' is not running. Start it first."
    exit 1
fi

docker exec -it "$CONTAINER_NAME" mysql -u root -p"$mysqlpass"
