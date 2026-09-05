FROM golang:1.26-alpine AS base

WORKDIR /notegic-backend

RUN apk add --no-cache git

COPY contracts/go.mod contracts/go.sum contracts/
COPY shared/go.mod shared/go.sum shared/
COPY runtimes/clientgateway/go.mod runtimes/clientgateway/go.sum runtimes/clientgateway/
WORKDIR /notegic-backend/runtimes/clientgateway
RUN go mod download
WORKDIR /notegic-backend

COPY . .

FROM base AS dev

WORKDIR /notegic-backend/runtimes/clientgateway

RUN go install github.com/air-verse/air@latest

FROM base AS builder

WORKDIR /notegic-backend/runtimes/clientgateway

RUN CGO_ENABLED=0 GOOS=linux go build -o /out/clientgateway ./commands

FROM alpine:latest AS production

WORKDIR /app

RUN apk --no-cache add ca-certificates

COPY --from=builder /out/clientgateway /app/clientgateway
COPY --from=base /notegic-backend/runtimes/email/templates/assets/common.svg /app/runtimes/email/templates/assets/common.svg

EXPOSE 7777

CMD ["/app/clientgateway"]
