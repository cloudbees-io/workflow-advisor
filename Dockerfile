FROM alpine:latest
RUN <<EOF
    apk update
    apk upgrade \
        libcrypto3 \
        libssl3
    apk cache clean
EOF
COPY advisor /usr/local/bin/