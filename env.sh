#!/bin/sh

export LINE_BOT_SECRET=
export LINE_BOT_TOKEN=
export LINE_LOGIN_ID=
export LINE_LOGIN_SECRET=
export LINE_LOGIN_REDIRECT_URL=https://choo-todo-bot.serveo.net/auth
export EDIT_URL=https://choo-todo-bot.serveo.net/
export MYSQL_USER=todo_user
export MYSQL_PASSWORD=todo_pass
export MYSQL_DATABASE=todo_db
export DATA_SOURCE_NAME="$MYSQL_USER:$MYSQL_PASSWORD@tcp(mysql:3306)/$MYSQL_DATABASE?parseTime=true"

export HEROKU_APP=
export PROD_LINE_LOGIN_REDIRECT_URL=
export PROD_EDIT_URL=
export PROD_DATA_SOURCE_NAME=""