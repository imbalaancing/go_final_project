FROM golang:1.22.3 AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN apt-get update && apt-get install -y gcc
RUN go mod download

COPY . .

RUN go build -o /my_app ./cmd/server/main.go

COPY web /app/web

ENV CGO_ENABLED=1
ENV TODO_PORT=7540
ENV TODO_DBFILE=/app/scheduler.db

EXPOSE 7540

CMD ["/my_app"]