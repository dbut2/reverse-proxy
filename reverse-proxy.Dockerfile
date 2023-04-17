FROM golang:alpine AS builder

WORKDIR /app

COPY ./go.mod ./go.mod
COPY ./go.sum ./go.sum
COPY ./reverse-proxy.go ./main.go

RUN go build -o reverse-proxy main.go

FROM alpine:latest

COPY --from=builder /app/reverse-proxy ./reverse-proxy

ENV PORT=8080

EXPOSE $PORT

CMD ["./reverse-proxy"]
