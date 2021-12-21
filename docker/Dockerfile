FROM golang:1.17.5 as build

WORKDIR /build

COPY . .

RUN go build -o mg-api ./cmd

FROM alpine as app

WORKDIR /app

COPY --from=build /build/mg-api .

EXPOSE 8888

CMD [ "cd /app && ./mg-api" ]