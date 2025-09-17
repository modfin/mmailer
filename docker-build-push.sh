#!/usr/bin/env bash

TARGET_ENV=${1:-dev}

VERSION=$(date +%Y-%m-%dT%H.%M.%S)-$(git log -1 --pretty=format:"%h")

if [ "$TARGET_ENV" == "prod" ]; then
  IMG=modfin/mmailer
else
  IMG=modfin/mmailer-dev
fi

COMMIT_MSG=$(git log -1 --pretty=format:"%s" .)
AUTHOR=$(git log -1 --pretty=format:"%an" .)

## Building latest mmailer
docker build -f Dockerfile.build \
    --label "CommitMsg=${COMMIT_MSG}" \
    --label "Author=${AUTHOR}" \
    -t ${IMG}:latest \
    -t ${IMG}:${VERSION} \
    . || exit 1

## Push to repo
docker push ${IMG}:latest
docker push ${IMG}:${VERSION}

## Cleaning up
docker rmi -f ${IMG}:latest
docker rmi -f ${IMG}:${VERSION}
