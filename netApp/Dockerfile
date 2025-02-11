FROM golang:alpine

WORKDIR /app

COPY . .

RUN go build -o build/migrate migrations/auto.go
RUN go build -o build/main cmd/main.go

COPY entrypoint.sh /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh

ENTRYPOINT ["/bin/sh", "/app/entrypoint.sh"]