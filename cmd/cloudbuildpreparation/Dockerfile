# build stage
FROM golang:1.12.10-alpine AS build-env
# FROM golang:alpine AS build-env
RUN apk --no-cache add build-base git bzr mercurial gcc
ADD . /src
# RUN cd /src && go mod download
RUN cd /src && go build -o goapp

# final stage
FROM alpine
RUN apk upgrade --update-cache \
    && apk add ca-certificates \
    && rm -rf /var/cache/apk/*
WORKDIR /app
COPY --from=build-env /src/goapp /app/
ENTRYPOINT ["./goapp"]
