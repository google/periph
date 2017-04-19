#!/bin/bash
# Copyright 2016 The Periph Authors. All rights reserved.
# Use of this source code is governed under the Apache License, Version 2.0
# that can be found in the LICENSE file.

# Cross compiles an executable to ARM and pushes it to another host.

set -eu
cd "$(dirname $0)"

if [ "$#" -le 1 ]; then
  echo "usage: $0 <hostname> <tool names...>"
  exit 1
fi

HOST="$1"
shift

for i; do
  NAME="${i%/}"
  echo "- Building: $NAME"
  cd "./$NAME"
  GOOS=linux GOARCH=arm go test -i .
  time GOOS=linux GOARCH=arm go build .
  echo ""
  echo -n "- Copying:  "
  time rsync -v --compress "$NAME" "$HOST:${NAME}"
  echo ""
  rm "$NAME"
  cd ..
done
