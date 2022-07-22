# Version negotiation

## Background

Building out a data-plane (DP) and control-plane (CP) communication spec for
Kong will be an iterative effort. It is unlikely that we will solve our
problems in the right way in the first few versions. To help with this, DPs and
CPs will communicate on versioned protocols. Negotiation must happen under the
hood without any involvement from the user and protocol versioning details
should be hidden away from users as much as possible.

Negotiation flow described below MUST be the first thing a DP performs as part
of its startup. Caching of negotiated version across restarts of DP are NOT
allowed. The behavior of DP using a protocol that has not been negotiated is
undefined.

## Negotiation Flow

The `NegotiationService` service is implemented as single wRPC call
`NegotiateServices`, which MUST be the first method call from the DP.

  1. Request data:
      - `node`: a structure containing DP node level metadata with the
        following fields:
        - `id`: a string value containing the node ID of the DP. This value must
          be unique across all nodes of a cluster on a best-effort basis.
        - `type`: an identifier that recognizes the type of the DP node.  This
          MUST be set to "KONG" for Kong Gateway nodes. CP MUST verify if this
          is set.
        - `version`: an opaque string that describes the version of the DP. This
          is same as Kong version for the Kong Gateway DP.
        - `hostname`: hostname of the underlying node of the DP.  This field is
          optional and is meant to be used for debugging purposes.
      - `services_requested`: a repeated list of services being
        requested by the DP. Each object within the array with the
        following keys:
        - `name`: an opaque string containing the name of the service the DP
          intends to use.
        - `versions`: a list of strings describing the versions that the
          DP understands for the service specified in `name` field. This field
          must contain at least one version.
      - `metadata`: optional field that may contain any additional data required
        for the initial handshake.

  2. Response data:
      - CP MUST validate the content of the request and ensure that required
        fields are set in the request. If not, CP MUST respond back only setting
        the `message` field with an opaque string describing the error in a human
        readable form
      - If the request is valid, CP should respond back with the following data:
        - `node`: an object containing CP node level metadata:
          - `id`: a string value containing the node ID of the CP.
        - `services_accepted`: a list comprising of accepted services. Each
          object in the array contains the following fields:
          - `name`: string containing the name of the service
          - `version`: string containing version that the CP accepted from the
            list that the DP proposed
          - `message`: an optional human-readable message that could denote info
            or warning messages like deprecations, upgrade prompts, etc.
        - `services_rejected`: a list comprising of rejected services. Each
          object in the array contains the following fields:
          - `name`: string containing the name of the service
          - `message`: string containing the reason for the rejection.

  3. Once the response is received by DP, DP can proceed with using a service on
    the selected protocol. It is not valid for a DP to use a protocol that has
    not been negotiated with a CP.

### Examples

Here is an example of how a valid request-response flow could look like:
```
NegotiateService ({
  "node": {
    "id": "42",
    "version": "2.6.1-beta",
    "type": "KONG"
  },
  "services_requested": [
    {
      "name": "configuration",
      "versions": [
        "v1",
        "v2"
      ]
    },
    {
      "name": "vitals",
      "versions": [
        "v1",
        "v2"
      ]
    }
  ]
})

<==
{
  "node": {
    "id": "4242",
  },
  "services_accepted": [
    {
      "name": "configuration",
      "versions": "v2"
    }
  ],
  "services_rejected": [
    {
      "name": "vitals",
      "message": "only v3 is available"
    }
  ]
}
```

## Rationale

This section explains the rationale behind the above design.

### Protocol

It's important to support many kinds of load balancers and proxies in front of
the CP nodes.  To avoid any posibility of negotiating with one CP node and
then connect to a different node, it was decided to include the negotiation
in the same connection, with the requirement that it MUST be the first method
call upon connection.

### Request structure

As we evolve Kong (and Konnect at large), some level of compartmentalization within the
software as well as the organization of people is expected. It is unlikely that
a single protocol definition will scale with the required velocity of change
that will be expected as Kong Inc grows.
Furthermore, with multiple teams maintaining different feature areas, it could be
beneficial for teams to have control over the protocol for specific feature areas.
This imparts agency to teams to evolve and roll out updates to the protocol as
required (within certain bounds).

The above negotiation workflow helps compartmentalize features into services,
which then can be iterated on in semi-independent fashion.

## Alternatives considered

### HTTP Protocol

Version negotiating could use an HTTP request and not a wRPC request.
This gives flexibility to change the RPC protocol in future iterations.
It is expected that all the layers below and including HTTP will not change,
but layers above HTTP (like wRPC) may change in future.

### ALPN in TLS for protocol negotiation

This was avoided to guard against cases where an L7 proxy is being used between
the DP and CP. The proxy could drop the ALPN as part of connection setup.

## Questions raised and answered

Is TLS required for version negotiation or not? If yes, how would this work
when mTLS is being used for authentication purposes?

Answer:
- Yes, TLS is required for integrity and confidentiality.
- Since authentication happens before negotiation, mTLS or any other auth scheme
  will be compatible. Details are noted in [authentication.md](authentication.md).

Are there security concerns since this could be used to leak version information?

Answer:
This question is no longer applicable since authentication will be performed
for version negotiation requests.

Are there other security concerns here with the request being unauthenticated?
DoS attacks on the CP?

Answer:
This question is no longer applicable since authentication will be performed
before version negotiation.

Because the request is unauthenticated, there is no way to have version
selection based on more specific metadata. For example, negotiation of version
for specific DPs for a customer or for a specific customer is not possible in
Konnect. This would be very useful for alpha/beta testing of our software.
Can this be addressed in any way?

Answer:
This question is no longer applicable since authentication will be performed
before version negotiation.

Why can't we use `host_id` field instead of a `hostname` to identify nodes?
Specifically from Fero:
Since it is possible for a hostname to be the same (whether on the same
network or not) WDYT about adding an additional field called host_id?
I feel this is safer than exposing the MAC address of the NIC for
potential spoofing.
- BSD uses /etc/hostid and smbios.system.uuid as a fallback
- Linux uses /etc/machine-id or /var/lib/dbus/machine-id
- OS X uses IOPlatformUUID
- Windows uses the MachineGuid from HKEY_LOCAL_MACHINE\SOFTWARE\Microsoft\Cryptography
Instead of host_id we can always have a just a identifier field that is filled
in by the user to help with debugging purposes in the instances where a hostname
is not distinct.


Answer:
`metadata` block exists for this reason.
Implementations are free to specify whatever they wish in there and a field
like `host_id` may be added there.
Hostname is used because:
- its widespread understanding in the industry
- Our existing products already use it and customers are familiar in terms of
  how our product use it
- hostname is usually set by a human and often a human-readable string which
  helps with debugging, host_id is autogen and is of less utility from that
  standpoint

host_id could be added in a future iteration if multiple systems want to rely on
it.
