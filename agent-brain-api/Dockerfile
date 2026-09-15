# syntax=docker/dockerfile:1
FROM golang:1.22-alpine AS build
WORKDIR /src
RUN apk add --no-cache ca-certificates git
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/agent-brain-api ./cmd/server

FROM alpine:3.20
WORKDIR /app
RUN apk add --no-cache ca-certificates tzdata \
  && adduser -D -H -u 10001 appuser
COPY --from=build /out/agent-brain-api /app/agent-brain-api
COPY migrations /app/migrations
USER appuser
EXPOSE 8220
ENV ADDR=:8220 \
    MIGRATIONS_DIR=migrations
ENTRYPOINT ["/app/agent-brain-api"]
