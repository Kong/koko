package compat

import (
	"testing"

	"github.com/kong/koko/internal/log"
	"github.com/stretchr/testify/require"
)

func TestExtraProcessing_CorrectAWSLambdaMutuallyExclusiveFields(t *testing.T) {
	tests := []struct {
		name                string
		uncompressedPayload string
		expectedPayload     string
	}{
		{
			name: "ensure 'host' is dropped when both 'aws_region' and 'host' are set",
			uncompressedPayload: `{
				"config_table": {
					"plugins": [
						{
							"name": "aws-lambda",
							"config": {
								"aws_region": "test",
								"host": "192.168.1.1"
							}
						}
					]
				}
			}`,
			expectedPayload: `{
				"config_table": {
					"plugins": [
						{
							"name": "aws-lambda",
							"config": {
								"aws_region": "test"
							}
						}
					]
				}
			}`,
		},
		{
			name: "ensure 'host' is not dropped when 'aws_region' is not set",
			uncompressedPayload: `{
				"config_table": {
					"plugins": [
						{
							"name": "aws-lambda",
							"config": {
								"host": "192.168.1.1"
							}
						}
					]
				}
			}`,
			expectedPayload: `{
				"config_table": {
					"plugins": [
						{
							"name": "aws-lambda",
							"config": {
								"host": "192.168.1.1"
							}
						}
					]
				}
			}`,
		},
		{
			name: "ensure 'aws_region' is not dropped when 'host' is not set",
			uncompressedPayload: `{
				"config_table": {
					"plugins": [
						{
							"name": "aws-lambda",
							"config": {
								"aws_region": "test"
							}
						}
					]
				}
			}`,
			expectedPayload: `{
				"config_table": {
					"plugins": [
						{
							"name": "aws-lambda",
							"config": {
								"aws_region": "test"
							}
						}
					]
				}
			}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			processedPayload := correctAWSLambdaMutuallyExclusiveFields(test.uncompressedPayload, 2006000000, log.Logger)
			require.JSONEq(t, test.expectedPayload, processedPayload)
		})
	}
}
