FROM golang:1.13.3

WORKDIR /go/src/github.com/panghostlin/Pictures/

ADD go.mod .
ADD go.sum .
RUN go mod download

ADD . /go/src/github.com/panghostlin/Pictures

RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o panghostlin-pictures

ENTRYPOINT ["./panghostlin-pictures"]
EXPOSE 8012