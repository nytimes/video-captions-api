FROM golang:1.15-alpine3.14 as builder
ADD . /build
WORKDIR /build
RUN CGO_ENABLED=0 GOOS=linux go build -a -o video-captions-api .

FROM alpine:3.14
COPY --from=builder /build/video-captions-api .
RUN addgroup service && adduser -DH -G service service
USER service

ENTRYPOINT [ "./video-captions-api" ]
