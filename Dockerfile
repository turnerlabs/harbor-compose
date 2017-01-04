FROM alpine:latest

RUN apk --update add ca-certificates

ADD dist/ncd_linux_amd64 /usr/local/bin/harbor-compose

RUN chmod +x /usr/local/bin/harbor-compose

WORKDIR /work

ENTRYPOINT ["harbor-compose"]
CMD ["--help"]