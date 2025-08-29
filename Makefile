-include .env
export

APP       ?= milos-residence
MAIN      ?= ./cmd/web
LDFLAGS   ?= -s -w 
GOFLAGS   ?=
DB        ?= postgres

DB_HOST   ?= localhost
DB_PORT   ?= 5432
DB_USER   ?= app
DB_NAME   ?= appdb
DB_SSLMODE?= disable

MIG       ?= ./migrations

DSN ?= host=$(DB_HOST) port=$(DB_PORT) user=$(DB_USER) password=$(DB_PASSWORD) dbname=$(DB_NAME) sslmode=$(DB_SSLMODE)

GOOSE = GOOSE_DRIVER=$(DB) GOOSE_DBSTRING="$(DSN)" GOOSE_MIGRATION_DIR=$(MIG) goose

.PHONY: help up up1 down down1 redo reset goto to version status create fix run build br dev clean fmt vet tidy test build-linux build-windows build-macos

build:
	go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(APP) $(MAIN)

run: build 
	DB_HOST=$(DB_HOST) DB_PORT=$(DB_PORT) DB_USER=$(DB_USER) DB_NAME=$(DB_NAME) DB_SSLMODE=$(DB_SSLMODE) ./$(APP)

br: run

dev:
	go run $(MAIN)

clean:
	rm -f $(APP) $(APP).exe $(APP)-linux $(APP)-darwin

fmt: ; go fmt ./...

vet: ; go vet ./...

tidy: ; go mod tidy

test: ; go test ./...

build-linux:
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(APP)-linux $(MAIN)

build-windows:
	GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(APP).exe $(MAIN)

build-macos:
	GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(APP)-darwin $(MAIN)

help:
	@echo "Targets:"
	@echo "  up|up1|down|redo|reset|status|version|goto v=NNN|create n=name [t=sql|go]"

up:        ; $(GOOSE) up

up1:       ; $(GOOSE) up-by-one

down:      ; $(GOOSE) down

down1:     ; $(GOOSE) down

redo:      ; $(GOOSE) redo

reset:     ; $(GOOSE) reset

status:    ; $(GOOSE) status

version:   ; $(GOOSE) version

goto:
ifndef v
	$(error Provide version with v=NNN)
endif
	$(GOOSE) goto $(v)

to: goto

create:
ifndef n
	$(error Provide name with n=your_migration_name)
endif
	$(GOOSE) create $(n) $(if $(t),$(t),sql)

fix:
	@echo DB=$(DB)
	@echo DB_HOST=$(DB_HOST) DB_PORT=$(DB_PORT) DB_USER=$(DB_USER) DB_NAME=$(DB_NAME) DB_SSLMODE=$(DB_SSLMODE)
	@echo MIG=$(MIG)
	@which goose || true

