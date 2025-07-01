FROM golang:alpine AS builder

WORKDIR /myapp

# Устанавливаем зависимости для SQLite3
RUN apk add --no-cache gcc musl-dev sqlite-dev

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -o ./cmd/app ./cmd/app  # CGO_ENABLED=1 

EXPOSE 8080

CMD ["./cmd/app/app"]