FROM golang:1.13
ENV GO111MODULE=on
RUN go get -u golang.org/x/lint/golint
RUN mkdir -p /workspace
WORKDIR /workspace
COPY go.mod ./
COPY notify.go ./
COPY notify_test.go ./
RUN go build -v ./...
