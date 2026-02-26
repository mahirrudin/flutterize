FROM golang:1.25-alpine AS builder

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o flutterize-api .

FROM alpine:3.19

RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /build/flutterize-api .
COPY --from=builder /build/migration/ ./migration/

EXPOSE 8443

CMD ["./flutterize-api"]
