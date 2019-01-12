#!/bin/bash

GOOS=linux GOARCH=arm GOARM=6 go build main.go
scp ./main pi@192.168.88.15:/home/pi/go/src/github.com/DeziderMesko/periph/experimental/cmd/ccs811/main
ssh -t pi@192.168.88.15 '/home/pi/go/src/github.com/DeziderMesko/periph/experimental/cmd/ccs811/main' status
#ssh -t pi@192.168.88.15 'cd /home/pi/go/src/github.com/DeziderMesko/periph/experimental/cmd/ccs811/; source ~/.profile; /usr/local/go/bin/go build main.go; ./main'