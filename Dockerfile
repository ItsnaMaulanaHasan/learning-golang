FROM golang:1.23.3-alpine AS builder

WORKDIR /build

COPY . .

RUN go get ./...

RUN go build -o backend main.go

FROM alpine

WORKDIR /app

COPY --from=builder /build/backend /app/backend

COPY --from=builder /build/templates /app/templates

ENTRYPOINT [ "/app/backend" ]