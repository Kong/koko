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
			"config_hash": "4fe09342df7064a94fe09342df7064a9",
			"hashes": {
				"plugins": "5e2a17eac1fa241e5e2a17eac1fa241e",
				"routes": "00000000000000000000000000000000",
				"services": "00000000000000000000000000000000",
				"targets": "00000000000000000000000000000000",
				"upstreams": "00000000000000000000000000000000"
			}
		}`,
		string(payload))
}
