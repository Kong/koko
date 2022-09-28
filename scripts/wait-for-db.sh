#!/bin/bash

if [ $# -eq 0 ]; then
  echo "Missing argument to indicate DB dialect"
  exit 1
fi

if [[ $1 == "sqlite3" ]]; then
  exit 0
fi

for _ in {1..120}; do
  running=$(docker inspect -f '{{.State.Running}}' "koko_$1" 2>&1)
  if [[ $running == "true" ]]; then
    echo 'db running'
    exit 0
  fi
  echo 'waiting for db...'
  sleep 1
done
exit 1
