FROM golang:1.22-alpine AS builder

WORKDIR /app

ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

# Сейчас у нас только go.mod, go.sum нет — поэтому копируем только go.mod
COPY go.mod ./
RUN go mod download || true

COPY . .

RUN go build -o /build/himera-bot ./cmd/bot

FROM alpine:3.20

RUN apk add --no-cache ca-certificates

RUN addgroup -S app && adduser -S app -G app

WORKDIR /app

COPY --from=builder /build/himera-bot /app/himera-bot

ENV APP_ENV=development

USER app

EXPOSE 8080

ENTRYPOINT ["./himera-bot"]
