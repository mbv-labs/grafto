set dotenv-load

# alias
alias r := run
alias rw := run-worker

alias wc := watch-css

alias sm := serve-mails

alias mm := make-migration
alias um := up-migrations
alias dm := down-migrations
alias dmt := down-migrations-to
alias rdb := reset-db

alias gdf := generate-db-functions

alias ct := compile-templates
alias ft := fmt-templates

alias rm := river-migrate-up

alias ex := explore

default:
    @just --list

# CSS
watch-css:
    @cd resources && npm run watch-css

# Mails
serve-mails:
    @cd ./pkg/mail/templates && wgo -file=.go -file=.templ -xfile=_templ.go templ generate :: go run ./server/main.go

# Database 
make-migration name:
	@goose -dir migrations $DB_KIND $DATABASE_URL create {{name}} sql

up-migrations:
	@goose -dir migrations $DB_KIND $DATABASE_URL up

down-migrations:
	@goose -dir migrations $DB_KIND $DATABASE_URL down

down-migrations-to version:
	@goose -dir migrations $DB_KIND $DATABASE_URL down-to {{version}}

reset-db:
	@goose -dir migrations $DB_KIND $DATABASE_URL reset

generate-db-functions:
	sqlc compile && sqlc generate

# Application
run:
    air -c .air.toml

# Worker
run-worker:
    @go run ./cmd/worker/main.go

# templates
compile-templates:
    templ generate 

fmt-templates:
    templ fmt ./views/ . 

# river
river-migrate-up:
	river migrate-up --database-url $QUEUE_DATABASE_URL

# exploration
explore:
    @go run ./cmd/explore/main.go
