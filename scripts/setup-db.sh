#!/bin/bash

docker run \
  -e POSTGRES_USER=koko \
  -e POSTGRES_PASSWORD=koko \
  -p 5432:5432 \
  postgres
