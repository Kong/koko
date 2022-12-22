#!/bin/bash -e

SCRIPT_REF="$1"
DSTDIR="$2"

curl -sSfL "https://raw.githubusercontent.com/Kong/$SCRIPT_REF" | bash -s -- "$DSTDIR"
