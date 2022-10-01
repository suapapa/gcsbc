#!/bin/bash
git tag -a $1 -m "add tag for $1"
git push --tags

IMAGE_TAG=gcr.io/homin-dev/gcsbc
docker buildx build --platform linux/amd64 -t $IMAGE_TAG:$1 .
docker push $IMAGE_TAG

IMAGE_TAG_LATEST=$IMAGE_TAG:latest 
docker tag $IMAGE_TAG $IMAGE_TAG_LATEST
docker push $IMAGE_TAG_LATEST
