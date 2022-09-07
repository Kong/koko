# Koko Control-Plane vs Kong Control-Plane

## Core features
This document provides a high-level comparison of Koko and Kong Gateway's
builtin Control-Plane.

The following key is used within the document:

- :x: : Not supported and isn't _likely_ to be supported.
- :heavy_check_mark: : Supported
- :calendar: : Planned on the roadmap for 2022

This table compares features between Kong's builtin Control-Plane and Koko.

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
| SQL-server                            | :x:                | :calendar:         |
| Cockroachdb                           | :x:                | :calendar:         |
| **Plugins**                           |                    |                    |
| Plugins with Lua-schemas              | :heavy_check_mark: | :heavy_check_mark: |
| OpenResty-specific schema validations | :heavy_check_mark: | :x:                |
| Plugins with Custom DAOs              | :heavy_check_mark: | :x:                |
| **Adopting standards**                |                    |                    |
| JSON schemas for resources            | :x:                | :heavy_check_mark: |
| OpenAPI spec                          | :x:                | :heavy_check_mark: |
| Protobuf-based gRPC spec              | :x:                | :heavy_check_mark: |
| **Observability**                     |                    |                    |
| Structured logs                       | :x:                | :heavy_check_mark: |
| Deep Prometheus metrics               | :x:                | :calendar:         |
| Distributed tracing (Admin)           | :x:                | :calendar:         |
| **Data-plane**                        |                    |                    |
| wRPC-based DP comm                    | :heavy_check_mark: | :heavy_check_mark: |
| Cluster state visibility              | :x:                | :heavy_check_mark: |
| Non mTLS-based Data-Plane auth        | :x:                | :calendar:         |
| Inbuilt k8s integration               | :x:                | :calendar:         |
| **Misc**                              |                    |                    |
| Secrets referencing                   | :heavy_check_mark: | :calendar:         |
| Version compatibility insights        | :x:               | :heavy_check_mark: |


## Plugins

All plugins, other than the ones noted below are supported by Koko.

The following plugins are not supported:

- key-auth
- basic-auth
- jwt
- hmac-auth
- acl
- oauth2

Of these, all plugins except the oauth2 plugin are planned for inclusion.
oauth2 plugin is not compatible with Hybrid mode of Kong and hence there are
no plans to support it.


#### Have a question?

If you have a question, please open a GitHub Issue in this repository.

