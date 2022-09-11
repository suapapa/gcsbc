# builder
FROM golang:alpine AS builder

# RUN apk --update --no-cache add git fuse-dev

## build gcsfuse v0.41.6 is latest release at point of writing (2022-09-11)
ARG GCSFUSE_VERSION=v0.41.6

RUN go install github.com/googlecloudplatform/gcsfuse@${GCSFUSE_VERSION}

## build app
WORKDIR /build
COPY . /build
RUN go build -o /build/app
# RUN strip /build/app
# RUN upx -q -9 /build/app

# image
FROM alpine:latest

RUN apk add --update --no-cache fuse

## install gcsfuse
COPY --from=builder /go/bin/gcsfuse /usr/bin

## install all
COPY --from=builder /build/app .

RUN mkdir /bucket

EXPOSE 8080

ENTRYPOINT ["./app"]
CMD ["-r", "/bucket/"]
