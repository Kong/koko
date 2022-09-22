# Getting started with Koko

## System requirements

- Docker
- cURL

## Create a new directory

```shell
mkdir koko-demo
cd  koko-demo
```

## Create certificates for CP-DP authentication

```shell
mkdir certs
docker run -u $(id -u ${USER}):$(id -g ${USER})  -v $(pwd)/certs:/certs kong kong hybrid gen_cert /certs/cluster.crt /certs/cluster.key
# make the key readable, required for following containers
chmod +r certs/cluster.key
```

## Docker compose file

Download the [docker-compose.yml](./assets/docker-compose.yml) file:

```shell
curl https://raw.githubusercontent.com/Kong/koko/main/docs/assets/docker-compose.yml \
  -o docker-compose.yml
```

## Start Docker containers

```shell
docker compose up
```

## Start using Kong

Configure a service and a route:

```shell
curl -X PUT \
http://localhost:3000/v1/services/001156b4-b228-4c15-91a1-2de253f2a95c \
-d '{ "name": "mockbin", "host": "mockbin.org" }' \
-H 'content-type: application/json'

curl -X POST \
http://localhost:3000/v1/routes \
-d '{ "name": "mockbin", "paths": [ "/foo" ], "service": { "id": "001156b4-b228-4c15-91a1-2de253f2a95c" } }' \
-H 'content-type: application/json'
```

Configure a rate-limting plugin on the service:

```shell
curl -X POST \
http://localhost:3000/v1/plugins  \
-d '{ "name": "rate-limiting", "config": { "minute": 10 }, "service": { "id": "001156b4-b228-4c15-91a1-2de253f2a95c" }}' \
-H 'content-type: application/json'
```

Execute a request against Kong:

```
curl -v http://localhost:8000/foo/status/200
```

Observe that Kong correctly routes the request to
[mockbin.org](https://mockbin.org).
Kong also enforces appropriate rate-limits and the rate-limits can be found in
the HTTP response headers.

## Check tracked data-plane node

Koko tracks each node that connects to it and you can grab that information
using an Admin API call:

```shell
curl http://localhost:3000/v1/nodes
```

You should see one node with its version information as well as compatibility
status.

## Next steps

Please explore the
[OpenAPI v2 spec](https://github.com/Kong/koko/blob/main/internal/gen/swagger/koko.swagger.json)
of the Admin API offered by Koko to configure Kong, or hook up another Kong data-plane.

## Seeking help

If you run into any issues, please open a GitHub Issue.
