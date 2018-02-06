FROM golang:1.9-alpine AS builder

ENV CGO_ENABLED 0
ENV PROJ github.com/NYTimes/video-captions-api

ADD . /go/src/$PROJ
RUN go test $PROJ/...
RUN go install $PROJ

FROM alpine:3.7

RUN  apk add --no-cache ca-certificates
COPY --from=builder /go/bin/video-captions-api /bin/captions-api

ENTRYPOINT ["/bin/captions-api"]
