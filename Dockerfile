FROM golang:1.13

WORKDIR /build
COPY . .

ENV GO111MODULES=on
RUN go install .

WORKDIR /srv

ENTRYPOINT ["iasipbot_mastodon"]
