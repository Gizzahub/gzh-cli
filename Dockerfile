FROM alpine:3.23
COPY gzh-manager /usr/bin/gzh-manager
ENTRYPOINT ["/usr/bin/gzh-manager"]
