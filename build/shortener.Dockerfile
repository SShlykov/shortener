FROM golang:1.23-alpine AS build

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN GOOS=linux GOARCH=amd64 go mod download

COPY . .

RUN GOOS=linux GOARCH=amd64 go build -o /app/server ./cmd/shortener/main.go

FROM alpine:3.12

WORKDIR /bin

USER 5000

COPY --from=build /app/server .
COPY --from=build /app/config /config

CMD ["/bin/server", "-config=/config"]