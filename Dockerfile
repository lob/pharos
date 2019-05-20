FROM golang:1.12.5 AS build

RUN curl -fsSL -o /usr/local/bin/dep https://github.com/golang/dep/releases/download/v0.5.0/dep-linux-amd64
RUN chmod +x /usr/local/bin/dep

WORKDIR /go/src/github.com/lob/pharos

COPY Gopkg.toml Gopkg.toml
COPY Gopkg.lock Gopkg.lock
RUN dep ensure -vendor-only

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-w -s" -o ./bin/pharos-api-server ./cmd/pharos-api-server

FROM alpine:3.8

RUN apk --no-cache add ca-certificates

COPY --from=build /go/src/github.com/lob/pharos/bin/pharos-api-server .

CMD ["./pharos-api-server"]
