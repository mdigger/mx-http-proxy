.PHONY: debug build sertificates
DATE    ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT     ?= $(shell git rev-parse --short HEAD 2>/dev/null)
FLAGS   := -ldflags "-X main.commit=$(GIT) -X main.date=$(DATE)"
APPNAME ?=$(shell basename ${PWD})

debug:
	go build -race -tags dev $(FLAGS) -o $(APPNAME)
	LOG=DEV,ALL ./$(APPNAME)

build:
	go generate
	go build -race $(FLAGS) -o $(APPNAME)
	LOG=COLOR,ALL ./$(APPNAME)

sertificates:
	openssl req -x509 -out localhost.crt -keyout localhost.key \
	-newkey rsa:2048 -nodes -sha256 \
	-subj '/CN=localhost' -extensions EXT -config <( \
	printf "[dn]\nCN=localhost\n[req]\ndistinguished_name = dn\n[EXT]\nsubjectAltName=DNS:localhost\nkeyUsage=digitalSignature\nextendedKeyUsage=serverAuth")

