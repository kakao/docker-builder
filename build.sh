#!/bin/sh
if [ $# != 1 ]; then
  echo "please input \"version\""
  echo "ex) $0 v1.0"
  exit 1
fi

# CURRENT_DIR=$(pwd)
# cd $GOROOT/src; sudo GOOS=linux GOARCH=amd64 ./make.bash —no-clean; cd $CURRENT_DIR;
rm -rf bin
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bin/docker-builder

VERSION=$1
IMAGE_NAME="d2hub.com/docker-builder"
REPLACE_IMAGE_NAME=$(echo $IMAGE_NAME | sed -e 's/\//'"\\\\\/"'/g')
CURRENT_IMAGE=$IMAGE_NAME:$VERSION
LATEST_IMAGE=$IMAGE_NAME:latest

docker build -t "$CURRENT_IMAGE" .
docker tag "$CURRENT_IMAGE" "$LATEST_IMAGE"
#docker push "$CURRENT_IMAGE"
#docker push "$LATEST_IMAGE"
