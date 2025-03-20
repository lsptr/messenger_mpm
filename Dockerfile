FROM golang:1.23.5-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o messenger .

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/messenger .
EXPOSE 8080
CMD ["./messenger"]