#!/bin/bash
function build_image(){
    echo "Building $1..."
    docker rmi -f webchat:$1
    rm -f webchat
    go build -v .
    docker build --rm -t go-webchat:$1 -f docker/Dockerfile .
}

build_image "latest"
