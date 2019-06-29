FROM golang:1.12.5

ADD . /go/src/github.com/jcalabro/counter.git
WORKDIR /go/src/github.com/jcalabro/counter.git
RUN go install ./...

CMD ["/go/bin/counterd"]
