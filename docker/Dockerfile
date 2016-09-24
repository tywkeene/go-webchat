FROM google/golang

MAINTAINER Tyrell Keene <tyrell.wkeene@gmail.com>

RUN useradd -ms /bin/bash webchat
USER webchat

WORKDIR /home/webchat

ADD ./webchat ./webchat
ADD ./etc/config.toml ./config.toml

RUN mkdir /home/webchat/secret
VOLUME /home/webchat/secret

RUN mkdir /home/webchat/docs
VOLUME /home/webchat/docs

RUN mkdir /home/webchat/static
VOLUME /home/webchat/static

RUN mkdir /home/webchat/data
VOLUME /home/webchat/data

EXPOSE 80 443

USER root

HEALTHCHECK CMD curl -A "Docker-Health-Check" --fail -k "https://0.0.0.0/" || exit 1

ENTRYPOINT ./webchat -config="/home/webchat/config.toml"
