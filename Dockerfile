FROM golang:1.12.5 AS build

RUN curl -fsSL -o /usr/local/bin/dep https://github.com/golang/dep/releases/download/v0.5.0/dep-linux-amd64
RUN chmod +x /usr/local/bin/dep

WORKDIR /go/src/github.com/lob/pharos

COPY Gopkg.toml Gopkg.toml
COPY Gopkg.lock Gopkg.lock
RUN dep ensure -vendor-only

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-w -s" -o ./bin/pharos-api-server ./cmd/pharos-api-server
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-w -s" -o ./bin/migrations ./cmd/migrations/*.go

FROM alpine:3.8

RUN apk update \
  && apk upgrade \
  && apk add --no-cache ca-certificates \
  && update-ca-certificates

# Add AWS RDS CA bundle and split the bundle into individual certs (prefixed with cert)
# See http://blog.swwomm.com/2015/02/importing-new-rds-ca-certificate-into.html
ADD https://s3.amazonaws.com/rds-downloads/rds-combined-ca-bundle.pem /tmp/rds-ca/aws-rds-ca-bundle.pem
RUN cd /tmp/rds-ca && awk '/-BEGIN CERTIFICATE-/{close(x); x=++i;}{print > "cert"x;}' ./aws-rds-ca-bundle.pem \
    && for CERT in /tmp/rds-ca/cert*; do mv $CERT /usr/local/share/ca-certificates/aws-rds-ca-$(basename $CERT).crt; done \
    && rm -rf /tmp/rds-ca \
    && update-ca-certificates

COPY --from=build /go/src/github.com/lob/pharos/bin .

CMD ["./pharos-api-server"]
