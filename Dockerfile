FROM golang:1.17 AS builder

WORKDIR /app

ENV GO111MODULE=on

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o mg-api cmd/main.go

FROM alpine

WORKDIR /app

COPY --from=builder /app/mg-api .

ENTRYPOINT [ "./mg-api" ]