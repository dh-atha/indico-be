#syntax=docker/dockerfile:1

ARG GO_VERSION=1.24

FROM golang:${GO_VERSION}-alpine AS build
WORKDIR /app

# Install build deps
RUN apk add --no-cache git ca-certificates

# Pre-copy go.mod to leverage layer cache
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/server ./

FROM gcr.io/distroless/static-debian12 AS runtime
WORKDIR /app

ENV PORT=8080 \
    WORKERS=8 \
    DATABASE_URL="postgres://postgres:password@localhost:5432/appdb?sslmode=disable"

COPY --from=build /out/server /app/server

EXPOSE 8080

ENTRYPOINT ["/app/server"]
