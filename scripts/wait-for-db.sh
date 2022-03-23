#!/bin/bash

for i in {1..120}; do
  running=`docker inspect -f {{.State.Running}} koko_postgres 2>&1`
  if [[ $running == "true" ]]; then
    echo 'db running'
    exit 0
  fi
  echo 'waiting for db...'
  sleep 1
done
exit 1
