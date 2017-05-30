FROM alpine:3.3

ADD app /bin/

ENTRYPOINT "/bin/app"

EXPOSE 8000
