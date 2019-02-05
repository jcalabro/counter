FROM golang:1.11.2

ADD . /go/src/github.com/jcalabro/counter.git
WORKDIR /go/src/github.com/jcalabro/counter.git
RUN go install ./...

CMD ["/go/bin/counterd"]
