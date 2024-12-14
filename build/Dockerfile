FROM golang:1.23.2 AS builder

RUN mkdir -p /app
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
RUN go install github.com/google/wire/cmd/wire@latest

COPY . .
ARG VERSION
RUN make wire
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-X github.com/thhyl/chatr/cmd.Version=$VERSION -w -s" -o server ./chatr.go

FROM alpine:3.14
RUN apk update && apk add --no-cache ca-certificates

RUN mkdir -p /app
WORKDIR /app
COPY --from=builder /app/server /app/server

ENTRYPOINT ["/app/server"]
