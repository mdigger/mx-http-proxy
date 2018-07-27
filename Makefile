.PHONY: debug
DATE    ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT     ?= $(shell git rev-parse --short HEAD 2>/dev/null)
FLAGS   := -ldflags "-X main.commit=$(GIT) -X main.date=$(DATE)"
APPNAME ?=$(shell basename ${PWD})

debug:
	go build -race -tags dev $(FLAGS) -o $(APPNAME)
	LOG=DEV,ALL ./$(APPNAME)
