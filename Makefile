
PACKAGE  = sshrunner
BASE     = $(GOPATH)/src/github.com/raravena80/$(PACKAGE)

.PHONY: all
all: | $(BASE)
	cd $(BASE) && $(GO) build -o $(GOPATH)/bin/$(PACKAGE) main.go
