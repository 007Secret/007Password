FROM golang:1.23 AS builder
WORKDIR /app
COPY ./go.mod ./go.sum ./
RUN go mod download
COPY ./ ./
RUN go build -o main .

FROM golang:1.23
WORKDIR /app
COPY --from=builder /app/main .
RUN apt-get update && apt-get install -y sqlite3 libsqlite3-dev && rm -rf /var/lib/apt/lists/*
EXPOSE 8080
CMD ["./main"] 