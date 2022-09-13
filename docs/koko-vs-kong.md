# Koko Control-Plane vs Kong Control-Plane

## Core features

This document provides a high-level comparison of Koko and Kong Gateway's
built-in Control-Plane.

The following key is used within the document:

- :x: : Not supported and isn't _likely_ to be supported.
- :heavy_check_mark: : Supported
- :calendar: : Planned for a future release

| Feature                               | Kong's builtin CP  | Koko               |
|---------------------------------------|--------------------|--------------------|
| **Core API**                          |                    |                    |
| Existing HTTP Admin-API               | :heavy_check_mark: | :calendar:         |
| All core entities                     | :heavy_check_mark: | :heavy_check_mark: |
| Auth plugins                          | :heavy_check_mark: | :calendar:         |
| decK integration                      | :heavy_check_mark: | :calendar:         |
| Tag-based filtering                   | :heavy_check_mark: | :heavy_check_mark: |
| New versioned Admin-API               | :x:                | :heavy_check_mark: |
| gRPC Admin API                        | :x:                | :heavy_check_mark: |
| Admin API language-specific clients   | :x:                | :calendar:         |
| **Storage**                           |                    |                    |
| Postgres                              | :heavy_check_mark: | :heavy_check_mark: |
| SQLite                                | :x:                | :heavy_check_mark: |
| MySQL                                 | :x:                | :calendar:         |
| SQL Server                            | :x:                | :calendar:         |
| CockroachDB                           | :x:                | :calendar:         |
| **Plugins**                           |                    |                    |
| Plugins with Lua-schemas              | :heavy_check_mark: | :heavy_check_mark: |
| OpenResty-specific schema validations | :heavy_check_mark: | :x:                |
| Plugins with Custom DAOs              | :heavy_check_mark: | :x:                |
| **Adopting standards**                |                    |                    |
| JSON schemas for resources            | :x:                | :heavy_check_mark: |
| OpenAPI spec                          | :x:                | :heavy_check_mark: |
| **Observability**                     |                    |                    |
| Structured logs                       | :x:                | :heavy_check_mark: |
| Deep Prometheus metrics               | :x:                | :calendar:         |
| Distributed tracing (Admin)           | :x:                | :calendar:         |
| **Data-plane**                        |                    |                    |
| wRPC-based DP communication           | :heavy_check_mark: | :heavy_check_mark: |
| Cluster state visibility              | :x:                | :heavy_check_mark: |
| mTLS-based Data-Plane auth            | :heavy_check_mark: | :calendar:         |
| Non mTLS-based Data-Plane auth        | :x:                | :calendar:         |
| k8s Gateway API integration           | :x:                | :calendar:         |
| **Misc**                              |                    |                    |
| Secrets referencing                   | :heavy_check_mark: | :calendar:         |
| Version compatibility insights        | :x:                | :heavy_check_mark: |


## Plugins

All plugins, other than the ones noted below are supported by Koko.

The following plugins are not supported:

- acl
- basic-auth
- hmac-auth
- jwt
- key-auth
- oauth2

Of these, all plugins except the `oauth2` plugin are planned for inclusion.
`oauth2` plugin is not compatible with Hybrid mode of Kong and hence there are
no plans to support it.


#### Have a question?

If you have a question, please open a GitHub Issue in this repository.

