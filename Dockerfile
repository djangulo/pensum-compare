FROM golang:1.16beta1-buster as builder
LABEL maintainer="Denis Angulo <djal@tuta.io>"
WORKDIR /app

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -p 1 -a -o ./pensum ./...

FROM alpine:latest
WORKDIR /app

RUN addgroup -S golang \
    && adduser -S -G golang golang
RUN apk update \
    && apk --no-cache add ca-certificates

COPY --from=builder /app/pensum ./pensum
RUN chmod +x ./pensum \
    && chown golang:golang ./pensum

USER golang
