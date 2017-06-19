FROM alpine:3.3

RUN apk -qq update && apk -qq add --no-cache ca-certificates
ADD /tmp/gcloud.json captions-api /bin/

ENTRYPOINT "/bin/captions-api"

EXPOSE 8000
