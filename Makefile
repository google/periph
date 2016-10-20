# Copyright 2016 The PIO Authors. All rights reserved.
# Use of this source code is governed under the Apache License, Version 2.0
# that can be found in the LICENSE file.

# This Makefile captures common tasks for the pio library. The hope is that this Makefile can remain
# simple and straightforward...

# *** This Makefile is a work in progress, please help impove it! ***

# not sure yet what all should do...
all:
	@echo Available targets: test build

.PHONY: all test clean

# test runs the platform independent tests
# (gofmt|grep is used to obtain a non-zero exit status if the formatting is off)
test:
	go test -i ./...
	go test ./...
	@if gofmt -l . | grep .go; then \
	  echo "Repo contains improperly formatted go files; run gofmt on above files" && exit 1; \
	else echo "OK gofmt"; fi
	-go vet ./...

# BUILD
#
# The build target cross compiles each program in cmd to a binary for each platform in the bin
# directory. It is assumed that each command has a main.go file in its directory. Trying to keep all
# this relatively simple and not descend into makefile hell...

# Get a list of all main.go in cmd subdirs
# MAINS becomes: cmd/pio-setup/main.go cmd/pio-info/main.go ...
MAINS := $(wildcard cmd/*/main.go)
# Get a list of all the commands, i.e. names of dirs that contain a main.go
# CMDS becomes: pio-setup pio-info ...
CMDS  := $(patsubst cmd/%/main.go,%,$(MAINS))
# Get a list of binaries to build
# BINS becomes: bin/pio-setup-arm bin/pio-info-arm ... bin/pio-setup-arm64 bin/pio-info-arm64 ...
ARCHS := arm arm64 amd64 win64.exe
BINS=$(foreach arch,$(ARCHS),$(foreach cmd,$(CMDS),bin/$(cmd)-$(arch)))

build: bin $(BINS)
bin:
	mkdir bin

# Rules to build binaries for a command in cmd. The prereqs could be improved...
bin/%-arm: cmd/%/*.go
	GOARCH=arm GOOS=linux go build -o $@ ./cmd/$*
bin/%-arm64: cmd/%/*.go
	GOARCH=arm64 GOOS=linux go build -o $@ ./cmd/$*
bin/%-amd64: cmd/%/*.go
	GOARCH=amd64 GOOS=linux go build -o $@ ./cmd/$*
bin/%-win64.exe: cmd/%/*.go
	GOARCH=amd64 GOOS=windows go build -o $@ ./cmd/$*
	
# clean removes all compiled binaries
clean:
	rm bin/*-*
	rmdir bin
