FROM golang:1.26-alpine AS base

WORKDIR /notegic-backend

RUN apk add --no-cache git

COPY contracts/go.mod contracts/go.sum contracts/
COPY shared/go.mod shared/go.sum shared/
COPY runtimes/notification/go.mod runtimes/notification/go.sum runtimes/notification/
WORKDIR /notegic-backend/runtimes/notification
RUN go mod download
WORKDIR /notegic-backend

COPY . .

FROM base AS dev

WORKDIR /notegic-backend/runtimes/notification

RUN go install github.com/air-verse/air@latest

FROM base AS builder

WORKDIR /notegic-backend/runtimes/notification

RUN CGO_ENABLED=0 GOOS=linux go build -o /out/notification ./commands

FROM alpine:latest AS production

WORKDIR /app

RUN apk --no-cache add ca-certificates

COPY --from=builder /out/notification /app/notification

EXPOSE 7781

CMD ["/app/notification"]
