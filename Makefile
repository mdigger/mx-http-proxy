APPNAME ?= $(shell basename ${PWD})
DATE	:= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT		:= $(shell git rev-parse --short HEAD 2>/dev/null)
TAG		:= $(shell git describe --tag --long --dirty 2>/dev/null)
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
debug: info
	go mod tidy
	go build -race -tags dev $(FLAGS) -o $(APPNAME) ./mxhttp

.PHONY: docker
docker: info
	docker build -t $(APPNAME) \
	--build-arg VERSION=$(TAG) \
	--build-arg COMMIT=$(GIT) \
	--build-arg DATE=$(DATE) \
	.
	docker run -p 8000:8000 -e MX=$(MX) $(APPNAME)

.PHONY: run
run:
	docker run -p 8000:8000 -e MX=$(MX) $(APPNAME) -log all,color