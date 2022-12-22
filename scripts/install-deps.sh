#!/bin/bash -e

curl -sSfL "https://raw.githubusercontent.com/Kong/go-atc-router/trunk/build-deps.sh" \
  | bash -s -- --build --cache --install --rm
