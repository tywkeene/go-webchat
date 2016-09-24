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
    go build -v .
    docker build --rm -t webchat:$1 -f docker/Dockerfile .
}

rm_container "webchat"
build_image "latest"

DATA_DIR="/home/$USER/srv/webchat/"
DOCS_DIR="/home/$USER/srv/webchat/docs/"
STATIC_DIR="/home/$USER/srv/webchat/static/"
SECRET_DIR="/home/$USER/secret/"
PORT=80

mkdir -p $DATA_DIR
echo "Running server: $(docker run -d \
    -p $PORT:80 \
    -p 443:443 \
    -v $DATA_DIR:/home/webchat/data/ \
    -v $SECRET_DIR:/home/webchat/secret/ \
    -v $DOCS_DIR:/home/webchat/docs/ \
    -v $STATIC_DIR:/home/webchat/static/ \
    --name webchat webchat:latest)"
docker logs webchat
