FROM alpine:3.3

# TODO: use gizmo LoadConfigFromEnv instead of the json file
ADD config.json captions-api /bin/

ENTRYPOINT "/bin/captions-api"

EXPOSE 8000
