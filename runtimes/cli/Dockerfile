FROM golang:1.26-alpine AS dev

WORKDIR /notegic-backend

RUN apk add --no-cache git

FROM dev AS builder

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /out/notegic ./runtimes/cli

FROM alpine:latest AS production

WORKDIR /app

RUN apk --no-cache add ca-certificates

COPY --from=builder /out/notegic /app/notegic

ENTRYPOINT ["/app/notegic"]
