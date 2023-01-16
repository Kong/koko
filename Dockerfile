FROM rust:1.66 AS atc-router

WORKDIR /koko

ADD . .

RUN ./scripts/build-library.sh go-atc-router/main/make-lib.sh /koko/lib

FROM golang:1.19 AS build

WORKDIR /koko

ARG GIT_COMMIT_HASH
ARG GIT_TAG
COPY go.mod ./
COPY go.sum ./
RUN go mod download

ADD . .

COPY --from=atc-router /koko/lib/*.a /tmp/lib/

RUN CGO_ENABLED=1 go build \
  -ldflags="-extldflags=-static -X github.com/kong/koko/internal/info.VERSION=$GIT_TAG -X github.com/kong/koko/internal/info.COMMIT=$GIT_COMMIT_HASH" \
  -tags sqlite_omit_load_extension,netgo,osusergo \
  -o koko \
  main.go

FROM alpine:3.17.1
RUN adduser --disabled-password --gecos "" koko
RUN apk --no-cache add ca-certificates bash
USER koko
COPY --from=build /koko/koko /usr/local/bin
ENTRYPOINT ["koko"]
