#!/bin/bash

env GOOS=linux GOARCH=arm GOARM=6 go build main.go
scp ./main pi@192.168.88.15:/home/pi/go/src/github.com/DeziderMesko/periph/experimental/main
ssh pi@@192.168.88.15 cd /home/pi/go/src/github.com/DeziderMesko/periph/experimental/; ./main