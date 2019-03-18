FROM golang:1.12.1-stretch

ADD . /go/src/github.com/kirugan/aviasales

RUN go get github.com/pkg/errors github.com/bradfitz/gomemcache/memcache
RUN go install github.com/kirugan/aviasales

ENTRYPOINT /go/bin/aviasales

EXPOSE 8080

