FROM golang:1.10.5 AS builder

ENV DEP_VERSION 0.5.0
RUN curl -fsSL -o /usr/local/bin/dep https://github.com/golang/dep/releases/download/v${DEP_VERSION}/dep-linux-amd64 \
  && chmod +x /usr/local/bin/dep

WORKDIR /go/src/github.com/geniousphp/autowire
COPY vendor .
COPY Gopkg.toml Gopkg.lock ./
# install the dependencies without checking for go code
RUN dep ensure -vendor-only
COPY . /go/src/github.com/geniousphp/autowire
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /usr/bin/autowire .

CMD ["/usr/bin/autowire"]