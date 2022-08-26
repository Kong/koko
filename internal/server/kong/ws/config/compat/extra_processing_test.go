package compat

import (
	"fmt"
	"testing"

	"github.com/kong/koko/internal/log"
	"github.com/kong/koko/internal/server/kong/ws/config"
	"github.com/kong/koko/internal/versioning"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/sjson"
)

func TestExtraProcessing_CorrectAWSLambdaMutuallyExclusiveFields(t *testing.T) {
	tests := []struct {
		name                   string
		uncompressedPayload    string
		expectedPayload        string
		expectedTrackedChanges config.TrackedChanges
	}{
		{
			name: "ensure 'host' is dropped when both 'aws_region' and 'host' are set",
			uncompressedPayload: `{
				"config_table": {
					"plugins": [
						{
							"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							"name": "aws-lambda",
							"config": {
								"aws_region": "test",
								"host": "192.168.1.1"
							}
						},
						{
							"name": "plugin",
							"config": {
								"object": {
									"key": "value"
								}
							}
						}
					]
				}
			}`,
			expectedPayload: `{
				"config_table": {
					"plugins": [
						{
							"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							"name": "aws-lambda",
							"config": {
								"aws_region": "test"
							}
						},
						{
							"name": "plugin",
							"config": {
								"object": {
									"key": "value"
								}
							}
						}
					]
				}
			}`,
			expectedTrackedChanges: config.TrackedChanges{
				ChangeDetails: []config.ChangeDetail{
					{
						ID: awsLambdaExclusiveFieldChangeID,
						Resources: []config.ResourceInfo{
							{
								Type: "plugin",
								ID:   "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							},
						},
					},
				},
			},
		},
		{
			name: "ensure 'host' is not dropped when 'aws_region' is not set",
			uncompressedPayload: `{
				"config_table": {
					"plugins": [
						{
							"name": "plugin",
							"config": {
								"object": {
									"key": "value"
								}
							}
						},
						{
							"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
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
							"name": "plugin",
							"config": {
								"object": {
									"key": "value"
								}
							}
						},
						{
							"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
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
							"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							"name": "aws-lambda",
							"config": {
								"aws_region": "test"
							}
						},
						{
							"name": "plugin",
							"config": {
								"object": {
									"key": "value"
								}
							}
						}
					]
				}
			}`,
			expectedPayload: `{
				"config_table": {
					"plugins": [
						{
							"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							"name": "aws-lambda",
							"config": {
								"aws_region": "test"
							}
						},
						{
							"name": "plugin",
							"config": {
								"object": {
									"key": "value"
								}
							}
						}
					]
				}
			}`,
		},
		{
			name: "ensure 'host' is dropped when 'aws_region' is set for multiple configured 'aws-lambda' plugins",
			uncompressedPayload: `{
				"config_table": {
					"plugins": [
						{
							"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							"name": "aws-lambda",
							"config": {
								"aws_region": "test",
								"host": "192.168.1.1"
							}
						},
						{
							"name": "plugin",
							"config": {
								"object": {
									"key": "value"
								}
							}
						},
						{
							"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							"name": "aws-lambda",
							"config": {
								"aws_region": "test",
								"host": "192.168.1.1"
							}
						},
						{
							"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
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
							"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							"name": "aws-lambda",
							"config": {
								"aws_region": "test"
							}
						},
						{
							"name": "plugin",
							"config": {
								"object": {
									"key": "value"
								}
							}
						},
						{
							"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							"name": "aws-lambda",
							"config": {
								"aws_region": "test"
							}
						},
						{
							"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							"name": "aws-lambda",
							"config": {
								"aws_region": "test"
							}
						}
					]
				}
			}`,
			expectedTrackedChanges: config.TrackedChanges{
				ChangeDetails: []config.ChangeDetail{
					{
						ID: awsLambdaExclusiveFieldChangeID,
						Resources: []config.ResourceInfo{
							{
								Type: "plugin",
								ID:   "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							},
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tracker := config.NewChangeTracker()
			processedPayload := correctAWSLambdaMutuallyExclusiveFields(test.
				uncompressedPayload, versionsPre260, tracker, log.Logger)
			require.JSONEq(t, test.expectedPayload, processedPayload)
			trackedChanged := tracker.Get()
			require.Equal(t, test.expectedTrackedChanges, trackedChanged)
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

func TestExtraProcessing_CorrectRoutesPathField(t *testing.T) {
	tests := []struct {
		name                   string
		uncompressedPayload    string
		expectedPayload        string
		expectedTrackedChanges config.TrackedChanges
	}{
		{
			name: "ensure 'paths' is de-normalized when '~' prefix is used in single path",
			uncompressedPayload: `{
				"config_table": [
					"routes": [
						{
							"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							paths = [
								"~/foo"
							],
						}
					]
				]
			}`,
			expectedPayload: `{
				"config_table": [
					"routes": [
						{
							"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							paths = [
								"~/foo"
							],
						}
					]
				]
			}`,
			expectedTrackedChanges: config.TrackedChanges{
				ChangeDetails: []config.ChangeDetail{
					{
						ID: pathRegexFieldChangeID,
						Resources: []config.ResourceInfo{
							{
								Type: "entity",
								ID:   "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							},
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tracker := config.NewChangeTracker()
			processedPayload := correctRoutesPathField(test.
				uncompressedPayload, versionsPre300, tracker, log.Logger)
			fmt.Println(processedPayload)
			// require.JSONEq(t, test.expectedPayload, processedPayload)
			// trackedChanged := tracker.Get()
			// require.Equal(t, test.expectedTrackedChanges, trackedChanged)
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

// TestExtraProcessing_EnsureExtraProcessing ensures the
// VersionCompatibilityExtraProcessing function works as expected for
// both OSS and EE version formats.
func TestExtraProcessing_EnsureExtraProcessing(t *testing.T) {
	tests := []struct {
		name                string
		uncompressedPayload string
		expectedPayload     string
	}{
		{
			name: "ensure 'host' is dropped when both 'aws_region' and 'host' are set for aws-lambda plugin",
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
			name: "correct http-log headers for a multiple header",
			uncompressedPayload: `{
				"config_table": {
					"plugins": [
						{
							"name": "http-log",
							"config": {
								"headers": {
									"header-1": "",
									"header-2": "value-1",
									"header-3": "value-2"
								}
							}
						}
					]
				}
			}`,
			expectedPayload: `{
				"config_table": {
					"plugins": [
						{
							"name": "http-log",
							"config": {
								"headers": {
									"header-2": ["value-1"],
									"header-3": ["value-2"]
								}
							}
						}
					]
				}
			}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			for _, version := range []string{"2.6.0", "2.6.0.0"} {
				tracker := config.NewChangeTracker()
				dataPlaneVersion := versioning.MustNewVersion(version)
				processedPayload, err := VersionCompatibilityExtraProcessing(test.uncompressedPayload, dataPlaneVersion,
					tracker, log.Logger)
				require.NoError(t, err)
				require.JSONEq(t, test.expectedPayload, processedPayload)
				// TODO(hbagdi): add code and assertions for tracked changes
				// in extra processors
			}
		})
	}
}
