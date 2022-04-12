FROM golang:1.18 AS differ

WORKDIR /usr/src/app

COPY go.mod go.mod
COPY go.sum go.sum

RUN go mod download

COPY . .

RUN go build -o /config-diff ./cmd/diff/main.go

FROM meisnate12/plex-meta-manager:nightly

RUN apt-get update && apt-get install git gettext-base -y

COPY entrypoint.sh .

RUN chmod +x entrypoint.sh

COPY --from=differ /config-diff /config-diff

ENTRYPOINT ["/entrypoint.sh"]
