FROM golang:1.13.3

WORKDIR /go/src/github.com/panghostlin/Pictures/

ADD go.mod .
ADD go.sum .
RUN go mod download

ADD . /go/src/github.com/panghostlin/Pictures

RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o panghostlin-pictures

ENTRYPOINT [ "/bin/bash", "-c" ]
CMD ["./wait-for-it.sh" , "panghostlin-postgre:54320" , "--strict" , "--timeout=300" , "--" , "./panghostlin-pictures"]
EXPOSE 8012