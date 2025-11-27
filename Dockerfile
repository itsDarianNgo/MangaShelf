FROM golang:1.25-alpine

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata git build-base

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o /usr/local/bin/mangashelf ./cmd/mangashelf

EXPOSE 8080
VOLUME ["/data"]

ENTRYPOINT ["mangashelf"]
