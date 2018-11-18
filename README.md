# choo-todo-bot

## Note
- Repo: https://github.com/choobot/choo-todo-bot/

## Live Testing
- The webhook URL for LINE Messaging API will be https://choo-todo-bot.herokuapp.com/callback
- Bot ID is [@gpd2291p](http://line.me/ti/p/~@gpd2291p)

## Prerequisites for Development
- Mac or Linux which can run shell script
- Docker
- Heroku CLI (for deployment only)

## Local Running and Expose to the internet
- Copy the project to $GOPATH/src/github.com/choobot/choo-todo-bot/
- Config environment variables in env.sh
- $ ./run.sh
- The webhook URL for LINE Messaging API will be https://choo-todo-bot.serveo.net/callback
- Config webhook URL for LINE Messaging API

## Unit Testing
- Config environment variables in env.sh
- $ ./test.sh

## Deployment
- Config environment variables in env.sh
- Config webhook URL for LINE Messaging API
- $ ./deploy.sh

## Tech Stack
- Go
- Echo
- Angular
- Bootstrap
- MySQL
- LINE Messaging API
- Docker
- Heroku
- Node.js, Karma, Jasmine, PhantomJS (for Front-End Unit Testing)