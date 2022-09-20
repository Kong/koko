#!/bin/bash -e

while [ -n "$*" ]; do
  case "$1" in
    --user )
      DOCKER_HUB_USERNAME="$2"
      shift
      ;;

    --token )
      DOCKER_HUB_ACCESS_TOKEN="$2"
      shift
      ;;

    --tag )
      GIT_TAG="$2"
      shift
      ;;
  esac
  shift
done

if [ -z "${DOCKER_HUB_USERNAME}" -o -z "${DOCKER_HUB_ACCESS_TOKEN}" ]; then
  echo "Missing Docker credentials."
  echo "Please set the variables DOCKER_HUB_USERNAME and DOCKER_HUB_ACCESS_TOKEN."
  exit
fi

if [ -z "${GIT_TAG}" ]; then
  echo "Missing version tag."
  echo "Please set the GIT_TAG variable with the version name."
  exit
fi

if [ -n "$(git status --porcelain)" ]; then
  echo "Git tree is dirty, please commit your changes."
  exit 1
fi

IMG_NAME="${IMG_NAME:-kong/koko}"
GIT_COMMIT_HASH="${GIT_COMMIT_HASH:-$(git rev-parse --short HEAD)}"

set -v

echo "${DOCKER_HUB_ACCESS_TOKEN}" |
  docker login --username "${DOCKER_HUB_USERNAME}" --password-stdin

docker buildx build \
  --build-arg "GIT_TAG=${GIT_TAG}" \
  --build-arg "GIT_COMMIT_HASH=${GIT_COMMIT_HASH}" \
  -t "${IMG_NAME}:${GIT_TAG}" \
  --push \
  .
