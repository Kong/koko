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

	payload, err := UncompressPayload(res.Payload)
	require.Nil(t, err)
	require.JSONEq(t,
		`{
			"type":"reconfigure",
			"config_table":{"plugins":["test"]},
			"config_hash":"1133ae8be08017e5460160635daa22f2"
		}`,
		string(payload))
}
