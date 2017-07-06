FROM alpine
MAINTAINER Kharkiv Gophers (kostyamol@gmail.com)

EXPOSE 6379 3030 3000 8100 2540

RUN useradd -c 'center-smart-house user' -m -d /home/center-user -s /bin/bash center-user
ENV HOME /home/center-user
USER center-user
COPY ./cmd/center-smart-house $HOME
CMD center-smart-house
