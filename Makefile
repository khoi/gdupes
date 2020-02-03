PKGS := $(shell go list ./... | grep -v /vendor)

.PHONY: test install

all: test install

test:
	go test -race -v $(PKGS)

install:
	go install
