#!/usr/bin/env bash

VERSION=$(date +%Y-%m-%dT%H.%M.%S)-$(git log -1 --pretty=format:"%h")

IMG=modfin/mmailer
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
