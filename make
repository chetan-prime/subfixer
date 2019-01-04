#!/bin/sh

# change these to fit your name/e-mail address
export DEBEMAIL="projects@inkus.net"
export DEBFULLNAME="Chetan Chauhan"

export GOPATH=$(pwd)/packages

GOBIN=go
GOFLAGS="-ldflags -w"

${GOBIN} get github.com/pkg/errors

${GOBIN} build ${GOFLAGS} -o subfixer -i *.go || exit

