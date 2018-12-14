FROM golang:1.11.3-alpine AS builder

ENV CGO_ENABLED 0

RUN     apk add --no-cache git
ADD     . /code
WORKDIR /code
RUN     go build -o /bin/captions-api

FROM alpine:3.8

RUN  apk add --no-cache ca-certificates
COPY --from=builder /bin/captions-api /bin/captions-api

ENTRYPOINT ["/bin/captions-api"]
