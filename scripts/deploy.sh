#!/usr/bin/env sh

rm_container(){
    if docker ps -f name='$1' &> /dev/null; then
        echo "Removing old container: $(docker rm -f $1)"
    fi
}

build_image(){
    echo "Building $1..."
    docker rmi -f webchat:$1
    rm -f webchat
    docker build --rm -t webchat:$1 -f docker/Dockerfile .
}

rm_container "webchat"
build_image "latest"

DATA_DIR="/home/autobd-container/data/server-data/"
DOCS_DIR="/home/$USER/go-webchat/docs/"
STATIC_DIR="/home/$USER/go-webchat/static/"
SECRET_DIR="/home/$USER/secret/"
ETC_DIR="/home/$USER/go-webchat/etc/"
PORT=80

echo "Running server: $(docker run -d \
    -p $PORT:80 \
    -p 443:443 \
    -v $DATA_DIR:/home/webchat/data/ \
    -v $SECRET_DIR:/home/webchat/secret/ \
    -v $DOCS_DIR:/home/webchat/docs/ \
    -v $STATIC_DIR:/home/webchat/static/ \
    -v $ETC_DIR:/home/webchat/etc/ \
    --name webchat webchat:latest)"
docker logs webchat

