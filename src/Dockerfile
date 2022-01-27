FROM golang:1.17.5 as builder

ARG PROJECT_NAME="app"
ARG VERSION="dev"
ARG COMMIT="none"
ENV CGO_ENABLED=0
ENV GOPROXY=https://proxy.golang.org

WORKDIR /app

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/app_bin .

FROM debian:buster

WORKDIR /app

COPY --from=builder /app/app_bin /app

ENTRYPOINT ["/app/app_bin"]