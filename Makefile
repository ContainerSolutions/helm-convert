HELM_HOME ?= $(shell helm home)
HELM_PLUGIN_DIR ?= $(HELM_HOME)/plugins/helm-convert
VERSION := $(shell sed -n -e 's/version:[ "]*\([^"]*\).*/\1/p' plugin.yaml)
DIST := $(CURDIR)/_dist
LDFLAGS := "-X main.version=${VERSION}"
BINARY := "helm-convert"
DOCKER_IMAGE ?= containersol/helm-convert
DOCKER_TAG ?= latest

# go mod ftw
unexport GOPATH

.PHONY: install
install: build
	cp $(BINARY) $(HELM_PLUGIN_DIR)
	cp plugin.yaml $(HELM_PLUGIN_DIR)

.PHONY: build
build:
	go build -o $(BINARY) -ldflags $(LDFLAGS) main.go

.PHONY: build.docker
build.docker:
	docker build \
		--build-arg LDFLAGS=$(LDFLAGS) \
		--cache-from ${DOCKER_IMAGE} \
		-t ${DOCKER_IMAGE}:$(DOCKER_TAG) .

.PHONY: dist
dist:
	mkdir -p $(DIST)
	GOOS=linux GOARCH=amd64 go build -o $(BINARY) -ldflags $(LDFLAGS) main.go
	tar -zcvf $(DIST)/helm-convert$(VERSION).tar.gz $(BINARY) README.md LICENSE plugin.yaml
	GOOS=darwin GOARCH=amd64 go build -o $(BINARY) -ldflags $(LDFLAGS) main.go
	tar -zcvf $(DIST)/helm-convert$(VERSION).tar.gz $(BINARY) README.md LICENSE plugin.yaml
	GOOS=windows GOARCH=amd64 go build -o $(BINARY).exe -ldflags $(LDFLAGS) main.go
	tar -zcvf $(DIST)/helm-convert$(VERSION).tar.gz $(BINARY).exe README.md LICENSE plugin.yaml

.PHONY: test-all
test-all: vet lint test

.PHONY: test
test:
	go test -v -parallel=4 ./...

.PHONY: lint
lint:
	@go get -u golang.org/x/lint/golint
	go list ./... | xargs -n1 $${HOME}/go/bin/golint

.PHONY: vet
vet:
	go vet ./...
