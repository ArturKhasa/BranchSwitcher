FROM golang:alpine AS builder

LABEL stage=gobuilder
ENV CGO_ENABLED 0

RUN apk update --no-cache && apk add --no-cache tzdata

WORKDIR /build

# Копируем файлы из корня проекта
COPY go.mod .
RUN go mod download
COPY . .

RUN go build -ldflags="-s -w" -o /app/main main.go


FROM debian:bullseye-slim
COPY --from=builder /app/main /app/main

WORKDIR /app
RUN apt-get update && apt-get install -y git openssh-client && \
    mkdir -p /root/.ssh && chmod 700 /root/.ssh && \
    ssh-keyscan github.com > /root/.ssh/known_hosts

CMD ["./main"]