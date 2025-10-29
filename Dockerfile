FROM golang:1.25-alpine

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o main ./cmd/service

COPY ./templates ./templates
COPY ./static ./static

EXPOSE 8080

CMD ["./main"]