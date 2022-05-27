# Version negotiation service

The version negotiation service is used to get the CP and DP agree on the set
of services and versions available for use by both nodes.

As soon as a DP establishes a connection, it must call the `NegotiateServices()`
method with its own information (in the `node` field) and the list of requested
services.

The DP must negotiate services on each wRPC connection, even in the case of
re-connections.

The response from the CP must contain only the `message` field in the error
case. For a successful response, all other fields must be set.

Refer to [version-negotiation.md](version-negotiation.md) for details.
