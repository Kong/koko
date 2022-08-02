package compat

import (
	"fmt"
	"testing"

	"github.com/kong/koko/internal/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/sjson"
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
			processedPayload := correctAWSLambdaMutuallyExclusiveFields(test.uncompressedPayload, "< 2.6.0", log.Logger)
			require.JSONEq(t, test.expectedPayload, processedPayload)
		})
	}
}

func Test_correctHTTPLogHeadersField(t *testing.T) {
	tests := []struct {
		name                string
		uncompressedPayload string
		expectedPayload     string
	}{
		{
			name: "non http-log plugin",
			uncompressedPayload: `{
				"name": "another-plugin",
				"config": {
					"headers": {
						"header-1": "value-1"
					}
				}
			}`,
			expectedPayload: `{
				"name": "another-plugin",
				"config": {
					"headers": {
						"header-1": "value-1"
					}
				}
			}`,
		},
		{
			name: "null headers",
			uncompressedPayload: `{
				"name": "http-log",
				"config": {
					"headers": null
				}
			}`,
			expectedPayload: `{
				"name": "http-log",
				"config": {
					"headers": null
				}
			}`,
		},
		{
			name: "empty headers",
			uncompressedPayload: `{
				"name": "http-log",
				"config": {
					"headers": ""
				}
			}`,
			expectedPayload: `{
				"name": "http-log",
				"config": {
					"headers": null
				}
			}`,
		},
		{
			name: "single header",
			uncompressedPayload: `{
				"name": "http-log",
				"config": {
					"headers": {
						"header-1": "value-1"
					}
				}
			}`,
			expectedPayload: `{
				"name": "http-log",
				"config": {
					"headers": {
						"header-1": ["value-1"]
					}
				}
			}`,
		},
		{
			name: "multiple headers",
			uncompressedPayload: `{
				"name": "http-log",
				"config": {
					"headers": {
						"header-1": "",
						"header-2": "value-1",
						"header-3": "value-2"
					}
				}
			}`,
			expectedPayload: `{
				"name": "http-log",
				"config": {
					"headers": {
						"header-2": ["value-1"],
						"header-3": ["value-2"]
					}
				}
			}`,
		},
	}
	for _, tt := range tests {
		// Duplicate plugin to ensure updating multiple JSON objects work.
		require.NoError(t, repeatJSONObject(
			"config_table.plugins",
			2,
			&tt.uncompressedPayload,
			&tt.expectedPayload,
		))

		t.Run(tt.name, func(t *testing.T) {
			actual, err := correctHTTPLogHeadersField(tt.uncompressedPayload)
			require.NoError(t, err)
			assert.JSONEq(t, tt.expectedPayload, actual)
		})
	}
}

// repeatJSONObject allows you to pass in raw JSON messages (in `jsonData`) and
// repeat that object `count` times, for the given JSON `key` path. It will replace
// the `jsonData` pointer values with the newly generated JSON object.
//
// The provided JSON key path must be an array, and said JSON `key` path should not
// exist in `jsonData`.
func repeatJSONObject(key string, count int, jsonData ...*string) error {
	for _, data := range jsonData {
		newData := "{}"
		for i := 0; i < count; i++ {
			var err error
			if newData, err = sjson.SetRaw(newData, fmt.Sprintf("%s.%d", key, i), *data); err != nil {
				return err
			}
		}
		*data = newData
	}

	return nil
}
