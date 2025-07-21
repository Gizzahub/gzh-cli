FROM alpine:3.22
COPY gzh-manager /usr/bin/gzh-manager
ENTRYPOINT ["/usr/bin/gzh-manager"]