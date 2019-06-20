# This Dockerfile is intended to be used in the CI environment and depends on
# an existing binary built outside Docker.

FROM alpine:3.10

RUN apk add --no-cache ca-certificates
ADD video-captions-api /bin/captions-api

ENTRYPOINT ["/bin/captions-api"]
