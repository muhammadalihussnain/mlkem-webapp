FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o mlkem-server ./cmd/server/

FROM alpine:3.20
WORKDIR /app
COPY --from=builder /app/mlkem-server .
EXPOSE 8080
CMD ["./mlkem-server"]
