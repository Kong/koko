FROM golang:1.17.0 AS build

WORKDIR /koko

COPY go.mod ./
COPY go.sum ./
RUN go mod download
RUN go mod verify

ADD . .
# TODO(hbagdi) pass along commit hash and tag details
RUN CGO_ENABLED=1 go build \
  -ldflags="-extldflags=-static" \
  -tags sqlite_omit_load_extension,netgo,osusergo \
  -o koko \
  main.go

FROM alpine:3.15.0
RUN adduser --disabled-password --gecos "" koko
RUN apk --no-cache add ca-certificates bash
USER koko
COPY --from=build /koko/koko /usr/local/bin
ENTRYPOINT ["koko"]
