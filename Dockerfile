FROM golang:1.22.3

WORKDIR /app

COPY . .

RUN apt-get update && apt-get install -y gcc
RUN go mod download

RUN go build -o /my_app ./cmd/server/main.go

ENV CGO_ENABLED=1
ENV TODO_PORT=7540
ENV TODO_DBFILE=/app/scheduler.db

EXPOSE ${TODO_PORT}

CMD ["/my_app"]