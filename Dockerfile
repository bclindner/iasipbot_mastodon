FROM golang:1.12

WORKDIR /build
COPY . .

ENV GO111MODULES=on
RUN go build .

FROM alpine:latest
WORKDIR /srv
COPY --from=0 /build/iasipbot_mastodon .

CMD ['./iasipbot_mastodon']
