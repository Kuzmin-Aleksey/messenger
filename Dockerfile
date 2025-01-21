FROM golang:1.17.1-alpine3.14 as modules
COPY go.mod go.sum /modules/
WORKDIR /modules
RUN go mod download

FROM golang:1.23.2
RUN mkdir /app
ADD . /app
WORKDIR /app
RUN go build -o main cmd/main.go
CMD ["/app/main"]
