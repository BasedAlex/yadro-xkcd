FROM golang:1.22.2-alpine AS builder

WORKDIR /app

COPY . .

RUN go build -o xkcdApp ./cmd/xkcd

FROM alpine:latest

RUN mkdir /app

COPY --from=builder /app/xkcdApp /app/xkcdApp

CMD [ "/app/xkcdApp" ]