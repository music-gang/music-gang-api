FROM golang:1.18 AS builder

WORKDIR /app

ENV GO111MODULE=on

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN GIT_COMMIT=$(git rev-list -1 HEAD) && \ 
    go build -ldflags "-X main.Commit=$GIT_COMMIT" -o mg-api cmd/mgd/main.go

FROM gcr.io/distroless/base-debian11

WORKDIR /app

COPY --from=builder /app/mg-api .

CMD [ "./mg-api" ]