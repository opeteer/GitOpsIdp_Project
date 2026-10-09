FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/idp-api ./cmd/server

FROM alpine:3.20
RUN apk --no-cache add ca-certificates git tzdata
WORKDIR /workspace

COPY --from=builder /app/idp-api /usr/local/bin/idp-api

EXPOSE 8080
ENV PORT=8080
ENV PROJECT_ROOT=/workspace

ENTRYPOINT ["/usr/local/bin/idp-api"]
