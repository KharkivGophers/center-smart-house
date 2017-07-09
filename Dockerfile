FROM golang:alpine
MAINTAINER Kharkiv Gophers (kostyamol@gmail.com)

EXPOSE 6379 3030 3000 8100 2540

WORKDIR /home
COPY ./cmd/center-smart-house .

RUN \  
 chown daemon center-smart-house && \
 chmod +x center-smart-house
  
USER daemon
ENTRYPOINT ./center-smart-house
