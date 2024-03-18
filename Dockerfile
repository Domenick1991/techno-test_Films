FROM golang:1.22.1-alpine3.19 as builder
WORKDIR /usr/local/src
COPY ["OnlineCinema/go.mod", "OnlineCinema/go.sum", "./"]
RUN go mod download
COPY OnlineCinema ./
RUN go build -o ./bin/app main.go

FROM alpine:3.19 as app
COPY --from=builder /usr/local/src/bin/app /
COPY OnlineCinema/config.yml config.yml
CMD ["/app"]