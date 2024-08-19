## About

The project is one api to allow you have your own Github actions to run your Github action pipelines.

## Requirements

- Golang(version 1.19)
- Docker and Docker compose
- Act installed. Tip: https://nektosact.com/installation/index.html to install act in your OS.
- Pm2 installed. Tip: can be helpfull tool to keep to golang api and job process running forever.

## Tecnologies

- Golang
- Database(sqlite)
- Queue(Redis + asynq)
- Act to run Github Action pipeline

## Instructions

- Clone project
- Create file **.env** file based **.env.example** file.
- Execute command **bash scripts/setupLocally.sh** to setup to create directories: logs and pipelines and sqlite file named **database**
- Execute command **docker-compose up -d** to run redis container. I'm using redis as queue in that project.
- Execute command **go run cmd/api/main.go** to start api at address http://localhost:3000 .
- Execute command **go run cmd/job/main.go** to start job process when is reponsable to consume message from the queue and execute GithubAction pipeline.
- Execute command **bash scripts/deploy.sh** to build the project and execute using pm2.
- Import the file named **insomnia.json** on Insominia to test the endpoints.


## Architecture

![the project architecture](./architecture.png)


## Explanation about envs

```
API_KEY="" // You set random value here, but many endpoint you will need to pass on header the requests. For example: curl --request GET \
  --url http://localhost:3000/triggers \
  --header 'Content-Type: application/json' \
  --header 'User-Agent: insomnia/9.3.3' \
  --header 'x-api-key: API_KEY_VALUE_HERE' \
  --data '{}'
API_BASE_URL="http://localhost:3000" // The address where your api is running
REDIS_URL="127.0.0.1:6379"  // The redis url connection

PHASE_TOKEN_SERVICE=""  // The phase token service will generate, to generate follow the instructions: https://docs.phase.dev/console/apps#service-tokens
PHASE_HOST="https://console.phase.dev" The phase secret manager api endpoint 
PHASE_PROJECT=""       // The project name
PHASE_ENV=Production  // The environment you will use to store the secrets
```

## Extra tips:

### What is Trigger?
​
The trigger is webhook url you will use to setup github to notify the api to run Github action pipeline
​
#### The request body to create trigger
##### The repository is public

```
{
  "actionToRun": "pipeline.yml",
  "linkRepository": "https://github.com/tiago123456789/simulate-github-actions-pipeline"
}
```
​
##### The respository is private
```
{
  "actionToRun": "pipeline.yml",
  "linkRepository": "https://github.com/tiago123456789/simulate-github-actions-pipeline",
  "isPrivate": true,
  "repositoryToken": "personal_access_tokens_classic_github_with_repository_permission"
}
```
​
##### The repository is private and pipeline with secrets
``` 
{
  "actionToRun": "pipeline.yml",
  "linkRepository": "https://github.com/tiago123456789/simulate-github-actions-pipeline",
  "isPrivate": true,
  "repositoryToken": "personal_access_tokens_classic_github_with_repository_permission",
  "envs": {
    "secret_key_name": "secret_value_here",
    "secret_key2_name": "secret_value2_here"
  }
}
```

### How to generate Repository token?

The repository token is a Github token has permission to execute action on your Github account.

To create a repository token following the steps:
- Access your Github account
- Click on your profile
- Click option **Settings**
- Click option **Developer settings**
- Click option **Personal access tokens**
- Click option **Tokens(classic)**
- Click button **Generate new token**
- Fill the inputs required to generate token. WARN: the scope part select option **repo**
- Click button **Generate token**


### How to setup trigger on Github repository

- First, create a trigger with all data required
- Access Github repository
- Click option **Settings**
- Click option **Webhooks**
- Click option **Add webhook**
- Fill input named *Payload URL* with key *webhookUrl* the response when created new trigger. Tip: if you running the application locally try use the tools: **ngrok** or **localtunnel** to allow Github send requests for your local application.
- Fill input named *Content type* for **application/json**
- Fill input named *Secret* with key *secret* the response when created new trigger
- Set when will trigger the URL
- Click the button *add webhook*
- Now you need only execute any action for Github trigger a URL.

### Phase secret manager

- Website: https://phase.dev/
- Version Cloud(free tier): https://console.phase.dev/?ref=navbar