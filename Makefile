GOPKG_BASE := github.com/raravena80/sshrunner
GOPKGS := $(shell go list ./... | grep -v /vendor/)
GOPKG_COVERS := $(shell go list ./... | grep -v '^$(GOPKG_BASE)/vendor/' | grep -v '^$(GOPKG_BASE)$$' | sed "s|^$(GOPKG_BASE)/|cover/|" | sed 's/$$/.cover/')
COVER_MODE := atomic
VERSION := $(shell git describe --tags --always)
FIRST_GOPATH=$(shell go env GOPATH | cut -d: -f1)
GOFILES = $(shell find . -name '*.go' -not -path './vendor/*')
# Define in your CI system as an env var
COVERALLS_TOKEN ?= undefined

default: build

cover/%.cover: %
	mkdir -p $(dir $@)
	go test -v -coverprofile=$@ -covermode=$(COVER_MODE) ./$<

cover/all: $(GOPKG_COVERS)
	echo mode: $(COVER_MODE) > $@
	for f in $(GOPKG_COVERS); do test -f $$f && sed 1d $$f >> $@ || true; done

goveralls: cover/all
	$(FIRST_GOPATH)/bin/goveralls -coverprofile=cover/all -service=circle-ci \
		-repotoken $(COVERALLS_TOKEN) || echo "not sending to coveralls"

circle: goveralls

workdir:
	mkdir -p workdir

build: workdir/sshrunner

build-native: $(GOFILES)
	go build -o workdir/sshrunner .

workdir/sshrunner: $(GOFILES)
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o workdir/sshrunner.linux.amd64 .

test: test-all

test-all:
	@go test -v $(GOPKGS)

clean:
	rm -rf cover workdir

.PHONY: default test goveralls circle build
