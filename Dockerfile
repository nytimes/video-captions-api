FROM alpine:3.3

ADD captions-api /bin/

ENTRYPOINT "/bin/captions-api"

EXPOSE 8000
