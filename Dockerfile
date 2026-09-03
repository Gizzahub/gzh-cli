FROM alpine:3.23
COPY gz /usr/bin/gz
ENTRYPOINT ["/usr/bin/gz"]
