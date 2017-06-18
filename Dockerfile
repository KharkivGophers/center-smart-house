FROM golang
MAINTAINER Kharkiv Gophers (kostyamol@gmail.com)

COPY . /go/src/github.com/KharkivGophers/center-smart-house
WORKDIR /go/src/github.com/KharkivGophers/center-smart-house

RUN go get ./
RUN go build
#RUN center-smart-house

#redis conn
EXPOSE 6379

#tcp conn for data from device
EXPOSE 3030

#tcp conn for config from device
EXPOSE 3000

#http conn with browser
EXPOSE 8100

#web-socket conn with browser for streaming
EXPOSE 2540