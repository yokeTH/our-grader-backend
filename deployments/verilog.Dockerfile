FROM golang:1.24

WORKDIR /app

RUN apt-get update && apt-get install -y \
    build-essential \
    python3 \
    python3-pip \
    python3-dev \
    gcc \
    make \
    iverilog

RUN pip3 install cocotb --break-system-packages

COPY ../go.mod ../go.sum ./

RUN go mod download

COPY .. .

RUN go build -o bin/server ./grading/server.go

RUN chmod +x /app/bin/server

CMD ["/app/bin/server"]