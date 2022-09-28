#!/bin/bash

if [ $# -eq 0 ]; then
  echo "Missing argument to indicate DB dialect"
  exit 1
fi

case $1 in
  mysql)
    docker run \
      -e MYSQL_DATABASE=koko \
      -e MYSQL_USER=koko \
      -e MYSQL_PASSWORD=koko \
      -e MYSQL_ALLOW_EMPTY_PASSWORD=yes \
      -p 3306:3306 \
      --name "koko_$1" \
      mysql:8.0.17
    ;;

  postgres)
    docker run \
      -e POSTGRES_USER=koko \
      -e POSTGRES_PASSWORD=koko \
      -p 5432:5432 \
      --name "koko_$1" \
      postgres
    ;;
esac
