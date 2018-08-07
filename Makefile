APPNAME ?= $(shell basename ${PWD})
DATE	:= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT		:= $(shell git rev-parse --short HEAD 2>/dev/null)
TAG		:= $(shell git describe --tag 2>/dev/null)
FLAGS   := -ldflags "-X main.version=$(TAG) -X main.commit=$(GIT) -X main.buildDate=$(DATE)"
MX      = 631hc.connector73.net

.PHONY: info
info:
	@echo "────────────────────────────────"
	@echo "Go:       $(subst go version ,,$(shell go version))"
	@echo "Date:     $(DATE)"
	@echo "Git:      $(GIT)"
	@echo "Version:  $(TAG)"
	@echo "────────────────────────────────"

.PHONY: debug
debug: build-debug run

.PHONY: release
release: build-release run

.PHONY: build-debug
build-debug: info
	go build -i -race -tags dev $(FLAGS) -o $(APPNAME) 

.PHONY: build-release
build-release: info
	go get -d -v github.com/shurcooL/vfsgen github.com/shurcooL/httpfs/filter
	go generate ./...
	go build -i -race $(FLAGS) -o $(APPNAME)

.PHONY: run
run:
	./$(APPNAME) -log all,color -mx $(MX) -port localhost:8000

.PHONY: docker
docker: | docker-build docker-run

.PHONY: docker-build
docker-build: info
	docker build -t $(APPNAME) \
	--build-arg VERSION=$(TAG) \
	--build-arg COMMIT=$(GIT) \
	--build-arg DATE=$(DATE) \
	.

.PHONY: docker-run
docker-run:
	docker run -p 8000:8000 -e MX=$(MX) $(APPNAME) -log all,color

.PHONY: sertificates
sertificates:
	openssl req -x509 -out localhost.crt -keyout localhost.key \
	-newkey rsa:2048 -nodes -sha256 \
	-subj '/CN=localhost' -extensions EXT -config <( \
	printf "[dn]\nCN=localhost\n[req]\ndistinguished_name = dn\n[EXT]\nsubjectAltName=DNS:localhost\nkeyUsage=digitalSignature\nextendedKeyUsage=serverAuth")
