## About

My own GithubAction

## Requirements

- Golang(version 1.19)
- Docker and Docker compose
- Act installed. Tip: https://nektosact.com/installation/index.html to install act in your OS.
- Pm2 installed. Tip: can be helpfull tool to keep to golang api and job process running forever.

## Instructions

- Clone project
- Execute command **bash scripts/setupLocally.sh** to setup to create directories: logs and pipelines and sqlite file named **database**
- Execute command **docker-compose up -d** to run redis container. I'm using redis as queue in that project.
- Execute command **go run cmd/api/main.go** to start api at address http://localhost:3000 .
- Execute command **go run cmd/job/main.go** to start job process when is reponsable to consume message from the queue and execute GithubAction pipeline.
- Execute command **bash scripts/deploy.sh** to build the project and execute using pm2.


## Architecture

![the project architecture](./architecture.png)