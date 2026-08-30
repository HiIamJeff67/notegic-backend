FROM golang:1.26-alpine AS base

WORKDIR /notegic-backend

RUN apk add --no-cache git

COPY contracts/go.mod contracts/go.sum contracts/
COPY shared/go.mod shared/go.sum shared/
COPY runtimes/realtimegateway/go.mod runtimes/realtimegateway/go.sum runtimes/realtimegateway/
WORKDIR /notegic-backend/runtimes/realtimegateway
RUN go mod download
WORKDIR /notegic-backend

COPY . .

FROM base AS dev

WORKDIR /notegic-backend/runtimes/realtimegateway

RUN go install github.com/air-verse/air@latest

FROM base AS builder

WORKDIR /notegic-backend/runtimes/realtimegateway

RUN CGO_ENABLED=0 GOOS=linux go build -o /out/realtimegateway ./commands

FROM alpine:latest AS production

WORKDIR /app

RUN apk --no-cache add ca-certificates

COPY --from=builder /out/realtimegateway /app/realtimegateway

EXPOSE 7779

CMD ["/app/realtimegateway"]
