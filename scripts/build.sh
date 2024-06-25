#!/bin/sh

set -e
set -o pipefail
set -u


COMMIT=$(git rev-parse --short HEAD)
IMAGE=sayheytofred/go-proxy


# check if the git repo is dirty
if [[ $(git diff --stat .) != '' ]]; then
        echo
        echo ' 🤖 I am sorry '`whoami`','
        echo
        echo '    I cannot let you do that.'
        echo
        echo '    The git repo is dirty. This script should only run on a clean repo.'
        echo
        echo 'The changes in question:'
        echo
        git diff --stat .
        echo
        exit 1
else
        echo
        echo '=== Preparing to build and push image: '$IMAGE':'$COMMIT
fi


# build the image
docker build -t $IMAGE . -f docker/Dockerfile

# tag the image with the commit and latest
docker tag $IMAGE:latest $IMAGE:$COMMIT

# push the images
docker push $IMAGE:latest
docker push $IMAGE:$COMMIT
