#!/bin/sh

source env.sh
docker-compose up --build -d
docker-compose exec go "/bin/sh" "-c" "go test ./... -coverprofile=coverage.out && go tool cover -func=coverage.out"
docker-compose stop

docker-compose -f node-docker-compose.yml up
docker-compose -f node-docker-compose.yml stop