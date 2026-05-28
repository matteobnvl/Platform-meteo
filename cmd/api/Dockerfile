# stage dev & build
FROM golang:1.26-alpine AS builder

WORKDIR /app

# installation de git et d'Air pour le live-reload en dev
RUN apk add --no-cache git && go install github.com/air-verse/air@latest

COPY go.mod ./
RUN go mod download

COPY . .

# compilation du binaire
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/api_bin ./cmd/api

# stage production
FROM alpine:3.19 AS prod

WORKDIR /app

RUN adduser -D appuser
USER appuser

# copie du binaire compilé
COPY --from=builder /app/api_bin .

EXPOSE 8080

CMD ["./api_bin"]