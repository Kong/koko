#!/bin/bash -e

MODULE_NAME="$1"
SCRIPT_FILE="$2"
DSTDIR="$3"

MODULE_VERSION=$(grep -F "$MODULE_NAME" go.mod | (read mod ver; echo "${ver##*-}"))

curl -sSfL "https://raw.githubusercontent.com/$MODULE_NAME/$MODULE_VERSION/$SCRIPT_FILE" | bash -s -- "$DSTDIR"
