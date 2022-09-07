# Public contract of Koko

This document lays out public contract, also known as "Compatibility promise"
of Koko. Koko adheres to Semantic Versioning and the maintainers will strive to
ensure that no breaking changes take place in minor or patch releases.
The maintainers of the project reserve rights to introduce breaking changes in
minor releases for the major version number '0'.

Any user-facing interface not listed in this document is not part of the public
contract and hence subject to change at any time. Please avoid integrations with
Koko if the component or API is not listed in this document as it has the
potential to break without any notice.

## Admin API

Koko offers Admin API of Kong in two forms:
- HTTP API
- gRPC API

Both of these are considered public contracts.

Please note that only the user-facing Admin API is considered public for gRPC.
Various sub-systems of Koko communicate using gRPC - these are internal details
and may change over time.

## Configuration file and environment variables

Koko accepts configuration in two forms:
- A JSON or YAML configuration file on disk
- Environment variables (higher precedence than config file)

Both of these are considered public contracts.

## Note on Go code

Koko does not follow module semantic versioning as described by Go modules.
Any function signature or interface defined inside this package is subject to
change at any point without notice.
This policy may change in the future.


## Note on what is not covered

Anything not noted in this document is considered non-public.
Some examples of non-public APIs of Koko (not an exhaustive list):

- Koko's database schema and how database is used
- gRPC service definitions and data-structures used internally in Koko
- Protocol used between Koko and Kong Gateway

## Exceptions

Project maintainers reserve rights to introduce a breaking change for cases
where fixing an security issue is not possible in a backwards-compatible manner.

