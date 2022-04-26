#!/bin/bash

DP_SPEC_FOLDER=$PWD/internal/wrpc/proto/kong

# clenup
rm -rf kong-dp-spec $DP_SPEC_FOLDER || exit 0

# clone repo
git clone git@github.com:Kong/kong-dp-spec.git
cp -r kong-dp-spec/spec/proto/kong $DP_SPEC_FOLDER

for spec in $(seq 1 $#); do
  echo removing $DP_SPEC_FOLDER/$1
  rm -r $DP_SPEC_FOLDER/$1
  shift
done

# clenup
rm -rf kong-dp-spec || exit 0