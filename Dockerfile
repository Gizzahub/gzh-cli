FROM alpine:3.20
COPY gzh-manager /usr/bin/gzh-manager
ENTRYPOINT ["/usr/bin/gzh-manager"]
