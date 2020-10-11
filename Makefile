.EXPORT_ALL_VARIABLES:

COMPOSE_PROJECT_NAME=kardia-explorer
DOCKER_FILE=.Dockerfile
# Variable for filename for store running processes id
PID_FILE = /tmp/explorer_backend.pid
# We can use such syntax to get main.go and other root Go files.
GO_FILES = $(wildcard *.go)

all: env build
env:
	if test ! -f .env ; \
    then \
         cp .env.sample .env ; \
    fi;
env_dev:
	cp .env.sample ./features/.env ; \
	cp .env.sample ./cmd/api/.env ; \
	cp .env.sample ./cmd/grabber/.env ; \
	cp .env.sample ./server/db/.env ;
build:
	docker-compose build
run-grabber:
	docker-compose up grabber
run-backend:
	docker-compose up backend
utest:
	go test ./... -cover -covermode=count -coverprofile=cover.out -coverpkg=./internal/...
	go tool cover -func=cover.out
list-service:
	docker-compose ps
exec-service:
	docker-compose exec $(service) bash
logs:
	docker-compose logs -f $(service)
destroy:
	docker-compose down
deploy:
	docker-compose up -d
start:
	go run $(GO_FILES) & echo $$! > $(PID_FILES)
stop:
	-kill `pstree -p \`cat $(PID)\` | tr "\n" " " |sed "s/[^0-9]/ /g" |sed "s/\s\s*/ /g"`
restart: stop start
	@echo "STARTED my-app" && printf '%*s\n' "40" '' | tr ' ' -
serve-be: start
	fswatch -or --event=Updated /go/src/github/kardiachain/explorer-backend | \
	xargs -n1 -I {} make restart
