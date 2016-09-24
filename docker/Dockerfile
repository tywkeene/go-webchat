FROM google/golang

MAINTAINER Tyrell Keene <tyrell.wkeene@gmail.com>

RUN useradd -ms /bin/bash webchat
USER webchat

WORKDIR /home/webchat

ENV GOPATH=/home/webchat/go

RUN go get github.com/tywkeene/go-webchat

WORKDIR $GOPATH/src/github.com/tywkeene/go-webchat/
RUN go build -v .

VOLUME /home/webchat/secret
VOLUME /home/webchat/etc
VOLUME /home/webchat/docs
VOLUME /home/webchat/static
VOLUME /home/webchat/data

EXPOSE 80 443

USER root

HEALTHCHECK CMD curl -A "Docker-Health-Check" --fail -k "https://0.0.0.0/" || exit 1

WORKDIR $GOPATH/src/github.com/tywkeene/go-webchat/
ENTRYPOINT ./go-webchat -config="/home/webchat/etc/config.toml"

