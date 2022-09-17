#!/bin/bash -e

if [ -z "${DOCKER_HUB_USERNAME}" -o -z "${DOCKER_HUB_ACCESS_TOKEN}" ]; then
  echo "Missing Docker credentials."
  echo "Please set the variables DOCKER_HUB_USERNAME and DOCKER_HUB_ACCESS_TOKEN."
  exit
fi

IMG_NAME="${IMG_NAME:-kong/koko}"
GIT_COMMIT_HASH="${GIT_COMMIT_HASH:-$(git rev-parse --short HEAD)}"

if [ -z "${GIT_TAG}" ]; then
  GIT_TAG=$(git describe --tags --match 'v*' || true)
fi
if [ -z "${GIT_TAG}" ]; then
  GIT_TAG="dev-${GIT_COMMIT_HASH}"
fi

set -v

echo "${DOCKER_HUB_ACCESS_TOKEN}" |
  docker login --username "${DOCKER_HUB_USERNAME}" --password-stdin

docker buildx build \
  --build-arg "GIT_TAG=${GIT_TAG}" \
  --build-arg "GIT_COMMIT_HASH=${GIT_COMMIT_HASH}" \
  -t "${IMG_NAME}:latest" -t "${IMG_NAME}:${GIT_TAG}" \
  --push \
  .
