# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -o Digest Engine cmd/Digest Engine/main.go

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates sqlite

WORKDIR /app

COPY --from=builder /app/Digest Engine .
COPY templates/ templates/

VOLUME ["/app/data"]

EXPOSE 9090

ENTRYPOINT ["./Digest Engine"]
CMD ["serve"]
