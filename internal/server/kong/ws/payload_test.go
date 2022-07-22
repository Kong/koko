package ws

import (
	"context"
	"testing"

	"github.com/kong/koko/internal/log"
	"github.com/kong/koko/internal/server/kong/ws/config"
	"github.com/stretchr/testify/require"
)

func TestConfigPayload_Cache(t *testing.T) {
	wsvc, err := config.NewVersionCompatibilityProcessor(config.VersionCompatibilityOpts{
		Logger:        log.Logger,
		KongCPVersion: "2.8.0",
	})
	require.Nil(t, err)
	err = wsvc.AddConfigTableUpdates(map[uint64][]config.ConfigTableUpdates{
		2007999999: {
			{
				Name: "plugin_1",
				Type: config.Plugin,
				RemoveFields: []string{
					"plugin_1_field_1",
				},
			},
		},
	}, false)
	require.Nil(t, err)

	payloadBytes := []byte(`{
		"config_table": {
			"plugins": [
				{
					"name": "plugin_1",
					"config": {
						"plugin_1_field_1": "element_1",
						"plugin_1_field_2": "element_2"
					}
				},
				{
					"name": "plugin_2",
					"config": {
						"plugin_2_field_1": "element_1"
					}
				}
			]
		}
	}`)
	compressedPayload, err := config.CompressPayload(payloadBytes)
	require.Nil(t, err)
	expectedPayloadBytes270 := []byte(`{
		"config_table": {
			"plugins": [
				{
					"name": "plugin_1",
					"config": {
						"plugin_1_field_2": "element_2"
					}
				},
				{
					"name": "plugin_2",
					"config": {
						"plugin_2_field_1": "element_1"
					}
				}
			]
		}
	}`)
	expectedPayload270, err := config.CompressPayload(expectedPayloadBytes270)
	require.Nil(t, err)

	t.Run("ensure payload can be retrieved for single version", func(t *testing.T) {
		payload, err := NewPayload(PayloadOpts{
			VersionCompatibilityProcessor: wsvc,
		})
		require.Nil(t, err)
		err = payload.UpdateBinary(context.Background(), config.Content{
			CompressedPayload: compressedPayload,
			Hash:              "1133ae8be08017e5460160635daa22f2",
		})
		require.Nil(t, err)

		updatedPayload, err := payload.Payload(context.Background(), "2.8.0")
		require.Nil(t, err)
		require.Equal(t, compressedPayload, updatedPayload.CompressedPayload)
		entry, err := payload.configCache.load("2.8.0")
		require.NoError(t, err)
		require.Greater(t, len(entry.CompressedPayload), 0)
		require.NoError(t, entry.Error)
	})

	t.Run("ensure payload can be retrieved using multiple versions", func(t *testing.T) {
		payload, err := NewPayload(PayloadOpts{
			VersionCompatibilityProcessor: wsvc,
		})
		require.Nil(t, err)
		err = payload.UpdateBinary(context.Background(), config.Content{
			CompressedPayload: compressedPayload,
			Hash:              "1133ae8be08017e5460160635daa22f2",
		})
		require.Nil(t, err)

		updatedPayload, err := payload.Payload(context.Background(), "2.8.0")
		require.Nil(t, err)
		require.Equal(t, compressedPayload, updatedPayload.CompressedPayload)
		_, found := payload.configCache["2.8.0"]
		require.True(t, found)
		require.Greater(t, len(payload.configCache["2.8.0"].CompressedPayload), 0)
		require.Nil(t, payload.configCache["2.8.0"].Error)

		updatedPayload, err = payload.Payload(context.Background(), "2.7.0")
		require.Nil(t, err)
		require.Equal(t, expectedPayload270, updatedPayload.CompressedPayload)
		entry, err := payload.configCache.load("2.7.0")
		require.NoError(t, err)
		require.Greater(t, len(entry.CompressedPayload), 0)
		require.NoError(t, entry.Error)
	})

	t.Run("ensure payload configCache is cleared when updated", func(t *testing.T) {
		payload, err := NewPayload(PayloadOpts{
			VersionCompatibilityProcessor: wsvc,
		})
		require.Nil(t, err)
		err = payload.UpdateBinary(context.Background(), config.Content{
			CompressedPayload: compressedPayload,
			Hash:              "1133ae8be08017e5460160635daa22f2",
		})
		require.Nil(t, err)

		updatedPayload, err := payload.Payload(context.Background(), "2.8.0")
		require.Nil(t, err)
		require.Equal(t, compressedPayload, updatedPayload.CompressedPayload)
		entry, err := payload.configCache.load("2.8.0")
		require.NoError(t, err)
		require.Greater(t, len(entry.CompressedPayload), 0)
		require.NoError(t, entry.Error)

		err = payload.UpdateBinary(context.Background(), config.Content{
			CompressedPayload: compressedPayload,
			Hash:              "1133ae8be08017e5460160635daa22f2",
		})
		_, err = payload.configCache.load("2.8.0")
		require.ErrorIs(t, err, errNotFound)
	})
}
