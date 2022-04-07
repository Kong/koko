package config

import (
	"fmt"
	"strings"
	"testing"

	"github.com/kong/koko/internal/log"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/sjson"
	"go.uber.org/zap"
)

func TestVersionCompatibility_NewVersionCompatibilityProcessor(t *testing.T) {
	t.Run("ensure logger is present", func(t *testing.T) {
		_, err := NewVersionCompatibilityProcessor(VersionCompatibilityOpts{
			KongCPVersion: "2.8.0",
		})
		require.NotNil(t, err)
		require.EqualError(t, err, "opts.Logger required")
	})

	t.Run("ensure Kong Gateway control plane version is present", func(t *testing.T) {
		_, err := NewVersionCompatibilityProcessor(VersionCompatibilityOpts{
			Logger: log.Logger,
		})
		require.NotNil(t, err)
		require.EqualError(t, err, "opts.KongCPVersion required")
	})

	t.Run("ensure Kong Gateway control plane version is valid", func(t *testing.T) {
		tests := []struct {
			kongCPVersion string
			wantsErr      bool
		}{
			{kongCPVersion: "0.3.3"},
			{kongCPVersion: "1.5.0"},
			{kongCPVersion: "2.3.0"},
			{kongCPVersion: "2.3.1"},
			{kongCPVersion: "2.3.111"},
			{kongCPVersion: "2.4.0"},
			{kongCPVersion: "2.5.0"},
			{kongCPVersion: "2.6.0"},
			{kongCPVersion: "2.7.0"},
			{kongCPVersion: "2.8.0-rc1"},
			{kongCPVersion: "2.8.0-beta1"},
			{kongCPVersion: "2.8.0-alpha1"},
			{kongCPVersion: "2.8.0"},
			{kongCPVersion: "2.3.0.2-enterprise-edition"},
			{kongCPVersion: "2.3.0.0-any-suffix-is-valid"},
			{
				kongCPVersion: "2.3.3.",
				wantsErr:      true,
			},
			{
				kongCPVersion: "two.three.zero",
				wantsErr:      true,
			},
		}

		for _, test := range tests {
			_, err := NewVersionCompatibilityProcessor(VersionCompatibilityOpts{
				Logger:        log.Logger,
				KongCPVersion: test.kongCPVersion,
			})
			if test.wantsErr {
				require.NotNil(t, err)
				require.True(t, strings.Contains(err.Error(), "unable to parse opts.KongCPVersion"))
			} else {
				require.Nil(t, err)
			}
		}
	})
}

func TestVersionCompatibility_ParseSemanticVersion(t *testing.T) {
	tests := []struct {
		versionStr      string
		wantsErr        bool
		expectedErr     string
		expectedVersion uint64
	}{
		{
			versionStr:      "0.33.3",
			expectedVersion: 33003000,
		},
		{
			versionStr:      "0.33.3-3-enterprise-edition",
			expectedVersion: 33003003,
		},
		{
			versionStr:      "0.33.3-3-enterprise",
			expectedVersion: 33003003,
		},
		{
			// go-kong won't parse build without suffix containing enterprise
			versionStr:      "0.33.3-3-build-will-not-be-parsed",
			expectedVersion: 33003000,
		},
		{
			versionStr:      "2.3.3.2",
			expectedVersion: 2003003000,
		},
		{
			versionStr:      "2.3.2",
			expectedVersion: 2003002000,
		},
		{
			versionStr:      "2.3.2-rc1",
			expectedVersion: 2003002000,
		},
		{
			versionStr:      "2.3.3-alpha",
			expectedVersion: 2003003000,
		},
		{
			versionStr:      "2.3.4-beta1",
			expectedVersion: 2003004000,
		},
		{
			versionStr:      "2.3.3.2-enterprise-edition",
			expectedVersion: 2003003002,
		},
		{
			versionStr:  "two.three.four",
			wantsErr:    true,
			expectedErr: "unknown Kong version : 'two.three.four'",
		},
		{
			versionStr:  "2.1234.1",
			wantsErr:    true,
			expectedErr: "minor version must not be >= 1000",
		},
		{
			versionStr:  "2.1.1234",
			wantsErr:    true,
			expectedErr: "patch version must not be >= 1000",
		},
		{
			versionStr:  "2.1.1.1234-enterprise",
			wantsErr:    true,
			expectedErr: "build version must not be >= 1000",
		},
	}

	for _, test := range tests {
		version, err := parseSemanticVersion(test.versionStr)
		if test.wantsErr {
			require.NotNil(t, err)
			require.EqualError(t, err, test.expectedErr)
			require.EqualValues(t, 0, version)
		} else {
			require.Nil(t, err)
			require.Equal(t, test.expectedVersion, version)
		}
	}
}

func TestVersionCompatibility_AddConfigTableUpdates(t *testing.T) {
	tests := []struct {
		name                       string
		configTablesUpdates        []map[uint64][]ConfigTableUpdates
		expectedConfigTableUpdates map[uint64][]ConfigTableUpdates
		expectedCount              int
	}{
		{
			name: "single addition of plugin payload updates",
			configTablesUpdates: []map[uint64][]ConfigTableUpdates{
				{
					2005999999: {
						{
							Name: "plugin_1",
							Type: Plugin,
							RemoveFields: []string{
								"plugin_1_field_1",
							},
						},
					},
				},
			},
			expectedConfigTableUpdates: map[uint64][]ConfigTableUpdates{
				2005999999: {
					{
						Name: "plugin_1",
						Type: Plugin,
						RemoveFields: []string{
							"plugin_1_field_1",
						},
					},
				},
			},
			expectedCount: 1,
		},
		{
			name: "multiple additions of plugin payload updates",
			configTablesUpdates: []map[uint64][]ConfigTableUpdates{
				{
					2005999999: {
						{
							Name: "plugin_1",
							Type: Plugin,
							RemoveFields: []string{
								"plugin_1_field_1",
							},
						},
					},
				},
				{
					2004999999: {
						{
							Name: "plugin_1",
							Type: Plugin,
							RemoveFields: []string{
								"plugin_1_field_1",
							},
						},
						{
							Name: "plugin_2",
							Type: Plugin,
							RemoveFields: []string{
								"plugin_2_field_1",
								"plugin_2_field_2",
							},
						},
					},
				},
			},
			expectedConfigTableUpdates: map[uint64][]ConfigTableUpdates{
				2005999999: {
					{
						Name: "plugin_1",
						Type: Plugin,
						RemoveFields: []string{
							"plugin_1_field_1",
						},
					},
				},
				2004999999: {
					{
						Name: "plugin_1",
						Type: Plugin,
						RemoveFields: []string{
							"plugin_1_field_1",
						},
					},
					{
						Name: "plugin_2",
						Type: Plugin,
						RemoveFields: []string{
							"plugin_2_field_1",
							"plugin_2_field_2",
						},
					},
				},
			},
			expectedCount: 2,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			wsvc, err := NewVersionCompatibilityProcessor(VersionCompatibilityOpts{
				Logger:        log.Logger,
				KongCPVersion: "2.8.0",
			})
			require.Nil(t, err)
			for _, configTableUpdate := range test.configTablesUpdates {
				err := wsvc.AddConfigTableUpdates(configTableUpdate)
				require.Nil(t, err)
			}
			require.Equal(t, test.expectedConfigTableUpdates, wsvc.configTableUpdates)
			require.Equal(t, test.expectedCount, len(wsvc.configTableUpdates))
		})
	}
}

// Used for TestVersionCompatibility_GetConfigTableUpdates.
var (
	pluginPayloadUpdates27x = map[uint64][]ConfigTableUpdates{
		2007999999: {
			{
				Name: "plugin_1",
				Type: Plugin,
				RemoveFields: []string{
					"plugin_1_field_1",
					"plugin_1_field_2",
				},
				RemoveElementsFromArray: []ConfigTableFieldCondition{
					{
						Field:     "plugin_1_array_field_1",
						Condition: "array_element_1=condition",
					},
				},
			},
			{
				Name: "plugin_2",
				Type: Plugin,
				RemoveFields: []string{
					"plugin_2_field_1",
				},
			},
			{
				Name: "plugin_3",
				Type: Plugin,
				RemoveFields: []string{
					"plugin_3_field_1",
					"plugin_3_field_2",
					"plugin_3_field_3",
					"plugin_3_field_4",
				},
				RemoveElementsFromArray: []ConfigTableFieldCondition{
					{
						Field:     "plugin_3_array_field_1",
						Condition: "array_element_1=condition",
					},
					{
						Field:     "plugin_3_array_field_2",
						Condition: "array_element_2=condition",
					},
				},
			},
		},
	}
	pluginPayloadUpdates26x = map[uint64][]ConfigTableUpdates{
		2006999999: {
			{
				Name: "plugin_1",
				Type: Plugin,
				RemoveFields: []string{
					"plugin_1_field_1",
				},
			},
		},
	}
	pluginPayloadUpdates25xAnd24x = map[uint64][]ConfigTableUpdates{
		2005999999: {
			{
				Name: "plugin_1",
				Type: Plugin,
				RemoveFields: []string{
					"plugin_1_field_1",
				},
			},
		},
		2004999999: {
			{
				Name: "plugin_1",
				Type: Plugin,
				RemoveFields: []string{
					"plugin_1_field_1",
				},
			},
			{
				Name: "plugin_2",
				Type: Plugin,
				RemoveFields: []string{
					"plugin_3_field_1",
					"plugin_3_field_2",
				},
				RemoveElementsFromArray: []ConfigTableFieldCondition{
					{
						Field:     "plugin_2_array_field_1",
						Condition: "array_element_1=condition",
					},
				},
			},
		},
	}
)

func allPluginPlayloadUpdates() map[uint64][]ConfigTableUpdates {
	pluginPayloadUpdates := make(map[uint64][]ConfigTableUpdates)
	for k, v := range pluginPayloadUpdates25xAnd24x {
		pluginPayloadUpdates[k] = v
	}
	for k, v := range pluginPayloadUpdates26x {
		pluginPayloadUpdates[k] = v
	}
	for k, v := range pluginPayloadUpdates27x {
		pluginPayloadUpdates[k] = v
	}
	return pluginPayloadUpdates
}

func TestVersionCompatibility_GetConfigTableUpdates(t *testing.T) {
	wsvc, err := NewVersionCompatibilityProcessor(VersionCompatibilityOpts{
		Logger:        log.Logger,
		KongCPVersion: "2.8.0",
	})
	require.Nil(t, err)
	err = wsvc.AddConfigTableUpdates(allPluginPlayloadUpdates())
	require.Nil(t, err)

	tests := []struct {
		name                       string
		dataPlaneVersion           uint64
		expectedConfigTableUpdates func() []ConfigTableUpdates
	}{
		{
			name:             "current version - no config table updates",
			dataPlaneVersion: 2008000000,
			expectedConfigTableUpdates: func() []ConfigTableUpdates {
				return []ConfigTableUpdates{}
			},
		},
		{
			name:             "previous version - < 2.8",
			dataPlaneVersion: 2007000000,
			expectedConfigTableUpdates: func() []ConfigTableUpdates {
				var pluginPayloadUpdates []ConfigTableUpdates
				pluginPayloadUpdates = append(pluginPayloadUpdates, pluginPayloadUpdates27x[2007999999]...)
				return pluginPayloadUpdates
			},
		},
		{
			name:             "previous version with a new minor version - < 2.8",
			dataPlaneVersion: 2007001000,
			expectedConfigTableUpdates: func() []ConfigTableUpdates {
				var pluginPayloadUpdates []ConfigTableUpdates
				pluginPayloadUpdates = append(pluginPayloadUpdates, pluginPayloadUpdates27x[2007999999]...)
				return pluginPayloadUpdates
			},
		},
		{
			name:             "older version - < 2.7",
			dataPlaneVersion: 2006000000,
			expectedConfigTableUpdates: func() []ConfigTableUpdates {
				var pluginPayloadUpdates []ConfigTableUpdates
				pluginPayloadUpdates = append(pluginPayloadUpdates, pluginPayloadUpdates27x[2007999999]...)
				pluginPayloadUpdates = append(pluginPayloadUpdates, pluginPayloadUpdates26x[2006999999]...)
				return pluginPayloadUpdates
			},
		},
		{
			name:             "older version - < 2.6",
			dataPlaneVersion: 2005000000,
			expectedConfigTableUpdates: func() []ConfigTableUpdates {
				var pluginPayloadUpdates []ConfigTableUpdates
				pluginPayloadUpdates = append(pluginPayloadUpdates, pluginPayloadUpdates27x[2007999999]...)
				pluginPayloadUpdates = append(pluginPayloadUpdates, pluginPayloadUpdates26x[2006999999]...)
				pluginPayloadUpdates = append(pluginPayloadUpdates, pluginPayloadUpdates25xAnd24x[2005999999]...)
				return pluginPayloadUpdates
			},
		},
		{
			name:             "older version - < 2.4",
			dataPlaneVersion: 2003002000,
			expectedConfigTableUpdates: func() []ConfigTableUpdates {
				var pluginPayloadUpdates []ConfigTableUpdates
				pluginPayloadUpdates = append(pluginPayloadUpdates, pluginPayloadUpdates27x[2007999999]...)
				pluginPayloadUpdates = append(pluginPayloadUpdates, pluginPayloadUpdates26x[2006999999]...)
				pluginPayloadUpdates = append(pluginPayloadUpdates, pluginPayloadUpdates25xAnd24x[2005999999]...)
				pluginPayloadUpdates = append(pluginPayloadUpdates, pluginPayloadUpdates25xAnd24x[2004999999]...)
				return pluginPayloadUpdates
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			pluginPayloadUpdates := wsvc.getConfigTableUpdates(test.dataPlaneVersion)
			require.ElementsMatch(t, test.expectedConfigTableUpdates(), pluginPayloadUpdates)
		})
	}
}

func TestVersionCompatibility_ProcessConfigTableUpdates(t *testing.T) {
	tests := []struct {
		name                string
		configTableUpdates  map[uint64][]ConfigTableUpdates
		uncompressedPayload string
		dataPlaneVersion    uint64
		expectedPayload     string
	}{
		{
			name: "single field element",
			configTableUpdates: map[uint64][]ConfigTableUpdates{
				2007999999: {
					{
						Name: "plugin_1",
						Type: Plugin,
						RemoveFields: []string{
							"plugin_1_field_1",
						},
					},
				},
			},
			uncompressedPayload: `{
				"config_table": {
					"plugins": [
						{
							"name": "plugin_1",
							"config": {
								"plugin_1_field_1": "element"
							}
						}
					]
				}
			}`,
			dataPlaneVersion: 2007000000,
			expectedPayload: `{
				"config_table": {
					"plugins": [
						{
							"name": "plugin_1",
							"config": {}
						}
					]
				}
			}`,
		},
		{
			name: "single field object",
			configTableUpdates: map[uint64][]ConfigTableUpdates{
				2007999999: {
					{
						Name: "plugin_1",
						Type: Plugin,
						RemoveFields: []string{
							"plugin_1_field_1",
						},
					},
				},
			},
			uncompressedPayload: `{
				"config_table": {
					"plugins": [
						{
							"name": "plugin_1",
							"config": {
								"plugin_1_field_1": {
									"object_1": "element",
									"object_2": "element"
								}
							}
						}
					]
				}
			}`,
			dataPlaneVersion: 2007000000,
			expectedPayload: `{
				"config_table": {
					"plugins": [
						{
							"name": "plugin_1",
							"config": {}
						}
					]
				}
			}`,
		},
		{
			name: "single field array",
			configTableUpdates: map[uint64][]ConfigTableUpdates{
				2007999999: {
					{
						Name: "plugin_1",
						Type: Plugin,
						RemoveFields: []string{
							"plugin_1_field_1",
						},
					},
				},
			},
			uncompressedPayload: `{
				"config_table": {
					"plugins": [
						{
							"name": "plugin_1",
							"config": {
								"plugin_1_field_1": [
									"item_1",
									"item_2"
								]
							}
						}
					]
				}
			}`,
			dataPlaneVersion: 2007000000,
			expectedPayload: `{
				"config_table": {
					"plugins": [
						{
							"name": "plugin_1",
							"config": {}
						}
					]
				}
			}`,
		},
		{
			name: "single field element where field is last",
			configTableUpdates: map[uint64][]ConfigTableUpdates{
				2007999999: {
					{
						Name: "plugin_1",
						Type: Plugin,
						RemoveFields: []string{
							"plugin_1_field_2",
						},
					},
				},
			},
			uncompressedPayload: `{
				"config_table": {
					"plugins": [
						{
							"name": "plugin_1",
							"config": {
								"plugin_1_field_1": "element",
								"plugin_1_field_2": "element"
							}
						}
					]
				}
			}`,
			dataPlaneVersion: 2007000000,
			expectedPayload: `{
				"config_table": {
					"plugins": [
						{
							"name": "plugin_1",
							"config": {
								"plugin_1_field_1": "element"
							}
						}
					]
				}
			}`,
		},
		{
			name: "multiple field elements",
			configTableUpdates: map[uint64][]ConfigTableUpdates{
				2007999999: {
					{
						Name: "plugin_1",
						Type: Plugin,
						RemoveFields: []string{
							"plugin_1_field_1",
						},
					},
					{
						Name: "plugin_3",
						Type: Plugin,
						RemoveFields: []string{
							"plugin_3_field_1",
						},
					},
				},
			},
			uncompressedPayload: `{
				"config_table": {
					"plugins": [
						{
							"name": "plugin_1",
							"config": {
								"plugin_1_field_1": "element"
							}
						},
						{
							"name": "plugin_3",
							"config": {
								"plugin_3_field_2": "element",
								"plugin_3_field_1": "element"
							}
						}
					]
				}
			}`,
			dataPlaneVersion: 2007000000,
			expectedPayload: `{
				"config_table": {
					"plugins": [
						{
							"name": "plugin_1",
							"config": {}
						},
						{
							"name": "plugin_3",
							"config": {
								"plugin_3_field_2": "element"
							}
						}
					]
				}
			}`,
		},
		{
			name: "multiple field elements from multiple plugin instances",
			configTableUpdates: map[uint64][]ConfigTableUpdates{
				2007999999: {
					{
						Name: "plugin_1",
						Type: Plugin,
						RemoveFields: []string{
							"plugin_1_field_1",
						},
					},
				},
			},
			uncompressedPayload: `{
				"config_table": {
					"plugins": [
						{
							"name": "plugin_1",
							"config": {
								"plugin_1_field_1": "element"
							}
						},
						{
							"name": "plugin_1",
							"config": {
								"plugin_1_field_2": "element",
								"plugin_1_field_1": "element"
							}
						}
					]
				}
			}`,
			dataPlaneVersion: 2007000000,
			expectedPayload: `{
				"config_table": {
					"plugins": [
						{
							"name": "plugin_1",
							"config": {}
						},
						{
							"name": "plugin_1",
							"config": {
								"plugin_1_field_2": "element"
							}
						}
					]
				}
			}`,
		},
		{
			name: "nested field element",
			configTableUpdates: map[uint64][]ConfigTableUpdates{
				2007999999: {
					{
						Name: "plugin_1",
						Type: Plugin,
						RemoveFields: []string{
							"plugin_1_field_1.plugin_1_nested_field_1",
						},
					},
				},
			},
			uncompressedPayload: `{
				"config_table": {
					"plugins": [
						{
							"name": "plugin_1",
							"config": {
								"plugin_1_field_1": {
									"plugin_1_nested_field_1": "element"
								}
							}
						}
					]
				}
			}`,
			dataPlaneVersion: 2007000000,
			expectedPayload: `{
				"config_table": {
					"plugins": [
						{
							"name": "plugin_1",
							"config": {
								"plugin_1_field_1": {}
							}
						}
					]
				}
			}`,
		},
		{
			name: "nested field with additional fields remaining",
			configTableUpdates: map[uint64][]ConfigTableUpdates{
				2007999999: {
					{
						Name: "plugin_1",
						Type: Plugin,
						RemoveFields: []string{
							"plugin_1_field_1.plugin_1_nested_field_1",
						},
					},
				},
			},
			uncompressedPayload: `{
				"config_table": {
					"plugins": [
						{
							"name": "plugin_1",
							"config": {
								"plugin_1_field_1": {
									"plugin_1_nested_field_1": "element",
									"plugin_1_nested_field_2": {
										"plugin_1_nested_field_2_nested": "element_nested_field_2_nested"
									}
								}
							}
						}
					]
				}
			}`,
			dataPlaneVersion: 2007000000,
			expectedPayload: `{
				"config_table": {
					"plugins": [
						{
							"name": "plugin_1",
							"config": {
								"plugin_1_field_1": {
									"plugin_1_nested_field_2": {
										"plugin_1_nested_field_2_nested": "element_nested_field_2_nested"
									}
								}
							}
						}
					]
				}
			}`,
		},
		{
			name: "two minor versions older",
			configTableUpdates: map[uint64][]ConfigTableUpdates{
				2007999999: {
					{
						Name: "plugin_1",
						Type: Plugin,
						RemoveFields: []string{
							"plugin_1_field_2",
						},
					},
				},
				2006999999: {
					{
						Name: "plugin_1",
						Type: Plugin,
						RemoveFields: []string{
							"plugin_1_field_1",
						},
					},
				},
			},
			uncompressedPayload: `{
				"config_table": {
					"plugins": [
						{
							"name": "plugin_1",
							"config": {
								"plugin_1_field_1": "element",
								"plugin_1_field_2": "element",
								"plugin_1_field_3": "element"
							}
						}
					]
				}
			}`,
			dataPlaneVersion: 2006000000,
			expectedPayload: `{
				"config_table": {
					"plugins": [
						{
							"name": "plugin_1",
							"config": {
								"plugin_1_field_3": "element"
							}
						}
					]
				}
			}`,
		},
		{
			name: "single field array removal with single item in array",
			configTableUpdates: map[uint64][]ConfigTableUpdates{
				2007999999: {
					{
						Name: "plugin_1",
						Type: Plugin,
						RemoveElementsFromArray: []ConfigTableFieldCondition{
							{
								Field:     "plugin_field_array_1",
								Condition: "array_element_1=condition",
							},
						},
					},
				},
			},
			uncompressedPayload: `{
				"config_table": {
					"plugins": [
						{
							"name": "plugin_1",
							"config": {
								"plugin_field_array_1": [
									{
										"array_element_1": "condition"
									}
								]
							}
						}
					]
				}
			}`,
			dataPlaneVersion: 2007000000,
			expectedPayload: `{
				"config_table": {
					"plugins": [
						{
							"name": "plugin_1",
							"config": {
								"plugin_field_array_1": []
							}
						}
					]
				}
			}`,
		},
		{
			name: "single nested field array removal with single item in array",
			configTableUpdates: map[uint64][]ConfigTableUpdates{
				2007999999: {
					{
						Name: "plugin_1",
						Type: Plugin,
						RemoveElementsFromArray: []ConfigTableFieldCondition{
							{
								Field:     "plugin_field_1.plugin_field_array_1",
								Condition: "array_element_1=condition",
							},
						},
					},
				},
			},
			uncompressedPayload: `{
				"config_table": {
					"plugins": [
						{
							"name": "plugin_1",
							"config": {
								"plugin_field_1": {
									"plugin_field_array_1": [
										{
											"array_element_1": "condition"
										}
									]
								}
							}
						}
					]
				}
			}`,
			dataPlaneVersion: 2007000000,
			expectedPayload: `{
				"config_table": {
					"plugins": [
						{
							"name": "plugin_1",
							"config": {
								"plugin_field_1": {
									"plugin_field_array_1": []
								}
							}
						}
					]
				}
			}`,
		},
		{
			name: "single field array removal with multiple items in array",
			configTableUpdates: map[uint64][]ConfigTableUpdates{
				2007999999: {
					{
						Name: "plugin_1",
						Type: Plugin,
						RemoveElementsFromArray: []ConfigTableFieldCondition{
							{
								Field:     "plugin_field_array_1",
								Condition: "array_element_1=condition",
							},
						},
					},
				},
			},
			uncompressedPayload: `{
				"config_table": {
					"plugins": [
						{
							"name": "plugin_1",
							"config": {
								"plugin_field_array_1": [
									{
										"array_element_1": "value_index_1",
										"array_element_2": "value_index_1",
										"array_element_3": "value_index_1"
									},
									{
										"array_element_1": "value_index_2",
										"array_element_2": "value_index_2",
										"array_element_3": "value_index_2"
									},
									{
										"array_element_1": "condition",
										"array_element_2": "value_index_3",
										"array_element_3": "value_index_3"
									},
									{
										"array_element_1": "value_index_4",
										"array_element_2": "value_index_4",
										"array_element_3": "value_index_4"
									}
								]
							}
						}
					]
				}
			}`,
			dataPlaneVersion: 2007000000,
			expectedPayload: `{
				"config_table": {
					"plugins": [
						{
							"name": "plugin_1",
							"config": {
								"plugin_field_array_1": [
									{
										"array_element_1": "value_index_1",
										"array_element_2": "value_index_1",
										"array_element_3": "value_index_1"
									},
									{
										"array_element_1": "value_index_2",
										"array_element_2": "value_index_2",
										"array_element_3": "value_index_2"
									},
									{
										"array_element_1": "value_index_4",
										"array_element_2": "value_index_4",
										"array_element_3": "value_index_4"
									}
								]
							}
						}
					]
				}
			}`,
		},
		{
			name: "field and array removal with multiple array removals",
			configTableUpdates: map[uint64][]ConfigTableUpdates{
				2007999999: {
					{
						Name: "plugin_1",
						Type: Plugin,
						RemoveFields: []string{
							"plugin_1_field_1",
						},
						RemoveElementsFromArray: []ConfigTableFieldCondition{
							{
								Field:     "plugin_field_array_2",
								Condition: "array_element_3=condition",
							},
						},
					},
				},
			},
			uncompressedPayload: `{
				"config_table": {
					"plugins": [
						{
							"name": "plugin_1",
							"config": {
								"plugin_1_field_1": "element"
								"plugin_field_array_2": [
									{
										"array_element_1": "value_index_1",
										"array_element_2": "value_index_1",
										"array_element_3": "value_index_1"
									},
									{
										"array_element_1": "value_index_2",
										"array_element_2": "value_index_2",
										"array_element_3": "condition"
									},
									{
										"array_element_1": "condition",
										"array_element_2": "value_index_3",
										"array_element_3": "value_index_3"
									},
									{
										"array_element_1": "value_index_4",
										"array_element_2": "value_index_4",
										"array_element_3": "condition"
									}
								],
								"plugin_1_field_3": "element"
							}
						}
					]
				}
			}`,
			dataPlaneVersion: 2007000000,
			expectedPayload: `{
				"config_table": {
					"plugins": [
						{
							"name": "plugin_1",
							"config": {
								"plugin_field_array_2": [
									{
										"array_element_1": "value_index_1",
										"array_element_2": "value_index_1",
										"array_element_3": "value_index_1"
									},
									{
										"array_element_1": "condition",
										"array_element_2": "value_index_3",
										"array_element_3": "value_index_3"
									}
								],
								"plugin_1_field_3": "element"
							}
						}
					]
				}
			}`,
		},
		{
			name: "no array removal",
			configTableUpdates: map[uint64][]ConfigTableUpdates{
				2007999999: {
					{
						Name: "plugin_1",
						Type: Plugin,
						RemoveFields: []string{
							"plugin_1_field_1",
						},
						RemoveElementsFromArray: []ConfigTableFieldCondition{
							{
								Field:     "plugin_field_array_2",
								Condition: "array_element_3=condition",
							},
						},
					},
				},
			},
			uncompressedPayload: `{
				"config_table": {
					"plugins": [
						{
							"name": "plugin_1",
							"config": {
								"plugin_1_field_1": "element",
								"plugin_field_array_2": [
									{
										"array_element_1": "condition",
										"array_element_2": "value_index_1",
										"array_element_3": "value_index_1"
									},
									{
										"array_element_1": "value_index_2",
										"array_element_2": "condition",
										"array_element_3": "value_index_2"
									}
								],
								"plugin_1_field_3": "element"
							}
						}
					]
				}
			}`,
			dataPlaneVersion: 2007000000,
			expectedPayload: `{
				"config_table": {
					"plugins": [
						{
							"name": "plugin_1",
							"config": {
								"plugin_field_array_2": [
									{
										"array_element_1": "condition",
										"array_element_2": "value_index_1",
										"array_element_3": "value_index_1"
									},
									{
										"array_element_1": "value_index_2",
										"array_element_2": "condition",
										"array_element_3": "value_index_2"
									}
								],
								"plugin_1_field_3": "element"
							}
						}
					]
				}
			}`,
		},
		{
			name: "no array removal with multiple versions defined",
			configTableUpdates: map[uint64][]ConfigTableUpdates{
				2007999999: {
					{
						Name: "plugin_1",
						Type: Plugin,
						RemoveFields: []string{
							"plugin_1_field_1",
						},
						RemoveElementsFromArray: []ConfigTableFieldCondition{
							{
								Field:     "plugin_field_array_2",
								Condition: "array_element_3=condition",
							},
						},
					},
				},
				2006999999: {
					{
						Name: "plugin_1",
						Type: Plugin,
						RemoveFields: []string{
							"plugin_1_field_4",
						},
						RemoveElementsFromArray: []ConfigTableFieldCondition{
							{
								Field:     "plugin_field_array_2",
								Condition: "array_element_2=condition",
							},
							{
								Field:     "plugin_field_array_5",
								Condition: "array_element_1=condition",
							},
						},
					},
				},
			},
			uncompressedPayload: `{
				"config_table": {
					"plugins": [
						{
							"name": "plugin_1",
							"config": {
								"plugin_1_field_1": "element",
								"plugin_field_array_2": [
									{
										"array_element_1": "condition",
										"array_element_2": "value_index_1",
										"array_element_3": "value_index_1"
									},
									{
										"array_element_1": "condition",
										"array_element_2": "value_index_2",
										"array_element_3": "value_index_2",
										"array_element_4": "condition"
									}
								],
								"plugin_1_field_3": "element",
								"plugin_1_field_4": "element",
								"plugin_field_array_5": [
									{
										"array_element_1": "value_index_1"
									}
								]
							}
						}
					]
				}
			}`,
			dataPlaneVersion: 2006000000,
			expectedPayload: `{
				"config_table": {
					"plugins": [
						{
							"name": "plugin_1",
							"config": {
								"plugin_field_array_2": [
									{
										"array_element_1": "condition",
										"array_element_2": "value_index_1",
										"array_element_3": "value_index_1"
									},
									{
										"array_element_1": "condition",
										"array_element_2": "value_index_2",
										"array_element_3": "value_index_2",
										"array_element_4": "condition"
									}
								],
								"plugin_1_field_3": "element",
								"plugin_field_array_5": [
									{
										"array_element_1": "value_index_1"
									}
								]
							}
						}
					]
				}
			}`,
		},
		{
			name: "single field array removal with single item in array",
			configTableUpdates: map[uint64][]ConfigTableUpdates{
				2007999999: {
					{
						Name: "plugin_1",
						Type: Plugin,
						RemoveElementsFromArray: []ConfigTableFieldCondition{
							{
								Field:     "plugin_field_array_1",
								Condition: "=item_3",
							},
						},
					},
				},
			},
			uncompressedPayload: `{
				"config_table": {
					"plugins": [
						{
							"name": "plugin_1",
							"config": {
								"plugin_field_array_1": [
									"item_1",
									"item_2",
									"item_3",
									"item_4",
									"item_5"
								]
							}
						}
					]
				}
			}`,
			dataPlaneVersion: 2007000000,
			expectedPayload: `{
				"config_table": {
					"plugins": [
						{
							"name": "plugin_1",
							"config": {
								"plugin_field_array_1": [
									"item_1",
									"item_2",
									"item_4",
									"item_5"
								]
							}
						}
					]
				}
			}`,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("plugin update for %s", test.name), func(t *testing.T) {
			wsvc, err := NewVersionCompatibilityProcessor(VersionCompatibilityOpts{
				Logger:        log.Logger,
				KongCPVersion: "2.8.0",
			})
			require.Nil(t, err)
			err = wsvc.AddConfigTableUpdates(test.configTableUpdates)
			require.Nil(t, err)

			processedPayload, err := wsvc.processConfigTableUpdates(test.uncompressedPayload,
				test.dataPlaneVersion)
			require.Nil(t, err)
			require.JSONEq(t, test.expectedPayload, processedPayload)
		})
	}

	t.Run("ensure processing does not occur", func(t *testing.T) {
		wsvc, err := NewVersionCompatibilityProcessor(VersionCompatibilityOpts{
			Logger:        log.Logger,
			KongCPVersion: "2.8.0",
		})
		require.Nil(t, err)

		payload := `{"config_table": {"extra_processing": "unprocessed"}, "type": "reconfigure"}`
		expectedPayload, err := CompressPayload([]byte(payload))
		require.Nil(t, err)
		processedPayloadCompressed, err := wsvc.ProcessConfigTableUpdates("2.8.0", expectedPayload)
		require.Nil(t, err)

		require.Equal(t, expectedPayload, processedPayloadCompressed)
	})
}

func TestVersionCompatibility_PerformExtraProcessing(t *testing.T) {
	tests := []struct {
		name             string
		wantsErr         bool
		wantsInvalidJSON bool
		isEnterprise     bool
		expectedPayload  string
	}{
		{
			name:            "data plane is OSS",
			expectedPayload: `{"extra_processing": "oss"}`,
		},
		{
			name:            "data plane is enterprise",
			isEnterprise:    true,
			expectedPayload: `{"extra_processing": "enterprise-edition"}`,
		},
		{
			name:     "extra processing returns an error",
			wantsErr: true,
		},
		{
			name:             "extra processing returns invalid JSON",
			wantsInvalidJSON: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			wsvc, err := NewVersionCompatibilityProcessor(VersionCompatibilityOpts{
				Logger:        log.Logger,
				KongCPVersion: "2.8.0",
				ExtraProcessing: func(uncompressedPayload string, dataPlaneVersion uint64, isEnterprise bool,
					logger *zap.Logger,
				) (string, error) {
					if test.wantsErr {
						return "", fmt.Errorf("extra processing error")
					}
					if test.wantsInvalidJSON {
						return "invalid JSON", nil
					}
					if isEnterprise {
						return `{"extra_processing": "enterprise-edition"}`, nil
					}
					return `{"extra_processing": "oss"}`, nil
				},
			})
			require.Nil(t, err)

			processedPayload, err := wsvc.performExtraProcessing("{}", 2008000000, test.isEnterprise)
			if test.wantsErr || test.wantsInvalidJSON {
				require.NotNil(t, err)
				if test.wantsErr {
					require.EqualError(t, err, "extra processing error")
				} else {
					require.EqualError(t, err, "processed payload is no longer valid JSON")
				}
			} else {
				require.Nil(t, err)
				require.JSONEq(t, test.expectedPayload, processedPayload)
			}
		})
	}

	t.Run("ensure extra processing occurs regardless of data plane version checking", func(t *testing.T) {
		wsvc, err := NewVersionCompatibilityProcessor(VersionCompatibilityOpts{
			Logger:        log.Logger,
			KongCPVersion: "2.8.0",
			ExtraProcessing: func(uncompressedPayload string, dataPlaneVersion uint64, isEnterprise bool,
				logger *zap.Logger,
			) (string, error) {
				return sjson.Set(uncompressedPayload, "config_table.extra_processing", "processed")
			},
		})
		require.Nil(t, err)

		payload := `{"config_table": {"extra_processing": "unprocessed"}, "type": "reconfigure"}`
		compressedPayload, err := CompressPayload([]byte(payload))
		require.Nil(t, err)
		processedPayloadCompressed, err := wsvc.ProcessConfigTableUpdates("2.8.0", compressedPayload)
		require.Nil(t, err)
		uncompressedPayload, err := UncompressPayload(processedPayloadCompressed)
		require.Nil(t, err)

		expectedPayload := `{
			"config_table": {
				"extra_processing": "processed"
			},
			"type": "reconfigure"
		}`
		require.JSONEq(t, expectedPayload, string(uncompressedPayload))
	})
}
