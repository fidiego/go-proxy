#!/bin/sh

set -e
set -o pipefail
set -u

COMMIT=$(git rev-parse --short HEAD)

docker build -t go-proxy . -f docker/Dockerfile
docker tag go-proxy:latest go-proxy:$COMMIT

