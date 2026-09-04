FROM alpine:3.23

# dockers_v2 stages the build context per platform (linux/<arch>/gz), not flat
# at the root the way the old `dockers` block did. TARGETOS/TARGETARCH are
# supplied by buildx for each entry in `platforms:`; without this the build
# fails with `"/gz": not found` even though the binary is right there.
ARG TARGETOS
ARG TARGETARCH

COPY ${TARGETOS}/${TARGETARCH}/gz /usr/bin/gz
ENTRYPOINT ["/usr/bin/gz"]
