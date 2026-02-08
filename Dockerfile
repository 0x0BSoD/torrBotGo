# Build stage
FROM golang:1.25-alpine AS builder

ARG VERSION=v0.0.0
ARG GIT_COMMIT=dev
ARG BUILD_DATE=unknown

RUN apk add --no-cache git ca-certificates tzdata
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath \
    -ldflags="-s -w -X main.Version=${VERSION} -X main.GitCommit=${GIT_COMMIT} -X main.BuildDate=${BUILD_DATE}" \
    -o torrbot .

# Final stage: distroless base
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/torrbot /app/torrbot
USER nonroot:nonroot
WORKDIR /app

ENTRYPOINT ["/app/torrbot"]

CMD ["serve"]
