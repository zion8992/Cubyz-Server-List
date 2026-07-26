# Cubyz Server List

## What is this?
This repository contains **two pieces of software**: Ironite and Silverite.

**Ironite**: Generates a static website with a list of servers
**Silverite**: Manages and updates a list of servers

## Ironite

Generates a static list of servers plus a server page for each server with the server details.

### Running
```sh
cd src/ironite
./generate.sh
```

That will generate the list into the `public/` directory. You can change the list of servers in `servers-NUM.json`. `NUM` is the file number to avoid large numbers.

## Silverite

Silverite is a Go webserver for managing a set of server JSON lists. It has both a web UI interface where admins can login to manage and approve servers (if approvals are enabled) and also an API for the Cubyz game to create a server and get servers.

### Running
```sh
cd src/ironite
./run.sh
```