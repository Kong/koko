#!/bin/bash -x

# Ensure the required services are started
sudo service docker start
sudo service postgresql start

(
  cd /workspace/koko || return 1

  # Generate a clustering certificate/key for the control plane
  openssl req -new \
              -x509 \
              -nodes \
              -newkey ec:<(openssl ecparam -name secp384r1) \
              -out cluster.crt \
              -keyout cluster.key \
              -days 365 \
              -subj "/CN=kong_clustering/ST=California/L=San Francisco/O=Kong, Inc./OU=Engineering"
)