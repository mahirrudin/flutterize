FROM golang:1.25-alpine AS builder
WORKDIR /build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o mocksms .

FROM alpine:3.19
WORKDIR /app
COPY --from=builder /build/mocksms .
EXPOSE 7600
CMD ["./mocksms"]
