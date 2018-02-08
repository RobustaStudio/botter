FROM golang:alpine

RUN apk update && apk add git

RUN go get github.com/RobustaStudio/botter

ENTRYPOINT ["botter"]

WORKDIR /root/
