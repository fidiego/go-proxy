# vim: set syntax=dockerfile
# Dockerfile for golang development
FROM golang:1.22.3-bullseye as builder

# set workdir
WORKDIR /app

# install dependencies
COPY go.mod go.sum /app/
RUN go mod download

# copy source code
COPY . /app/

# build the go binary
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/go-proxy /app/cmd/web/

# From scratch
FROM gcr.io/distroless/base
COPY --from=builder /app/go-proxy /usr/local/bin/go-proxy
EXPOSE 8080

# run the server
CMD ["go-proxy"]
