FROM golang:1.14.1-alpine AS builder

ENV SRC_DIR=/go/src/github.com/0x0BSoD/transmission-bot/
ENV GO111MODULE=on
ENV CGO_ENABLED=0
ENV GOOS=linux

WORKDIR $SRC_DIR

# Dependencies
COPY go.mod go.sum ./
RUN mkdir /app &&  \
    apk add git gcc && \
    go mod download

# Update CA
RUN apk --update add ca-certificates

# Copy the code from the host and compile it
ADD . ./

# Build, don't include C libs
RUN  go build -a -installsuffix nocgo -o transmission-bot && \
     cp transmission-bot /app/transmission-bot

# run
FROM scratch

LABEL name="Transmission interface for telegram"
LABEL maintainer="https://github.com/0x0BSoD"
LABEL version="1.0"

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /app/transmission-bot /transmission-bot

COPY templates /templates
COPY error.mp4 /error.mp4

ENTRYPOINT ["/transmission-bot"]