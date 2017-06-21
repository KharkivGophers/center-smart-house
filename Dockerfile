FROM golang
MAINTAINER Kharkiv Gophers (kostyamol@gmail.com)

EXPOSE 6379 3030 3000 8100 2540

RUN useradd -c 'center-smart-house user' -m -d /home/center -s /bin/bash center
ENV HOME /home/center
ENV GOPATH $HOME/go

COPY . $HOME/go/src/github.com/KharkivGophers/center-smart-house
WORKDIR $HOME/go/src/github.com/KharkivGophers/center-smart-house

RUN go get ./
RUN go build

CMD rm -r !(center-smart-house)
USER center
#CMD center-smart-house
