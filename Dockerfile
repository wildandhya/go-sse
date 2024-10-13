FROM golang:1.19-alpine AS builder

WORKDIR /app

COPY ./go.mod ./go.sum ./
RUN go mod download

COPY . .

RUN go build -o sse-api .

FROM alpine:latest
WORKDIR /root/

COPY --from=builder /app/sse-api .

EXPOSE 8080

CMD ["./sse-api"]