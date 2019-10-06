FROM golang:1.13

WORKDIR /build
COPY . .

RUN go install .

WORKDIR /data

CMD ["iasipbot_mastodon"]
