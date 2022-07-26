package compat

import (
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"go.uber.org/zap"
)

const (
	dataPlaneVersion3000 = 3000000000
)

func correctAWSLambdaMutuallyExclusiveFields(payload string, dataPlaneVersion uint64, logger *zap.Logger) string {
	pluginName := "aws-lambda"
	processedPayload := payload
	results := gjson.Get(processedPayload, "config_table.plugins.#(name=aws-lambda)#")
	indexUpdate := 0
	for _, res := range results.Array() {
		updatedRaw := res.Raw
		awsRegionResult := gjson.Get(res.Raw, "config.aws_region")
		hostResult := gjson.Get(res.Raw, "config.host")
		// 'aws_region' and 'host' were mutually exclusive fields until
		// Kong version 2.8 but both are accepted in 3.x. If both are set
		// with DPs < 3.x, we decide to drop the 'host' field in order
		// to prevent a failure in the DP.
		if awsRegionResult.Exists() && hostResult.Exists() {
			var err error
			if updatedRaw, err = sjson.Delete(updatedRaw, "config.host"); err != nil {
				logger.With(zap.String("plugin", pluginName)).
					With(zap.String("field", "host")).
					With(zap.Uint64("data-plane", dataPlaneVersion)).
					With(zap.Error(err)).
					Error("plugin configuration field was not removed from configuration")
			} else {
				logger.With(zap.String("plugin", pluginName)).
					With(zap.String("field", "host")).
					With(zap.Uint64("data-plane", dataPlaneVersion)).
					Warn("removing plugin configuration field which is incompatible with data plane")
			}
		}

		// Update the processed payload
		resIndex := res.Index - indexUpdate
		updatedPayload := processedPayload[:resIndex] + updatedRaw +
			processedPayload[resIndex+len(res.Raw):]
		indexUpdate = len(processedPayload) - len(updatedPayload)
		processedPayload = updatedPayload
	}
	return processedPayload
}

func VersionCompatibilityExtraProcessing(payload string, dataPlaneVersion uint64, isEnterprise bool,
	logger *zap.Logger,
) (string, error) {
	processedPayload := payload

	if dataPlaneVersion < dataPlaneVersion3000 {
		// 'aws_region' and 'host' are mutually exclusive for DP < 3.x
		processedPayload = correctAWSLambdaMutuallyExclusiveFields(processedPayload, dataPlaneVersion, logger)
	}

	return processedPayload, nil
}
