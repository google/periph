# Copyright 2016 The PIO Authors. All rights reserved.
# Use of this source code is governed under the Apache License, Version 2.0
# that can be found in the LICENSE file.

# This Makefile captures common tasks for the pio library. The hope is that this Makefile can remain
# simple and straightforward...

# *** This Makefile is a work in progress, please help impove it! ***

all:

# test runs the platform independent tests
test:
	go test ./...
	@if gofmt -l . | grep .go; then \
	  echo "Repo contains improperly formatted go files; run gofmt on above files" && exit 1; \
	else echo "OK gofmt"; fi
	-go vet ./...

# cross generates binaries for all executables in cmd for "all" platforms. This allows these
# binaries to be published and it also verifies that they can be built on all platforms
cross:
	@cd cmd; \
	set +x; \
	for i in *; do ( \
	  test -d $$i || continue; \
	  cd $$i; \
	  GOARCH=arm GOOS=linux go build -o $$i-arm .; \
	  GOARCH=arm64 GOOS=linux go build -o $$i-arm64 .; \
	  GOARCH=amd64 GOOS=linux go build -o $$i-amd64 .; \
	  GOARCH=amd64 GOOS=windows go build -o $$i-win64 .; \
	); done
	
# clean removes all compiled binaries, especially from the cross target
clean:
	bash -c "rm cmd/*/*-{arm,arm64,amd64,win64}"
