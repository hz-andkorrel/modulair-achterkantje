# HotelHub: Backend (Modulair Achterkantje)

This repository contains the backend code for the HotelHub.
The broker is the main component of the hub and manages communications between the front-end and various plugins.
An eventbus is used to facilitate communication between different plugins in a decoupled manner.
This allows plugins to respond to each other's events without direct dependencies.


## Running the Development Environment

The development environment can be started using Docker Compose.
It uses the `compose.dev.yml` configuration file.
The containers are configured to build the application internally through volumes.
Removing the need to rebuild the Docker images after every code change.
To start the development environment, run the following command:

```bash
docker-compose -f compose.dev.yml up 
```

There are three services defined in the `compose.dev.yml` file.
`postgres` is the database service, `broker` and `eventbus` speak for themselves.
No aditional configuration is required to get started from scratch.

To **recreate the database** and start fresh, the folder `data/postgres_dev` should be deleted.
Emptying this folder is not sufficient, as Postgres keeps some metadata files that will prevent a clean start.
