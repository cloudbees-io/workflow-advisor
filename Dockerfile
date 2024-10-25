FROM alpine:latest
RUN  apk upgrade --no-cache --update libcrypto3 libssl3
COPY advisor /usr/local/bin/