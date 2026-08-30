FROM golang:1.26-alpine AS base

WORKDIR /notegic-backend

RUN apk add --no-cache git

COPY contracts/go.mod contracts/go.sum contracts/
COPY shared/go.mod shared/go.sum shared/
COPY runtimes/core/go.mod runtimes/core/go.sum runtimes/core/
WORKDIR /notegic-backend/runtimes/core
RUN go mod download
WORKDIR /notegic-backend

COPY . .

FROM base AS dev

WORKDIR /notegic-backend/runtimes/core
RUN go install github.com/air-verse/air@latest

FROM base AS builder

WORKDIR /notegic-backend/runtimes/core
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/core ./commands

FROM alpine:latest AS production

WORKDIR /app

RUN apk --no-cache add ca-certificates

COPY --from=builder /out/core /app/core

EXPOSE 7778

CMD ["/app/core"]
