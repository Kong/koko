# koko - Control Plane for Kong Gateway

[![Build](https://github.com/Kong/koko/actions/workflows/dev-builds.yaml/badge.svg)](https://github.com/Kong/koko/actions/workflows/dev-builds.yaml)

Koko is the second generation Control-Plane (CP) for
[Kong](https://github.com/kong/kong) API Gateway's
[Hybrid mode](https://docs.konghq.com/gateway/latest/plan-and-deploy/hybrid-mode/)
deployment.
Koko is designed to simplify Kong Gateway deployments and to decouple 
Control-Plane (CP) aspects from the powerful underlying NGINX/OpenResty 
Data-Plane (DP) stack.

## Table of Contents

- [**Status**](#status)
- [**Documentation**](#documentation)
- [**Features**](#features)
- [**Compatibility**](#compatibility)
- [**Seeking help**](#seeking-help)
- [**License**](#license)

## Status

Koko is currently under heavy development and is considered beta quality software.
Breaking changes, although rare, should be expected.

Koko is used in production environments at Kong Inc since April 2022.

## Documentation

Documentation can be found inside [docs](docs/) directory.

## Features

Some notable features of Koko:

- Redesigned Admin API, supports gRPC and HTTP
-  [OpenAPI v2 spec](https://github.com/Kong/koko/blob/main/internal/gen/swagger/koko.swagger.json)
  and [JSON Schema](https://github.com/Kong/koko/tree/main/internal/gen/jsonschema/schemas)
  to aid in development of custom integrations, APIs as well as UIs
- Supports SQLite, Postgres and MySQL as storage backends to run Kong in any
  environment including edge platforms as well as Serverless
- Instrumented for operations: structured logs, Prometheus metrics and planned
  support for OpenTracing integration
- Version compatibility insights to easy upgrades of Kong Gateway and detailed
  insights around deprecations as well as compatibility of new feature with older
  Kong versions

## Compatibility

Koko is compatible for all Kong Gateway version >= 2.5.
The recommended version of Kong Gateway is 3.0.

## Seeking help

Please open a GitHub [Issue](/issues/new/choose) for reporting bugs, providing
feedback, asking questions, and feature requests.

## License

```
Copyright 2022 Kong Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
```

