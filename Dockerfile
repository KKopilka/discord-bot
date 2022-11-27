FROM golang:1.19-alpine as build-env
# Enable go mod environment for building dependencies
ENV GO111MODULE=on \
    CGO_ENABLED=1 \
    GOOS=linux \
    GOARCH=amd64 \
    GOSUMDB=off
# install apps
RUN apk add --update --no-cache ca-certificates git tzdata build-base openssh-client\
    && apk add --update --no-cache --repository http://dl-3.alpinelinux.org/alpine/edge/community \
    --repository http://dl-3.alpinelinux.org/alpine/edge/main

# setup working directory
ARG WORKDIR=/discord-bot
WORKDIR ${WORKDIR}
