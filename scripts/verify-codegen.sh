#!/bin/bash

function is_repo_dirty() {
  status=$(git status --porcelain)
  if [[ -z "$status" ]];
  then
    echo 0
  else
    echo 1
  fi
}

if [[ $(is_repo_dirty) -eq 1 ]];
then
  echo "Git tree is dirty, please commit your changes."
  exit 1
fi

rm -rf ./internal/gen

./scripts/update-codegen.sh

if [[ $? -ne 0 ]];
then
  exit $?
fi

if [[ $(is_repo_dirty) -eq 1 ]];
then
  echo "Codegen was out of date"
  exit 1
fi
