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

	payload, err := UncompressPayload(res)
	require.Nil(t, err)
	require.Equal(t, string(payload), "{\"config_table\":{\"plugins\":[\"test\"]},\"type\":\"reconfigure\"}\n")
}
