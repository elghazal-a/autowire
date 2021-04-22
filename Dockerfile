FROM golang:1.16.2-buster AS builder


WORKDIR /go/src/github.com/geniousphp/autowire


COPY go.mod go.sum ./
# Download all dependencies
RUN go mod download


COPY . .

RUN CGO_ENABLED=0 GOOS=linux GARCH=amd64 go build -a -installsuffix cgo -o /usr/bin/autowire .

CMD ["/usr/bin/autowire"]