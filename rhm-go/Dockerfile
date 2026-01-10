# Build Stage
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod init rhm-go && go mod tidy
RUN go build -o rhm-server ./cmd/rhm-server

# Run Stage
FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/rhm-server .
EXPOSE 8080
CMD ["./rhm-server"]
