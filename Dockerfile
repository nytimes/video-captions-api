FROM alpine:3.3

RUN apk -qq update && apk -qq add --no-cache ca-certificates
# TODO: use gizmo LoadConfigFromEnv instead of the json file
ADD config.json captions-api /bin/

ENTRYPOINT "/bin/captions-api"

EXPOSE 8000
