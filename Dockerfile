FROM alpine:latest

ADD dist/ncd_linux_386 /usr/local/bin/harbor-compose
RUN chmod +x /usr/local/bin/harbor-compose

WORKDIR /work

ENTRYPOINT ["/usr/local/bin/harbor-compose"]
CMD ["--help"]