FROM golang:alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY cmd cmd/
COPY config config/
COPY db db/
COPY internal internal/
COPY docs docs/

RUN go build -o main ./cmd/geoservice

FROM alpine:latest
COPY --from=builder /app/config config/
COPY --from=builder /app/main main
COPY --from=builder /app/docs docs/
CMD ["./main"]
