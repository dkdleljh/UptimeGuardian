FROM golang:1.22-alpine AS builder
WORKDIR /src
COPY go.mod ./
COPY cmd ./cmd
COPY internal ./internal
RUN go build -o /out/healthcheck-bot ./cmd/server

FROM alpine:3.20
WORKDIR /app
RUN addgroup -S app && adduser -S app -G app
COPY --from=builder /out/healthcheck-bot /app/healthcheck-bot
USER app
EXPOSE 8080
ENTRYPOINT ["/app/healthcheck-bot"]
