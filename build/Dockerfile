FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /foundry ./cmd/foundry

FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /foundry /usr/local/bin/foundry

EXPOSE 8080

ENTRYPOINT ["foundry"]
