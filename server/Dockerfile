FROM golang:1.26.4

WORKDIR /app

COPY . .

RUN go mod download

RUN go build -o server ./cmd/server

CMD ["./server"]