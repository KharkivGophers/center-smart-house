FROM alpine
MAINTAINER Kharkiv Gophers (kostyamol@gmail.com)

EXPOSE 6379 3030 3000 8100 2540

COPY ./cmd/center-smart-house /home
RUN \
  cd /home && \
  chown daemon center-smart-house && \
  chmod +x center-smart-house
  
USER daemon
CMD center-smart-house
