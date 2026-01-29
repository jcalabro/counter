FROM golang:1.25.6

ENV GODEBUG netdns=cgo

ADD . /go/src/github.com/jcalabro/counter.git
WORKDIR /go/src/github.com/jcalabro/counter.git
RUN go install ./...

CMD ["counterd"]
