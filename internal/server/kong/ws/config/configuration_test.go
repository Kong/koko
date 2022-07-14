package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReconfigurePayload(t *testing.T) {
	var configTable DataPlaneConfig = map[string]interface{}{}
	configTable["plugins"] = []string{"test"}

	res, err := ReconfigurePayload(configTable)
	require.Nil(t, err)

	payload, err := UncompressPayload(res.CompressedPayload)
	require.Nil(t, err)
	require.JSONEq(t,
		`{
			"type": "reconfigure",
			"config_table": {"plugins":["test"]},
			"config_hash": "c5e45a8c77675db925e0ac92d854410b",
			"hashes": {
				"plugins": "23235a224da3cb921fc8722198f0e76a",
				"routes": "00000000000000000000000000000000",
				"services": "00000000000000000000000000000000",
				"targets": "00000000000000000000000000000000",
				"upstreams": "00000000000000000000000000000000"
			}
		}`,
		string(payload))
}
