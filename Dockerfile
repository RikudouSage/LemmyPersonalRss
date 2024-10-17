FROM golang:1.22 as build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o rss

FROM alpine:latest

COPY --from=build /app/rss /app

CMD ["/app"]
