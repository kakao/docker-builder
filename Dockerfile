FROM docker:18.03-dind
MAINTAINER cloud deploy <ccc@kakaocorp.com>

RUN apk add --no-cache \
    git \
    openssh-client \
    bash

RUN mkdir -p /usr/src/app
COPY bin /usr/src/app/

WORKDIR /usr/src/app

CMD ./docker-builder
