FROM golang:latest as builder
ARG CGO_ENABLED=0
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN go build -o main cmd/main.go

FROM scratch
COPY --from=builder /app/config/config.yaml /config/config.yaml
COPY --from=builder /app/main /main

EXPOSE 8080

ENTRYPOINT ["/main"]