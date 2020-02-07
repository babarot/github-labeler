FROM golang:1.13.7-stretch

WORKDIR /go/src/app
COPY . /go/src/app
RUN go get -d -v ./...
RUN go build -o /usr/local/bin/github-labeler main.go

COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
