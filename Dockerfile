FROM golang
MAINTAINER Kharkiv Gophers (kostyamol@gmail.com)

EXPOSE 6379 3030 3000 8100 2540

COPY . /go/src/github.com/KharkivGophers/center-smart-house
WORKDIR /go/src/github.com/KharkivGophers/center-smart-house

RUN go get ./
RUN go build
CMD center-smart-house
