# docker build -f server/Dockerfile -t kerasevo/server .
FROM golang:latest

COPY server /tmp/server

WORKDIR /tmp/server
RUN go mod tidy 
RUN go build -o /main main.go

WORKDIR /
ENTRYPOINT ["/main"]