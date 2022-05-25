FROM golang:alpine3.15 AS builder

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
RUN  go build -x -a -installsuffix nocgo -o transmission-bot && \
     cp transmission-bot /app/transmission-bot

# run
FROM alpine:3.15 AS release

LABEL name="Transmission interface for telegram"
LABEL maintainer="https://github.com/0x0BSoD"
LABEL version="1.0"

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /app/transmission-bot /app/transmission-bot

COPY ./files/templates /app/templates
COPY ./files/media/error.mp4 /app/media/error.mp4

ENTRYPOINT ["/app/transmission-bot"]

CMD ["-h"]
