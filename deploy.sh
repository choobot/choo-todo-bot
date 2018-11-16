#!/bin/sh

heroku container:login
heroku container:push web --app=choo-todo-bot
heroku container:release web --app=choo-todo-bot