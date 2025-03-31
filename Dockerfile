FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY .env ./
COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o gophermart ./cmd/gophermart/main.go

FROM alpine:latest

WORKDIR /app

RUN apk --no-cache add ca-certificates

COPY --from=builder /app/gophermart .

EXPOSE 8080

CMD ["./gophermart"]
