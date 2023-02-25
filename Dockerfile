FROM golang:1.20
RUN mkdir -p /workspace
WORKDIR /workspace
COPY go.mod ./
COPY notify.go ./
COPY notify_test.go ./
RUN go get github.com/LadySerena/discord-notification-function
RUN go build -v ./...
