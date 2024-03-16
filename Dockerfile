FROM golang:1.22.1-alpine3.19 as builder
WORKDIR /app
COPY go.* ./
RUN go mod download
COPY . ./
RUN GOOS=linux GOARCH=amd64 go build -o app ./OnlineCinema

FROM alpine:3.19 as app
RUN apk --no-cache upgrade && apk --no-cache add ca-certificates
COPY --from=builder /app/app /usr/local/bin/app
WORKDIR /usr/local/bin
CMD ["app"]