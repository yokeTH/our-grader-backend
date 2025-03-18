FROM golang:1.24-alpine

WORKDIR /app

RUN apk add --no-cache \
    build-base \
    python3 \
    py3-pip \
    python3-dev \
    gcc

COPY ../go.mod ../go.sum ./

RUN go mod download

COPY .. .

RUN go build -o bin/server ./api/cmd/server/main.go

RUN chmod +x /app/bin/server

CMD ["/app/bin/server"]