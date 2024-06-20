FROM golang:latest AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN apt-get update && apt-get install -y gcc
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 go build -o /my_app ./cmd/server/main.go

COPY web /app/web

EXPOSE 7540

ENV HTTP_PORT=7540
ENV TODO_DBFILE=/app/scheduler.db

CMD ["/my_app"]