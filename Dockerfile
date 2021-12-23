FROM golang:1.17.5

WORKDIR /app

COPY . .

RUN go build -o mg-api cmd/main.go

CMD [ "/app/mg-api" ]