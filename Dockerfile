FROM golang:1.13

WORKDIR /build
COPY . .

RUN go install .

WORKDIR /data

ENTRYPOINT ["iasipbot_mastodon"]
