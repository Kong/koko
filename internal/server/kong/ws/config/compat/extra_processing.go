package compat

import (
	"fmt"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"go.uber.org/zap"
)

const (
	dataPlaneVersion3000 = 3000000000
)

func VersionCompatibilityExtraProcessing(payload string, dataPlaneVersion uint64, isEnterprise bool,
	logger *zap.Logger,
) (string, error) {
	processedPayload := payload

	// opentelemetry plugin was not available before 3.0.0
	if dataPlaneVersion < dataPlaneVersion3000 {
		processedPayload = removePlugin(processedPayload, "opentelemetry", dataPlaneVersion, logger)
	}

	return processedPayload, nil
}

func removePlugin(uncompressedPayload string, pluginName string, dataPlaneVersion uint64, logger *zap.Logger) string {
	processedPayload := uncompressedPayload
	if gjson.Get(processedPayload, fmt.Sprintf("config_table.plugins.#(name=%s)#", pluginName)).Exists() {
		plugins := gjson.Get(processedPayload, "config_table.plugins")
		if plugins.IsArray() {
			removeCount := 0
			for i, res := range plugins.Array() {
				pluginCondition := fmt.Sprintf("..#(name=%s)", pluginName)
				if gjson.Get(res.Raw, pluginCondition).Exists() {
					var err error
					pluginDelete := fmt.Sprintf("config_table.plugins.%d", i-removeCount)
					if processedPayload, err = sjson.Delete(processedPayload, pluginDelete); err != nil {
						logger.With(zap.String("plugin", pluginName)).
							With(zap.Uint64("data-plane", dataPlaneVersion)).
							With(zap.Error(err)).
							Error("plugin was not removed from configuration")
					} else {
						logger.With(zap.String("plugin", pluginName)).
							With(zap.Uint64("data-plane", dataPlaneVersion)).
							Warn("removing plugin which is incompatible with data plane")
						removeCount++
					}
				}
			}
		}
	}
	return processedPayload
}
