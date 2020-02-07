FROM golang:1.13
#AS build

WORKDIR /go/src/app
COPY . /go/src/app
RUN go get -d -v ./...
RUN go build -o /go/bin/github-labeler main.go

# FROM busybox
# COPY --from=build /go/bin/app /github-labeler

ENTRYPOINT ["/go/bin/github-labeler"]
