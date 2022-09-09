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

func TestExtraProcessing_IsRegexLike(t *testing.T) {
	for _, test := range []string{
		"/blog-\\d+",
		"/[ab]",
		"/(a|b)",
		"/end?",
		"/\\w",
		"/seg<ment>/",
	} {
		require.True(t, isRegexLike(test), "expected isRegexLike(%#v) == true but got false", test)
	}

	for _, test := range []string{
		"/login",
		"/~usnam",
		"/segmented.path",
		"/multi_word",
		"/more-words",
		"/with%sign",
	} {
		require.False(t, isRegexLike(test), "expected isRegexLike(%#v) == false but got true", test)
	}
}

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
			trackedChanges := tracker.Get()
			require.Equal(t, test.expectedTrackedChanges, trackedChanges)
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

func TestExtraProcessing_CorrectRoutesPathFieldPre300(t *testing.T) {
	tests := []struct {
		name                   string
		uncompressedPayload    string
		expectedPayload        string
		expectedTrackedChanges config.TrackedChanges
	}{
		{
			name: "ensure 'paths' is de-normalized when '~' prefix is used in single path",
			uncompressedPayload: `{
				"config_table": {
					"routes": [
						{
							"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							"paths": [ "~/foo" ]
						}
					]
				}
			}`,
			expectedPayload: `{
				"config_table": {
					"routes": [
						{
							"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							"paths": [ "/foo" ]
						}
					]
				}
			}`,
			expectedTrackedChanges: config.TrackedChanges{
				ChangeDetails: []config.ChangeDetail{
					{
						ID: pathRegexFieldChangeID,
						Resources: []config.ResourceInfo{
							{
								Type: "route",
								ID:   "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							},
						},
					},
				},
			},
		},
		{
			name: "non-prefixed path is left alone",
			uncompressedPayload: `{
				"config_table": {
					"routes": [
						{
							"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							"paths": [ "/foo" ]
						}
					]
				}
			}`,
			expectedPayload: `{
				"config_table": {
					"routes": [
						{
							"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							"paths": [ "/foo" ]
						}
					]
				}
			}`,
			expectedTrackedChanges: config.TrackedChanges{},
		},
		{
			name: "mixed with and without '~' prefix",
			uncompressedPayload: `{
				"config_table": {
					"routes": [
						{
							"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							"paths": [ "/foo", "~/bar", "~/baz", "/fum" ]
						}
					]
				}
			}`,
			expectedPayload: `{
				"config_table": {
					"routes": [
						{
							"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							"paths": [ "/foo", "/bar", "/baz", "/fum" ]
						}
					]
				}
			}`,
			expectedTrackedChanges: config.TrackedChanges{
				ChangeDetails: []config.ChangeDetail{
					{
						ID: pathRegexFieldChangeID,
						Resources: []config.ResourceInfo{
							{
								Type: "route",
								ID:   "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							},
						},
					},
				},
			},
		},
		{
			name: "don't denormalize prefixed path",
			uncompressedPayload: `{
				"config_table": {
					"routes": [
						{
							"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							"paths": [ "~/foo/hi%%thing" ]
						}
					]
				}
			}`,
			expectedPayload: `{
				"config_table": {
					"routes": [
						{
							"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							"paths": [ "/foo/hi%%thing" ]
						}
					]
				}
			}`,
			expectedTrackedChanges: config.TrackedChanges{
				ChangeDetails: []config.ChangeDetail{
					{
						ID: pathRegexFieldChangeID,
						Resources: []config.ResourceInfo{
							{
								Type: "route",
								ID:   "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							},
						},
					},
				},
			},
		},
		{
			name: "don't denormalize plain path",
			uncompressedPayload: `{
				"config_table": {
					"routes": [
						{
							"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							"paths": [ "/foo/hi%%thing" ]
						}
					]
				}
			}`,
			expectedPayload: `{
				"config_table": {
					"routes": [
						{
							"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							"paths": [ "/foo/hi%%thing" ]
						}
					]
				}
			}`,
			expectedTrackedChanges: config.TrackedChanges{},
		},
		{
			name: "multiple routes",
			uncompressedPayload: `{
				"config_table": {
					"routes": [
						{
							"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							"paths": [ "~/foo" ]
						},
						{
							"id": "7c242a47-fdd1-4177-b657-20bf60c49a0e",
							"paths": [ "/foo", "/foo/hi%%thing" ]
						},
						{
							"id": "c9eac010-eee2-45a4-b867-c672088389c9",
							"paths": [ "/foo", "~/bar", "~/baz", "/fum" ]
						},
						{
							"id": "ecc628f0-8415-4bf8-b6c9-45ca9327cc52",
							"paths": [ "~/foo/hi%%thing", "/fim%%fum" ]
						},
						{
							"id": "0834e1cf-0f41-49c2-a177-fbc81526535e",
							"paths": [ "/fim.*fum" ]
						},
						{
							"id": "c82b849b-4308-4dad-a963-52cc3ef2a16a",
							"paths": [ "~/fim.*fum" ]
						},
						{
							"id": "0d0e62d5-08ac-4c26-b7cd-8ff0d8a486ef",
							"paths": [ "/fim.*fum", "~/blog-\\d+", "/post-\\w*" ]
						},
						{
							"id": "70e72b38-0430-4049-83ca-9a0f10c146f6",
							"paths": [ "/fim.*fum", "~/blog-\\d+",  "/foo", "/foo/hi%%thing" ]
						}
					]
				}
			}`,
			expectedPayload: `{
				"config_table": {
					"routes": [
						{
							"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							"paths": [ "/foo" ]
						},
						{
							"id": "7c242a47-fdd1-4177-b657-20bf60c49a0e",
							"paths": [ "/foo", "/foo/hi%%thing" ]
						},
						{
							"id": "c9eac010-eee2-45a4-b867-c672088389c9",
							"paths": [ "/foo", "/bar", "/baz", "/fum" ]
						},
						{
							"id": "ecc628f0-8415-4bf8-b6c9-45ca9327cc52",
							"paths": [ "/foo/hi%%thing", "/fim%%fum" ]
						},
						{
							"id": "0834e1cf-0f41-49c2-a177-fbc81526535e",
							"paths": [ "/fim.*fum" ]
						},
						{
							"id": "c82b849b-4308-4dad-a963-52cc3ef2a16a",
							"paths": [ "/fim.*fum" ]
						},
						{
							"id": "0d0e62d5-08ac-4c26-b7cd-8ff0d8a486ef",
							"paths": [ "/fim.*fum", "/blog-\\d+", "/post-\\w*" ]
						},
						{
							"id": "70e72b38-0430-4049-83ca-9a0f10c146f6",
							"paths": [ "/fim.*fum", "/blog-\\d+",  "/foo", "/foo/hi%%thing" ]
						}
					]
				}
			}`,
			expectedTrackedChanges: config.TrackedChanges{
				ChangeDetails: []config.ChangeDetail{
					{
						ID: pathRegexFieldChangeID,
						Resources: []config.ResourceInfo{
							{
								Type: "route",
								ID:   "0d0e62d5-08ac-4c26-b7cd-8ff0d8a486ef",
							},
							{
								Type: "route",
								ID:   "70e72b38-0430-4049-83ca-9a0f10c146f6",
							},
							{
								Type: "route",
								ID:   "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							},
							{
								Type: "route",
								ID:   "c82b849b-4308-4dad-a963-52cc3ef2a16a",
							},
							{
								Type: "route",
								ID:   "c9eac010-eee2-45a4-b867-c672088389c9",
							},
							{
								Type: "route",
								ID:   "ecc628f0-8415-4bf8-b6c9-45ca9327cc52",
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
			processedPayload, err := migrateRoutesPathFieldPre300(test.
				uncompressedPayload, versionsPre300, tracker, log.Logger)
			require.NoError(t, err)
			require.JSONEq(t, test.expectedPayload, processedPayload)
			trackedChanges := tracker.Get()
			require.Equal(t, test.expectedTrackedChanges, trackedChanges)
		})
	}
}

func TestExtraProcessing_checkRoutePaths300AndAbove(t *testing.T) {
	tests := []struct {
		name                   string
		paths                  []any
		expectedErr            string
		expectedTrackedChanges config.TrackedChanges
	}{
		{
			name:                   "ensure 'paths' with '~' prefix pass through",
			paths:                  []any{"~/foo"},
			expectedTrackedChanges: config.TrackedChanges{},
		},
		{
			name:                   "non-prefixed path is left alone if plain",
			paths:                  []any{"/foo"},
			expectedTrackedChanges: config.TrackedChanges{},
		},
		{
			name:                   "mixed with and without '~' prefix",
			paths:                  []any{"/foo", "~/bar", "~/baz", "/fum"},
			expectedTrackedChanges: config.TrackedChanges{},
		},
		{
			name:                   "don't denormalize prefixed path",
			paths:                  []any{"~/foo/hi%%thing"},
			expectedTrackedChanges: config.TrackedChanges{},
		},
		{
			name:                   "don't denormalize plain path",
			paths:                  []any{"/foo/hi%%thing"},
			expectedTrackedChanges: config.TrackedChanges{},
		},
		{
			name:                   "don't denormalize mixed paths",
			paths:                  []any{"~/foo/hi%%thing", "/fim%%fum"},
			expectedTrackedChanges: config.TrackedChanges{},
		},
		{
			name:  "warn on non-prefixed regex-like paths",
			paths: []any{"/fim.*fum"},
			expectedTrackedChanges: config.TrackedChanges{
				ChangeDetails: []config.ChangeDetail{
					{
						ID: pathRegexFieldUnprefixedChangeID,
						Resources: []config.ResourceInfo{
							{
								Type: "route",
								ID:   "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							},
						},
					},
				},
			},
			expectedErr: "found",
		},
		{
			name:                   "silent on prefixed regex-like paths",
			paths:                  []any{"~/fim.*fum"},
			expectedTrackedChanges: config.TrackedChanges{},
		},
		{
			name:  "mixed prefixed and non-prefixed regex-like paths",
			paths: []any{"/fim.*fum", "~/blog-\\d+", "/post-\\w*"},
			expectedTrackedChanges: config.TrackedChanges{
				ChangeDetails: []config.ChangeDetail{
					{
						ID: pathRegexFieldUnprefixedChangeID,
						Resources: []config.ResourceInfo{
							{
								Type: "route",
								ID:   "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							},
						},
					},
				},
			},
			expectedErr: "found",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tracker := config.NewChangeTracker()
			err := checkRoutePaths300AndAbove(test.paths, "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a", tracker)
			if test.expectedErr == "" {
				require.NoError(t, err)
			} else {
				require.Errorf(t, err, test.expectedErr)
			}
			require.Equal(t, test.expectedTrackedChanges, tracker.Get())
		})
	}
}

func TestExtraProcessing_CorrectRoutesPathField300AndAbove(t *testing.T) {
	tests := []struct {
		name                   string
		uncompressedPayload    string
		expectedPayload        string
		expectedTrackedChanges config.TrackedChanges
	}{
		{
			name: "ensure 'paths' with '~' prefix pass through",
			uncompressedPayload: `{
				"config_table": {
					"routes": [
						{
							"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							"paths": [ "~/foo" ]
						}
					]
				}
			}`,
			expectedPayload: `{
				"config_table": {
					"routes": [
						{
							"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							"paths": [ "~/foo" ]
						}
					]
				}
			}`,
			expectedTrackedChanges: config.TrackedChanges{},
		},
		{
			name: "non-prefixed path is left alone if plain",
			uncompressedPayload: `{
				"config_table": {
					"routes": [
						{
							"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							"paths": [ "/foo" ]
						}
					]
				}
			}`,
			expectedPayload: `{
				"config_table": {
					"routes": [
						{
							"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							"paths": [ "/foo" ]
						}
					]
				}
			}`,
			expectedTrackedChanges: config.TrackedChanges{},
		},
		{
			name: "mixed with and without '~' prefix",
			uncompressedPayload: `{
				"config_table": {
					"routes": [
						{
							"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							"paths": [ "/foo", "~/bar", "~/baz", "/fum" ]
						}
					]
				}
			}`,
			expectedPayload: `{
				"config_table": {
					"routes": [
						{
							"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							"paths": [ "/foo", "~/bar", "~/baz", "/fum" ]
						}
					]
				}
			}`,
			expectedTrackedChanges: config.TrackedChanges{},
		},
		{
			name: "don't denormalize prefixed path",
			uncompressedPayload: `{
				"config_table": {
					"routes": [
						{
							"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							"paths": [ "~/foo/hi%%thing" ]
						}
					]
				}
			}`,
			expectedPayload: `{
				"config_table": {
					"routes": [
						{
							"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							"paths": [ "~/foo/hi%%thing" ]
						}
					]
				}
			}`,
			expectedTrackedChanges: config.TrackedChanges{},
		},
		{
			name: "don't denormalize plain path",
			uncompressedPayload: `{
				"config_table": {
					"routes": [
						{
							"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							"paths": [ "/foo/hi%%thing" ]
						}
					]
				}
			}`,
			expectedPayload: `{
				"config_table": {
					"routes": [
						{
							"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							"paths": [ "/foo/hi%%thing" ]
						}
					]
				}
			}`,
			expectedTrackedChanges: config.TrackedChanges{},
		},
		{
			name: "don't denormalize mixed paths",
			uncompressedPayload: `{
				"config_table": {
					"routes": [
						{
							"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							"paths": [ "~/foo/hi%%thing", "/fim%%fum" ]
						}
					]
				}
			}`,
			expectedPayload: `{
				"config_table": {
					"routes": [
						{
							"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							"paths": [ "~/foo/hi%%thing", "/fim%%fum" ]
						}
					]
				}
			}`,
			expectedTrackedChanges: config.TrackedChanges{},
		},
		{
			name: "warn on non-prefixed regex-like paths",
			uncompressedPayload: `{
				"config_table": {
					"routes": [
						{
							"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							"paths": [ "/fim.*fum" ]
						}
					]
				}
			}`,
			expectedPayload: `{
				"config_table": {
					"routes": [
						{
							"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							"paths": [ "/fim.*fum" ]
						}
					]
				}
			}`,
			expectedTrackedChanges: config.TrackedChanges{
				ChangeDetails: []config.ChangeDetail{
					{
						ID: pathRegexFieldUnprefixedChangeID,
						Resources: []config.ResourceInfo{
							{
								Type: "route",
								ID:   "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							},
						},
					},
				},
			},
		},
		{
			name: "silent on prefixed regex-like paths",
			uncompressedPayload: `{
				"config_table": {
					"routes": [
						{
							"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							"paths": [ "~/fim.*fum" ]
						}
					]
				}
			}`,
			expectedPayload: `{
				"config_table": {
					"routes": [
						{
							"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							"paths": [ "~/fim.*fum" ]
						}
					]
				}
			}`,
			expectedTrackedChanges: config.TrackedChanges{},
		},
		{
			name: "mixed prefixed and non-prefixed regex-like paths",
			uncompressedPayload: `{
				"config_table": {
					"routes": [
						{
							"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							"paths": [ "/fim.*fum", "~/blog-\\d+", "/post-\\w*" ]
						}
					]
				}
			}`,
			expectedPayload: `{
				"config_table": {
					"routes": [
						{
							"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							"paths": [ "/fim.*fum", "~/blog-\\d+", "/post-\\w*" ]
						}
					]
				}
			}`,
			expectedTrackedChanges: config.TrackedChanges{
				ChangeDetails: []config.ChangeDetail{
					{
						ID: pathRegexFieldUnprefixedChangeID,
						Resources: []config.ResourceInfo{
							{
								Type: "route",
								ID:   "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							},
						},
					},
				},
			},
		},
		{
			name: "multiple routes",
			uncompressedPayload: `{
				"config_table": {
					"routes": [
						{
							"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							"paths": [ "~/foo" ]
						},
						{
							"id": "7c242a47-fdd1-4177-b657-20bf60c49a0e",
							"paths": [ "/foo", "/foo/hi%%thing" ]
						},
						{
							"id": "c9eac010-eee2-45a4-b867-c672088389c9",
							"paths": [ "/foo", "~/bar", "~/baz", "/fum" ]
						},
						{
							"id": "ecc628f0-8415-4bf8-b6c9-45ca9327cc52",
							"paths": [ "~/foo/hi%%thing", "/fim%%fum" ]
						},
						{
							"id": "0834e1cf-0f41-49c2-a177-fbc81526535e",
							"paths": [ "/fim.*fum" ]
						},
						{
							"id": "c82b849b-4308-4dad-a963-52cc3ef2a16a",
							"paths": [ "~/fim.*fum" ]
						},
						{
							"id": "0d0e62d5-08ac-4c26-b7cd-8ff0d8a486ef",
							"paths": [ "/fim.*fum", "~/blog-\\d+", "/post-\\w*" ]
						},
						{
							"id": "70e72b38-0430-4049-83ca-9a0f10c146f6",
							"paths": [ "/fim.*fum", "~/blog-\\d+",  "/foo", "/foo/hi%%thing" ]
						}
					]
				}
			}`,
			expectedPayload: `{
				"config_table": {
					"routes": [
						{
							"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							"paths": [ "~/foo" ]
						},
						{
							"id": "7c242a47-fdd1-4177-b657-20bf60c49a0e",
							"paths": [ "/foo", "/foo/hi%%thing" ]
						},
						{
							"id": "c9eac010-eee2-45a4-b867-c672088389c9",
							"paths": [ "/foo", "~/bar", "~/baz", "/fum" ]
						},
						{
							"id": "ecc628f0-8415-4bf8-b6c9-45ca9327cc52",
							"paths": [ "~/foo/hi%%thing", "/fim%%fum" ]
						},
						{
							"id": "0834e1cf-0f41-49c2-a177-fbc81526535e",
							"paths": [ "/fim.*fum" ]
						},
						{
							"id": "c82b849b-4308-4dad-a963-52cc3ef2a16a",
							"paths": [ "~/fim.*fum" ]
						},
						{
							"id": "0d0e62d5-08ac-4c26-b7cd-8ff0d8a486ef",
							"paths": [ "/fim.*fum", "~/blog-\\d+", "/post-\\w*" ]
						},
						{
							"id": "70e72b38-0430-4049-83ca-9a0f10c146f6",
							"paths": [ "/fim.*fum", "~/blog-\\d+",  "/foo", "/foo/hi%%thing" ]
						}
					]
				}
			}`,
			expectedTrackedChanges: config.TrackedChanges{
				ChangeDetails: []config.ChangeDetail{
					{
						ID: pathRegexFieldUnprefixedChangeID,
						Resources: []config.ResourceInfo{
							{
								Type: "route",
								ID:   "0834e1cf-0f41-49c2-a177-fbc81526535e",
							},
							{
								Type: "route",
								ID:   "0d0e62d5-08ac-4c26-b7cd-8ff0d8a486ef",
							},
							{
								Type: "route",
								ID:   "70e72b38-0430-4049-83ca-9a0f10c146f6",
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
			processedPayload := checkRoutesPathFieldPost300(test.
				uncompressedPayload, versionsPre300, tracker, log.Logger)
			require.JSONEq(t, test.expectedPayload, processedPayload)
			trackedChanges := tracker.Get()
			require.Equal(t, test.expectedTrackedChanges, trackedChanges)
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

func TestExtraProcessing_EmitCorrectRoutePath(t *testing.T) {
	type expect struct {
		processedPayload string
		trackedChanges   config.TrackedChanges
	}
	tests := []struct {
		name                                string
		uncompressedPayload                 string
		expectedPre300, expected300AndAbove expect
	}{
		// a plaintext path is the same both before and after 3.0; no warnings
		{
			name: "plaintext path",
			uncompressedPayload: `{
				"config_table": {
					"routes": [
						{
							"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							"paths": [ "/foo" ]
						}
					]
				}
			}`,
			expectedPre300: expect{
				processedPayload: `{
					"config_table": {
						"routes": [
							{
								"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
								"paths": [ "/foo" ]
							}
						]
					}
				}`,
				trackedChanges: config.TrackedChanges{},
			},
			expected300AndAbove: expect{
				processedPayload: `{
					"config_table": {
						"_format_version": "3.0",
						"routes": [
							{
								"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
								"paths": [ "/foo" ]
							}
						]
					}
				}`,
				trackedChanges: config.TrackedChanges{},
			},
		},
		// a path that is "regex-like" is passed unchanged
		// but on >= 3.0 generates a change warning
		{
			name: "regex-like path",
			uncompressedPayload: `{
				"config_table": {
					"routes": [
						{
							"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							"paths": [ "/foo-\\d+" ]
						}
					]
				}
			}`,
			expectedPre300: expect{
				processedPayload: `{
					"config_table": {
						"routes": [
							{
								"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
								"paths": [ "/foo-\\d+" ]
							}
						]
					}
				}`,
				trackedChanges: config.TrackedChanges{},
			},
			expected300AndAbove: expect{
				processedPayload: `{
					"config_table": {
						"_format_version": "3.0",
						"routes": [
							{
								"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
								"paths": [ "/foo-\\d+" ]
							}
						]
					}
				}`,
				trackedChanges: config.TrackedChanges{
					ChangeDetails: []config.ChangeDetail{
						{
							ID: pathRegexFieldUnprefixedChangeID,
							Resources: []config.ResourceInfo{
								{
									Type: "route",
									ID:   "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
								},
							},
						},
					},
				},
			},
		},
		// a regex path, prefixed with `~` is back-ported for
		// pre-3.0, and passed unchanged to >= 3.0
		{
			name: "prefixed regex path",
			uncompressedPayload: `{
				"config_table": {
					"routes": [
						{
							"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							"paths": [ "~/foo-\\d+" ]
						}
					]
				}
			}`,
			expectedPre300: expect{
				// NOTE: should backporting be silent?
				processedPayload: `{
					"config_table": {
						"routes": [
							{
								"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
								"paths": [ "/foo-\\d+" ]
							}
						]
					}
				}`,
				trackedChanges: config.TrackedChanges{
					ChangeDetails: []config.ChangeDetail{
						{
							ID: pathRegexFieldChangeID,
							Resources: []config.ResourceInfo{
								{
									Type: "route",
									ID:   "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
								},
							},
						},
					},
				},
			},
			expected300AndAbove: expect{
				processedPayload: `{
					"config_table": {
						"_format_version": "3.0",
						"routes": [
							{
								"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
								"paths": [ "~/foo-\\d+" ]
							}
						]
					}
				}`,
			},
		},
		{
			name: "multiple routes",
			uncompressedPayload: `{
				"config_table": {
					"routes": [
						{
							"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							"paths": [ "~/foo" ]
						},
						{
							"id": "7c242a47-fdd1-4177-b657-20bf60c49a0e",
							"paths": [ "/foo", "/foo/hi%%thing" ]
						},
						{
							"id": "c9eac010-eee2-45a4-b867-c672088389c9",
							"paths": [ "/foo", "~/bar", "~/baz", "/fum" ]
						},
						{
							"id": "ecc628f0-8415-4bf8-b6c9-45ca9327cc52",
							"paths": [ "~/foo/hi%%thing", "/fim%%fum" ]
						},
						{
							"id": "0834e1cf-0f41-49c2-a177-fbc81526535e",
							"paths": [ "/fim.*fum" ]
						},
						{
							"id": "c82b849b-4308-4dad-a963-52cc3ef2a16a",
							"paths": [ "~/fim.*fum" ]
						},
						{
							"id": "0d0e62d5-08ac-4c26-b7cd-8ff0d8a486ef",
							"paths": [ "/fim.*fum", "~/blog-\\d+", "/post-\\w*" ]
						},
						{
							"id": "70e72b38-0430-4049-83ca-9a0f10c146f6",
							"paths": [ "/fim.*fum", "~/blog-\\d+",  "/foo", "/foo/hi%%thing" ]
						}
					]
				}
			}`,
			expectedPre300: expect{
				processedPayload: `{
					"config_table": {
						"routes": [
							{
								"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
								"paths": [ "/foo" ]
							},
							{
								"id": "7c242a47-fdd1-4177-b657-20bf60c49a0e",
								"paths": [ "/foo", "/foo/hi%%thing" ]
							},
							{
								"id": "c9eac010-eee2-45a4-b867-c672088389c9",
								"paths": [ "/foo", "/bar", "/baz", "/fum" ]
							},
							{
								"id": "ecc628f0-8415-4bf8-b6c9-45ca9327cc52",
								"paths": [ "/foo/hi%%thing", "/fim%%fum" ]
							},
							{
								"id": "0834e1cf-0f41-49c2-a177-fbc81526535e",
								"paths": [ "/fim.*fum" ]
							},
							{
								"id": "c82b849b-4308-4dad-a963-52cc3ef2a16a",
								"paths": [ "/fim.*fum" ]
							},
							{
								"id": "0d0e62d5-08ac-4c26-b7cd-8ff0d8a486ef",
								"paths": [ "/fim.*fum", "/blog-\\d+", "/post-\\w*" ]
							},
							{
								"id": "70e72b38-0430-4049-83ca-9a0f10c146f6",
								"paths": [ "/fim.*fum", "/blog-\\d+",  "/foo", "/foo/hi%%thing" ]
							}
						]
					}
				}`,
				trackedChanges: config.TrackedChanges{
					ChangeDetails: []config.ChangeDetail{
						{
							ID: pathRegexFieldChangeID,
							Resources: []config.ResourceInfo{
								{
									Type: "route",
									ID:   "0d0e62d5-08ac-4c26-b7cd-8ff0d8a486ef",
								},
								{
									Type: "route",
									ID:   "70e72b38-0430-4049-83ca-9a0f10c146f6",
								},
								{
									Type: "route",
									ID:   "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
								},
								{
									Type: "route",
									ID:   "c82b849b-4308-4dad-a963-52cc3ef2a16a",
								},
								{
									Type: "route",
									ID:   "c9eac010-eee2-45a4-b867-c672088389c9",
								},
								{
									Type: "route",
									ID:   "ecc628f0-8415-4bf8-b6c9-45ca9327cc52",
								},
							},
						},
					},
				},
			},
			expected300AndAbove: expect{
				processedPayload: `{
					"config_table": {
						"_format_version": "3.0",
						"routes": [
							{
								"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
								"paths": [ "~/foo" ]
							},
							{
								"id": "7c242a47-fdd1-4177-b657-20bf60c49a0e",
								"paths": [ "/foo", "/foo/hi%%thing" ]
							},
							{
								"id": "c9eac010-eee2-45a4-b867-c672088389c9",
								"paths": [ "/foo", "~/bar", "~/baz", "/fum" ]
							},
							{
								"id": "ecc628f0-8415-4bf8-b6c9-45ca9327cc52",
								"paths": [ "~/foo/hi%%thing", "/fim%%fum" ]
							},
							{
								"id": "0834e1cf-0f41-49c2-a177-fbc81526535e",
								"paths": [ "/fim.*fum" ]
							},
							{
								"id": "c82b849b-4308-4dad-a963-52cc3ef2a16a",
								"paths": [ "~/fim.*fum" ]
							},
							{
								"id": "0d0e62d5-08ac-4c26-b7cd-8ff0d8a486ef",
								"paths": [ "/fim.*fum", "~/blog-\\d+", "/post-\\w*" ]
							},
							{
								"id": "70e72b38-0430-4049-83ca-9a0f10c146f6",
								"paths": [ "/fim.*fum", "~/blog-\\d+",  "/foo", "/foo/hi%%thing" ]
							}
						]
					}
				}`,
				trackedChanges: config.TrackedChanges{
					ChangeDetails: []config.ChangeDetail{
						{
							ID: pathRegexFieldUnprefixedChangeID,
							Resources: []config.ResourceInfo{
								{
									Type: "route",
									ID:   "0834e1cf-0f41-49c2-a177-fbc81526535e",
								},
								{
									Type: "route",
									ID:   "0d0e62d5-08ac-4c26-b7cd-8ff0d8a486ef",
								},
								{
									Type: "route",
									ID:   "70e72b38-0430-4049-83ca-9a0f10c146f6",
								},
							},
						},
					},
				},
			},
		},
	}

	pre300Version := versioning.MustNewVersion("2.8.4")
	v300Version := versioning.MustNewVersion("3.0.0")

	for _, test := range tests {
		t.Run(test.name+" - v2.8.4", func(t *testing.T) {
			tracker := config.NewChangeTracker()
			processedPayload, err := VersionCompatibilityExtraProcessing(
				test.uncompressedPayload, pre300Version, tracker, log.Logger)
			require.NoError(t, err)
			require.JSONEq(t, test.expectedPre300.processedPayload, processedPayload)
			require.Equal(t, test.expectedPre300.trackedChanges, tracker.Get())
		})

		t.Run(test.name+" - v3.0.0", func(t *testing.T) {
			tracker := config.NewChangeTracker()
			processedPayload, err := VersionCompatibilityExtraProcessing(
				test.uncompressedPayload, v300Version, tracker, log.Logger)
			require.NoError(t, err)
			require.JSONEq(t, test.expected300AndAbove.processedPayload, processedPayload)
			require.Equal(t, test.expected300AndAbove.trackedChanges, tracker.Get())
		})
	}
}

func TestExtraProcessing_CorrectStatsdIdentifiers(t *testing.T) {
	tests := []struct {
		name                   string
		uncompressedPayload    string
		expectedPayload        string
		expectedTrackedChanges config.TrackedChanges
	}{
		{
			name: "ensure unsupported '*_identifier' fields are removed",
			uncompressedPayload: `{
				"config_table": {
					"plugins": [
						{
							"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							"name": "statsd",
							"config": {
								"metrics": [
									{
										"name": "unique_users",
										"stat_type": "set",
										"consumer_identifier": null,
										"workspace_identifier": null,
										"service_identifier": null
									}
								]
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
							"name": "statsd",
							"config": {
								"metrics": [
									{
										"name": "unique_users",
										"stat_type": "set",
										"consumer_identifier": "custom_id"
									}
								]
							}
						}
					]
				}
			}`,
			expectedTrackedChanges: config.TrackedChanges{
				ChangeDetails: []config.ChangeDetail{
					{
						ID: statsdUnsupportedMetricFieldChangeID,
						Resources: []config.ResourceInfo{
							{
								Type: "plugin",
								ID:   "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							},
						},
					},
					{
						ID: statsdAddDefaultMetricFieldValueChangeID,
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
			name: "ensure null '*_identifier' fields are properly filled with defaults for one metric",
			uncompressedPayload: `{
				"config_table": {
					"plugins": [
						{
							"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							"name": "statsd",
							"config": {
								"metrics": [
									{
										"name": "request_per_user",
										"sample_rate": 1,
										"stat_type": "counter"
									}
								]
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
							"name": "statsd",
							"config": {
								"metrics": [
									{
										"consumer_identifier": "custom_id",
										"name": "request_per_user",
										"sample_rate": 1,
										"stat_type": "counter"
									}
								]
							}
						}
					]
				}
			}`,
			expectedTrackedChanges: config.TrackedChanges{
				ChangeDetails: []config.ChangeDetail{
					{
						ID: statsdAddDefaultMetricFieldValueChangeID,
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
			name: "ensure null '*_identifier' fields are properly filled with defaults for multiple metrics",
			uncompressedPayload: `{
				"config_table": {
					"plugins": [
						{
							"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							"name": "statsd",
							"config": {
								"metrics": [
									{
										"name": "request_per_user",
										"sample_rate": 1,
										"stat_type": "counter"
									},
									{
										"name": "status_count_per_user",
										"sample_rate": 1,
										"stat_type": "counter"
									}
								]
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
							"name": "statsd",
							"config": {
								"metrics": [
									{
										"consumer_identifier": "custom_id",
										"name": "request_per_user",
										"sample_rate": 1,
										"stat_type": "counter"
									},
									{
										"consumer_identifier": "custom_id",
										"name": "status_count_per_user",
										"sample_rate": 1,
										"stat_type": "counter"
									}
								]
							}
						}
					]
				}
			}`,
			expectedTrackedChanges: config.TrackedChanges{
				ChangeDetails: []config.ChangeDetail{
					{
						ID: statsdAddDefaultMetricFieldValueChangeID,
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
			name: "ensure non-default metric doesn't get changed",
			uncompressedPayload: `{
				"config_table": {
					"plugins": [
						{
							"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							"name": "statsd",
							"config": {
								"metrics": [
									{
										"name": "cache_datastore_misses_total",
										"sample_rate": 1,
										"stat_type": "counter"
									}
								]
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
							"name": "statsd",
							"config": {
								"metrics": [
									{
										"name": "cache_datastore_misses_total",
										"sample_rate": 1,
										"stat_type": "counter"
									}
								]
							}
						}
					]
				}
			}`,
		},
		{
			name: "ensure unsupported metrics are removed",
			uncompressedPayload: `{
				"config_table": {
					"plugins": [
						{
							"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							"name": "statsd",
							"config": {
								"metrics": [
									{
										"name": "response_size",
										"stat_type": "timer",
										"service_identifier": "service_name"
									},
									{
										"name": "shdict_usage",
										"stat_type": "gauge",
										"sample_rate": 1,
										"service_identifier": "service_name_or_host"
									},
									{
										"name": "status_count_per_user_per_route",
										"stat_type": "counter",
										"sample_rate": 1,
										"service_identifier": "service_name_or_host",
										"consumer_identifier": "custom_id"
									},
									{
										"workspace_identifier": "workspace_id",
										"name": "status_count_per_workspace",
										"stat_type": "counter",
										"sample_rate": 1
									}
								]
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
							"name": "statsd",
							"config": {
								"metrics": [
									{
										"name": "response_size",
										"stat_type": "timer"
									}
								]
							}
						}
					]
				}
			}`,
			expectedTrackedChanges: config.TrackedChanges{
				ChangeDetails: []config.ChangeDetail{
					{
						ID: statsdUnsupportedMetricChangeID,
						Resources: []config.ResourceInfo{
							{
								Type: "plugin",
								ID:   "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							},
						},
					},
					{
						ID: statsdUnsupportedMetricFieldChangeID,
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
			require.JSONEq(
				t,
				test.expectedPayload,
				correctStatsdMetrics(test.uncompressedPayload, "2.6.0.0", tracker, log.Logger),
			)
			trackedChanged := tracker.Get()
			require.Equal(t, test.expectedTrackedChanges, trackedChanged)
		})
	}
}
