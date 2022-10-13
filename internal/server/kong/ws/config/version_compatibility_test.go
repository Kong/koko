package config

import (
	"fmt"
	"strings"
	"testing"

	"github.com/kong/koko/internal/log"
	"github.com/kong/koko/internal/versioning"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
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

func TestVersionCompatibility_AddConfigTableUpdates(t *testing.T) {
	tests := []struct {
		name                       string
		configTablesUpdates        []map[string][]ConfigTableUpdates
		expectedConfigTableUpdates map[string][]ConfigTableUpdates
		expectedCount              int
	}{
		{
			name: "single addition of plugin payload updates",
			configTablesUpdates: []map[string][]ConfigTableUpdates{
				{
					"< 2.6.0": {
						{
							Name: "plugin_1",
							Type: Plugin,
							RemoveFields: []string{
								"plugin_1_field_1",
							},
							ChangeID: "P042",
						},
					},
				},
			},
			expectedConfigTableUpdates: map[string][]ConfigTableUpdates{
				"< 2.6.0": {
					{
						Name: "plugin_1",
						Type: Plugin,
						RemoveFields: []string{
							"plugin_1_field_1",
						},
						ChangeID: "P042",
					},
				},
			},
			expectedCount: 1,
		},
		{
			name: "multiple additions of plugin payload updates",
			configTablesUpdates: []map[string][]ConfigTableUpdates{
				{
					"< 2.6.0": {
						{
							Name: "plugin_1",
							Type: Plugin,
							RemoveFields: []string{
								"plugin_1_field_1",
							},
							ChangeID: "P042",
						},
					},
				},
				{
					"< 2.5.0": {
						{
							Name: "plugin_1",
							Type: Plugin,
							RemoveFields: []string{
								"plugin_1_field_1",
							},
							ChangeID: "P043",
						},
						{
							Name: "plugin_2",
							Type: Plugin,
							RemoveFields: []string{
								"plugin_2_field_1",
								"plugin_2_field_2",
							},
							ChangeID: "P044",
						},
					},
				},
			},
			expectedConfigTableUpdates: map[string][]ConfigTableUpdates{
				"< 2.6.0": {
					{
						Name: "plugin_1",
						Type: Plugin,
						RemoveFields: []string{
							"plugin_1_field_1",
						},
						ChangeID: "P042",
					},
				},
				"< 2.5.0": {
					{
						Name: "plugin_1",
						Type: Plugin,
						RemoveFields: []string{
							"plugin_1_field_1",
						},
						ChangeID: "P043",
					},
					{
						Name: "plugin_2",
						Type: Plugin,
						RemoveFields: []string{
							"plugin_2_field_1",
							"plugin_2_field_2",
						},
						ChangeID: "P044",
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

	t.Run("ensure mutually exclusive attributes for FieldUpdate", func(t *testing.T) {
		tests := []map[string][]ConfigTableUpdates{
			{
				"< 2.6.0": {
					{
						Name: "plugin_1",
						Type: Plugin,
						FieldUpdates: []ConfigTableFieldCondition{
							{
								Field:     "plugin_1_field_1",
								Condition: "plugin_1_field_1=condition",
								Updates: []ConfigTableFieldUpdate{
									{
										Field:          "plugin_1_field_1",
										Value:          "value",
										ValueFromField: "plugin_1_field_2",
									},
								},
							},
						},
						ChangeID: "P032",
					},
				},
			},
			{
				"< 2.6.0": {
					{
						Name: "plugin_1",
						Type: Plugin,
						FieldUpdates: []ConfigTableFieldCondition{
							{
								Field:     "plugin_1_field_2",
								Condition: "plugin_1_field_2=condition",
								Updates: []ConfigTableFieldUpdate{
									{
										Field:          "plugin_1_field_1",
										ValueFromField: "plugin_1_field_2",
									},
								},
							},
						},
						ChangeID: "P033",
					},
				},
				"< 2.5.0": {
					{
						Name: "plugin_1",
						Type: Plugin,
						FieldUpdates: []ConfigTableFieldCondition{
							{
								Field:     "plugin_1_field_1",
								Condition: "plugin_1_field_1=condition",
								Updates: []ConfigTableFieldUpdate{
									{
										Field:          "plugin_1_field_2",
										Value:          "value",
										ValueFromField: "plugin_1_field_1",
									},
								},
							},
						},
						ChangeID: "P034",
					},
				},
			},
		}
		for _, test := range tests {
			wsvc, err := NewVersionCompatibilityProcessor(VersionCompatibilityOpts{
				Logger:        log.Logger,
				KongCPVersion: "2.8.0",
			})
			require.Nil(t, err)
			err = wsvc.AddConfigTableUpdates(test)
			require.NotNil(t, err)
			require.EqualError(t, err, "'Value' and 'ValueFromField' are mutually exclusive")
		}
	})
}

// Used for TestVersionCompatibility_GetConfigTableUpdates.
var (
	pluginPayloadUpdates27x = map[string][]ConfigTableUpdates{
		"< 2.8.0": {
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
				ChangeID: "T101",
			},
			{
				Name: "plugin_2",
				Type: Plugin,
				RemoveFields: []string{
					"plugin_2_field_1",
				},
				ChangeID: "T101",
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
				ChangeID: "T102",
			},
		},
	}
	pluginPayloadUpdates26x = map[string][]ConfigTableUpdates{
		"< 2.7.0": {
			{
				Name: "plugin_1",
				Type: Plugin,
				RemoveFields: []string{
					"plugin_1_field_1",
				},
				FieldUpdates: []ConfigTableFieldCondition{
					{
						Field:     "plugin_1_field_1",
						Condition: "plugin_1_field_1=condition",
						Updates: []ConfigTableFieldUpdate{
							{
								Field: "plugin_1_field_2",
								Value: "value",
							},
							{
								Field: "plugin_1_field_1",
							},
						},
					},
				},
				ChangeID: "T103",
			},
		},
	}
	pluginPayloadUpdates25xAnd24x = map[string][]ConfigTableUpdates{
		"< 2.6.0": {
			{
				Name: "plugin_1",
				Type: Plugin,
				RemoveFields: []string{
					"plugin_1_field_1",
				},
				ChangeID: "T104",
			},
		},
		"< 2.5.0": {
			{
				Name: "plugin_1",
				Type: Plugin,
				RemoveFields: []string{
					"plugin_1_field_1",
				},
				ChangeID: "T105",
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
				ChangeID: "T106",
			},
		},
	}
)

func allPluginPlayloadUpdates() map[string][]ConfigTableUpdates {
	pluginPayloadUpdates := make(map[string][]ConfigTableUpdates)
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
		dataPlaneVersion           string
		expectedConfigTableUpdates func() []ConfigTableUpdates
	}{
		{
			name:             "current version - no config table updates",
			dataPlaneVersion: "2.8.0",
			expectedConfigTableUpdates: func() []ConfigTableUpdates {
				return []ConfigTableUpdates{}
			},
		},
		{
			name:             "previous version - < 2.8",
			dataPlaneVersion: "2.7.0",
			expectedConfigTableUpdates: func() []ConfigTableUpdates {
				var pluginPayloadUpdates []ConfigTableUpdates
				pluginPayloadUpdates = append(pluginPayloadUpdates, pluginPayloadUpdates27x["< 2.8.0"]...)
				return pluginPayloadUpdates
			},
		},
		{
			name:             "previous version with a new minor version - < 2.8",
			dataPlaneVersion: "2.7.1",
			expectedConfigTableUpdates: func() []ConfigTableUpdates {
				var pluginPayloadUpdates []ConfigTableUpdates
				pluginPayloadUpdates = append(pluginPayloadUpdates, pluginPayloadUpdates27x["< 2.8.0"]...)
				return pluginPayloadUpdates
			},
		},
		{
			name:             "older version - < 2.7",
			dataPlaneVersion: "2.6.0",
			expectedConfigTableUpdates: func() []ConfigTableUpdates {
				var pluginPayloadUpdates []ConfigTableUpdates
				pluginPayloadUpdates = append(pluginPayloadUpdates, pluginPayloadUpdates27x["< 2.8.0"]...)
				pluginPayloadUpdates = append(pluginPayloadUpdates, pluginPayloadUpdates26x["< 2.7.0"]...)
				return pluginPayloadUpdates
			},
		},
		{
			name:             "older version - < 2.6",
			dataPlaneVersion: "2.5.0",
			expectedConfigTableUpdates: func() []ConfigTableUpdates {
				var pluginPayloadUpdates []ConfigTableUpdates
				pluginPayloadUpdates = append(pluginPayloadUpdates, pluginPayloadUpdates27x["< 2.8.0"]...)
				pluginPayloadUpdates = append(pluginPayloadUpdates, pluginPayloadUpdates26x["< 2.7.0"]...)
				pluginPayloadUpdates = append(pluginPayloadUpdates, pluginPayloadUpdates25xAnd24x["< 2.6.0"]...)
				return pluginPayloadUpdates
			},
		},
		{
			name:             "older version - < 2.4",
			dataPlaneVersion: "2.3.2",
			expectedConfigTableUpdates: func() []ConfigTableUpdates {
				var pluginPayloadUpdates []ConfigTableUpdates
				pluginPayloadUpdates = append(pluginPayloadUpdates, pluginPayloadUpdates27x["< 2.8.0"]...)
				pluginPayloadUpdates = append(pluginPayloadUpdates, pluginPayloadUpdates26x["< 2.7.0"]...)
				pluginPayloadUpdates = append(pluginPayloadUpdates, pluginPayloadUpdates25xAnd24x["< 2.6.0"]...)
				pluginPayloadUpdates = append(pluginPayloadUpdates, pluginPayloadUpdates25xAnd24x["< 2.5.0"]...)
				return pluginPayloadUpdates
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			version, err := versioning.NewVersion(test.dataPlaneVersion)
			require.NoError(t, err)
			pluginPayloadUpdates := wsvc.getConfigTableUpdates(version)
			require.ElementsMatch(t, test.expectedConfigTableUpdates(), pluginPayloadUpdates)
		})
	}
}

func TestVersionCompatibility_ProcessConfigTableUpdates(t *testing.T) {
	tests := []struct {
		name                string
		configTableUpdates  map[string][]ConfigTableUpdates
		uncompressedPayload string
		dataPlaneVersion    string
		expectedPayload     string
		expectedChanges     TrackedChanges
	}{
		{
			name: "single field element",
			configTableUpdates: map[string][]ConfigTableUpdates{
				"< 2.8.0": {
					{
						Name: "plugin_1",
						Type: Plugin,
						RemoveFields: []string{
							"plugin_1_field_1",
						},
						ChangeID: "T101",
					},
				},
			},
			uncompressedPayload: `{
				"config_table": {
					"plugins": [
						{
							"id": "933a565e-b645-4101-ab4a-a999fd1951eb",
							"name": "plugin_1",
							"config": {
								"plugin_1_field_1": "element"
							}
						}
					]
				}
			}`,
			dataPlaneVersion: "2.7.0",
			expectedPayload: `{
				"config_table": {
					"plugins": [
						{
							"id": "933a565e-b645-4101-ab4a-a999fd1951eb",
							"name": "plugin_1",
							"config": {}
						}
					]
				}
			}`,
			expectedChanges: TrackedChanges{
				ChangeDetails: []ChangeDetail{
					{
						ID: "T101",
						Resources: []ResourceInfo{
							{
								Type: "plugin",
								ID:   "933a565e-b645-4101-ab4a-a999fd1951eb",
							},
						},
					},
				},
			},
		},
		{
			name: "single field object",
			configTableUpdates: map[string][]ConfigTableUpdates{
				"< 2.8.0": {
					{
						Name: "plugin_1",
						Type: Plugin,
						RemoveFields: []string{
							"plugin_1_field_1",
						},
						ChangeID: "T101",
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
			dataPlaneVersion: "2.7.0",
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
			expectedChanges: TrackedChanges{},
		},
		{
			name: "single field array",
			configTableUpdates: map[string][]ConfigTableUpdates{
				"< 2.8.0": {
					{
						Name: "plugin_1",
						Type: Plugin,
						RemoveFields: []string{
							"plugin_1_field_1",
						},
						ChangeID: "T101",
					},
				},
			},
			uncompressedPayload: `{
				"config_table": {
					"plugins": [
						{
						"id": "47e46c41-e781-49d1-b4b8-d02e419b7ef5",
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
			dataPlaneVersion: "2.7.0",
			expectedPayload: `{
				"config_table": {
					"plugins": [
						{
							"id": "47e46c41-e781-49d1-b4b8-d02e419b7ef5",
							"name": "plugin_1",
							"config": {}
						}
					]
				}
			}`,
			expectedChanges: TrackedChanges{
				ChangeDetails: []ChangeDetail{
					{
						ID: "T101",
						Resources: []ResourceInfo{
							{
								Type: "plugin",
								ID:   "47e46c41-e781-49d1-b4b8-d02e419b7ef5",
							},
						},
					},
				},
			},
		},
		{
			name: "single field element where field is last",
			configTableUpdates: map[string][]ConfigTableUpdates{
				"< 2.8.0": {
					{
						Name: "plugin_1",
						Type: Plugin,
						RemoveFields: []string{
							"plugin_1_field_2",
						},
						ChangeID: "T101",
					},
				},
			},
			uncompressedPayload: `{
				"config_table": {
					"plugins": [
						{
							"id": "ab3b5a6d-923e-4e71-83b4-77e4b68d3e55",
							"name": "plugin_1",
							"config": {
								"plugin_1_field_1": "element",
								"plugin_1_field_2": "element"
							}
						}
					]
				}
			}`,
			dataPlaneVersion: "2.7.0",
			expectedPayload: `{
				"config_table": {
					"plugins": [
						{
							"id": "ab3b5a6d-923e-4e71-83b4-77e4b68d3e55",
							"name": "plugin_1",
							"config": {
								"plugin_1_field_1": "element"
							}
						}
					]
				}
			}`,
			expectedChanges: TrackedChanges{
				ChangeDetails: []ChangeDetail{
					{
						ID: "T101",
						Resources: []ResourceInfo{
							{
								Type: "plugin",
								ID:   "ab3b5a6d-923e-4e71-83b4-77e4b68d3e55",
							},
						},
					},
				},
			},
		},
		{
			name: "multiple field elements",
			configTableUpdates: map[string][]ConfigTableUpdates{
				"< 2.8.0": {
					{
						Name: "plugin_1",
						Type: Plugin,
						RemoveFields: []string{
							"plugin_1_field_1",
						},
						ChangeID: "T101",
					},
					{
						Name: "plugin_3",
						Type: Plugin,
						RemoveFields: []string{
							"plugin_3_field_1",
						},
						ChangeID: "T102",
					},
				},
			},
			uncompressedPayload: `{
				"config_table": {
					"plugins": [
						{
							"id": "ab3b5a6d-923e-4e71-83b4-77e4b68d3e55",
							"name": "plugin_1",
							"config": {
								"plugin_1_field_1": "element"
							}
						},
						{
							"id": "af1ab105-c9a3-42c8-9b12-442e3fd87a7f",
							"name": "plugin_3",
							"config": {
								"plugin_3_field_2": "element",
								"plugin_3_field_1": "element"
							}
						}
					]
				}
			}`,
			dataPlaneVersion: "2.7.0",
			expectedPayload: `{
				"config_table": {
					"plugins": [
						{
							"id": "ab3b5a6d-923e-4e71-83b4-77e4b68d3e55",
							"name": "plugin_1",
							"config": {}
						},
						{
							"id": "af1ab105-c9a3-42c8-9b12-442e3fd87a7f",
							"name": "plugin_3",
							"config": {
								"plugin_3_field_2": "element"
							}
						}
					]
				}
			}`,
			expectedChanges: TrackedChanges{
				ChangeDetails: []ChangeDetail{
					{
						ID: "T101",
						Resources: []ResourceInfo{
							{
								Type: "plugin",
								ID:   "ab3b5a6d-923e-4e71-83b4-77e4b68d3e55",
							},
						},
					},
					{
						ID: "T102",
						Resources: []ResourceInfo{
							{
								Type: "plugin",
								ID:   "af1ab105-c9a3-42c8-9b12-442e3fd87a7f",
							},
						},
					},
				},
			},
		},
		{
			name: "multiple field elements from multiple plugin instances",
			configTableUpdates: map[string][]ConfigTableUpdates{
				"< 2.8.0": {
					{
						Name: "plugin_1",
						Type: Plugin,
						RemoveFields: []string{
							"plugin_1_field_1",
						},
						ChangeID: "T101",
					},
					{
						Name: "plugin_2",
						Type: Plugin,
						RemoveFields: []string{
							"plugin_2_field_1",
						},
						ChangeID: "T102",
					},
				},
			},
			uncompressedPayload: `{
				"config_table": {
					"plugins": [
						{
							"id": "76a1d3aa-3eb5-4684-a626-6ce0f5afd380",
							"name": "plugin_2",
							"config": {
								"plugin_2_field_2": "element",
								"plugin_2_field_1": "element",
								"plugin_2_field_4": "element"
							}
						},
						{
							"id": "ab3b5a6d-923e-4e71-83b4-77e4b68d3e55",
							"name": "plugin_1",
							"config": {
								"plugin_1_field_1": "element"
							}
						},
						{
							"id": "5441b100-f441-4d4b-bcc2-3bb153e2bd40",
							"name": "plugin_1",
							"config": {
								"plugin_1_field_2": "element",
								"plugin_1_field_1": "element"
							}
						},
						{
							"id": "81295196-8963-4abd-9daa-4d39bf763e40",
							"name": "plugin_1",
							"config": {
								"plugin_1_field_2": "element",
								"plugin_1_field_1": "element",
								"plugin_1_field_4": "element"
							}
						},
						{
							"id": "76a1d3aa-3eb5-4684-a626-6ce0f5afd381",
							"name": "plugin_2",
							"config": {
								"plugin_2_field_2": "element"
							}
						},
						{
							"id": "81295196-8963-4abd-9daa-4d39bf763e41",
							"name": "plugin_1",
							"config": {
								"plugin_1_field_2": "element",
								"plugin_1_field_1": "element",
								"plugin_1_field_3": "element"
							}
						},
						{
							"id": "76a1d3aa-3eb5-4684-a626-6ce0f5afd382",
							"name": "plugin_2",
							"config": {
								"plugin_2_field_2": "element",
								"plugin_2_field_1": "element"
							}
						},
						{
							"id": "76a1d3aa-3eb5-4684-a626-6ce0f5afd383",
							"name": "plugin_2",
							"config": {
								"plugin_2_field_2": "element",
								"plugin_2_field_1": "element",
								"plugin_2_field_3": "element"
							}
						}
					]
				}
			}`,
			dataPlaneVersion: "2.7.0",
			expectedPayload: `{
				"config_table": {
					"plugins": [
						{
							"id": "76a1d3aa-3eb5-4684-a626-6ce0f5afd380",
							"name": "plugin_2",
							"config": {
								"plugin_2_field_2": "element",
								"plugin_2_field_4": "element"
							}
						},
						{
							"id": "ab3b5a6d-923e-4e71-83b4-77e4b68d3e55",
							"name": "plugin_1",
							"config": {}
						},
						{
							"id": "5441b100-f441-4d4b-bcc2-3bb153e2bd40",
							"name": "plugin_1",
							"config": {
								"plugin_1_field_2": "element"
							}
						},
						{
							"id": "81295196-8963-4abd-9daa-4d39bf763e40",
							"name": "plugin_1",
							"config": {
								"plugin_1_field_2": "element",
								"plugin_1_field_4": "element"
							}
						},
						{
							"id": "76a1d3aa-3eb5-4684-a626-6ce0f5afd381",
							"name": "plugin_2",
							"config": {
								"plugin_2_field_2": "element"
							}
						},
						{
							"id": "81295196-8963-4abd-9daa-4d39bf763e41",
							"name": "plugin_1",
							"config": {
								"plugin_1_field_2": "element",
								"plugin_1_field_3": "element"
							}
						},
						{
							"id": "76a1d3aa-3eb5-4684-a626-6ce0f5afd382",
							"name": "plugin_2",
							"config": {
								"plugin_2_field_2": "element"
							}
						},
						{
							"id": "76a1d3aa-3eb5-4684-a626-6ce0f5afd383",
							"name": "plugin_2",
							"config": {
								"plugin_2_field_2": "element",
								"plugin_2_field_3": "element"
							}
						}
					]
				}
			}`,
			expectedChanges: TrackedChanges{
				ChangeDetails: []ChangeDetail{
					{
						ID: "T101",
						Resources: []ResourceInfo{
							{
								Type: "plugin",
								ID:   "5441b100-f441-4d4b-bcc2-3bb153e2bd40",
							},
							{
								Type: "plugin",
								ID:   "81295196-8963-4abd-9daa-4d39bf763e40",
							},
							{
								Type: "plugin",
								ID:   "81295196-8963-4abd-9daa-4d39bf763e41",
							},
							{
								Type: "plugin",
								ID:   "ab3b5a6d-923e-4e71-83b4-77e4b68d3e55",
							},
						},
					},
					{
						ID: "T102",
						Resources: []ResourceInfo{
							{
								Type: "plugin",
								ID:   "76a1d3aa-3eb5-4684-a626-6ce0f5afd380",
							},
							{
								Type: "plugin",
								ID:   "76a1d3aa-3eb5-4684-a626-6ce0f5afd382",
							},
							{
								Type: "plugin",
								ID:   "76a1d3aa-3eb5-4684-a626-6ce0f5afd383",
							},
						},
					},
				},
			},
		},
		{
			name: "nested field element",
			configTableUpdates: map[string][]ConfigTableUpdates{
				"< 2.8.0": {
					{
						Name: "plugin_1",
						Type: Plugin,
						RemoveFields: []string{
							"plugin_1_field_1.plugin_1_nested_field_1",
						},
						ChangeID: "T101",
					},
				},
			},
			uncompressedPayload: `{
				"config_table": {
					"plugins": [
						{
							"id": "ab3b5a6d-923e-4e71-83b4-77e4b68d3e55",
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
			dataPlaneVersion: "2.7.0",
			expectedPayload: `{
				"config_table": {
					"plugins": [
						{
							"id": "ab3b5a6d-923e-4e71-83b4-77e4b68d3e55",
							"name": "plugin_1",
							"config": {
								"plugin_1_field_1": {}
							}
						}
					]
				}
			}`,
			expectedChanges: TrackedChanges{
				ChangeDetails: []ChangeDetail{
					{
						ID: "T101",
						Resources: []ResourceInfo{
							{
								Type: "plugin",
								ID:   "ab3b5a6d-923e-4e71-83b4-77e4b68d3e55",
							},
						},
					},
				},
			},
		},
		{
			name: "nested field with additional fields remaining",
			configTableUpdates: map[string][]ConfigTableUpdates{
				"< 2.8.0": {
					{
						Name: "plugin_1",
						Type: Plugin,
						RemoveFields: []string{
							"plugin_1_field_1.plugin_1_nested_field_1",
						},
						ChangeID: "T101",
					},
				},
			},
			uncompressedPayload: `{
				"config_table": {
					"plugins": [
						{
							"id": "ab3b5a6d-923e-4e71-83b4-77e4b68d3e55",
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
			dataPlaneVersion: "2.7.0",
			expectedPayload: `{
				"config_table": {
					"plugins": [
						{
							"id": "ab3b5a6d-923e-4e71-83b4-77e4b68d3e55",
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
			expectedChanges: TrackedChanges{
				ChangeDetails: []ChangeDetail{
					{
						ID: "T101",
						Resources: []ResourceInfo{
							{
								Type: "plugin",
								ID:   "ab3b5a6d-923e-4e71-83b4-77e4b68d3e55",
							},
						},
					},
				},
			},
		},
		{
			name: "two minor versions older",
			configTableUpdates: map[string][]ConfigTableUpdates{
				"< 2.8.0": {
					{
						Name: "plugin_1",
						Type: Plugin,
						RemoveFields: []string{
							"plugin_1_field_2",
						},
						ChangeID: "T101",
					},
				},
				"< 2.7.0": {
					{
						Name: "plugin_1",
						Type: Plugin,
						RemoveFields: []string{
							"plugin_1_field_1",
						},
						ChangeID: "T102",
					},
				},
			},
			uncompressedPayload: `{
				"config_table": {
					"plugins": [
						{
							"id": "ab3b5a6d-923e-4e71-83b4-77e4b68d3e55",
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
			dataPlaneVersion: "2.6.0",
			expectedPayload: `{
				"config_table": {
					"plugins": [
						{
							"id": "ab3b5a6d-923e-4e71-83b4-77e4b68d3e55",
							"name": "plugin_1",
							"config": {
								"plugin_1_field_3": "element"
							}
						}
					]
				}
			}`,
			expectedChanges: TrackedChanges{
				ChangeDetails: []ChangeDetail{
					{
						ID: "T101",
						Resources: []ResourceInfo{
							{
								Type: "plugin",
								ID:   "ab3b5a6d-923e-4e71-83b4-77e4b68d3e55",
							},
						},
					},
					{
						ID: "T102",
						Resources: []ResourceInfo{
							{
								Type: "plugin",
								ID:   "ab3b5a6d-923e-4e71-83b4-77e4b68d3e55",
							},
						},
					},
				},
			},
		},
		{
			name: "single field array removal with single item in array",
			configTableUpdates: map[string][]ConfigTableUpdates{
				"< 2.8.0": {
					{
						Name: "plugin_1",
						Type: Plugin,
						RemoveElementsFromArray: []ConfigTableFieldCondition{
							{
								Field:     "plugin_field_array_1",
								Condition: "array_element_1=condition",
							},
						},
						ChangeID: "T101",
					},
				},
			},
			uncompressedPayload: `{
				"config_table": {
					"plugins": [
						{
							"id": "ab3b5a6d-923e-4e71-83b4-77e4b68d3e55",
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
			dataPlaneVersion: "2.7.0",
			expectedPayload: `{
				"config_table": {
					"plugins": [
						{
							"id": "ab3b5a6d-923e-4e71-83b4-77e4b68d3e55",
							"name": "plugin_1",
							"config": {
								"plugin_field_array_1": []
							}
						}
					]
				}
			}`,
			expectedChanges: TrackedChanges{
				ChangeDetails: []ChangeDetail{
					{
						ID: "T101",
						Resources: []ResourceInfo{
							{
								Type: "plugin",
								ID:   "ab3b5a6d-923e-4e71-83b4-77e4b68d3e55",
							},
						},
					},
				},
			},
		},
		{
			name: "single nested field array removal with single item in array",
			configTableUpdates: map[string][]ConfigTableUpdates{
				"< 2.8.0": {
					{
						Name: "plugin_1",
						Type: Plugin,
						RemoveElementsFromArray: []ConfigTableFieldCondition{
							{
								Field:     "plugin_field_1.plugin_field_array_1",
								Condition: "array_element_1=condition",
							},
						},
						ChangeID: "T101",
					},
				},
			},
			uncompressedPayload: `{
				"config_table": {
					"plugins": [
						{
							"id": "29b2b210-3a3a-4344-8208-36c698aa9f5a",
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
			dataPlaneVersion: "2.7.0",
			expectedPayload: `{
				"config_table": {
					"plugins": [
						{
							"id": "29b2b210-3a3a-4344-8208-36c698aa9f5a",
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
			expectedChanges: TrackedChanges{
				ChangeDetails: []ChangeDetail{
					{
						ID: "T101",
						Resources: []ResourceInfo{
							{
								Type: "plugin",
								ID:   "29b2b210-3a3a-4344-8208-36c698aa9f5a",
							},
						},
					},
				},
			},
		},
		{
			name: "single field array removal with multiple items in array",
			configTableUpdates: map[string][]ConfigTableUpdates{
				"< 2.8.0": {
					{
						Name: "plugin_1",
						Type: Plugin,
						RemoveElementsFromArray: []ConfigTableFieldCondition{
							{
								Field:     "plugin_field_array_1",
								Condition: "array_element_1=condition",
							},
						},
						ChangeID: "T101",
					},
				},
			},
			uncompressedPayload: `{
				"config_table": {
					"plugins": [
						{
							"id": "29b2b210-3a3a-4344-8208-36c698aa9f5a",
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
			dataPlaneVersion: "2.7.0",
			expectedPayload: `{
				"config_table": {
					"plugins": [
						{
							"id": "29b2b210-3a3a-4344-8208-36c698aa9f5a",
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
			expectedChanges: TrackedChanges{
				ChangeDetails: []ChangeDetail{
					{
						ID: "T101",
						Resources: []ResourceInfo{
							{
								Type: "plugin",
								ID:   "29b2b210-3a3a-4344-8208-36c698aa9f5a",
							},
						},
					},
				},
			},
		},
		{
			name: "field and array removal with multiple array removals",
			configTableUpdates: map[string][]ConfigTableUpdates{
				"< 2.8.0": {
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
						ChangeID: "T101",
					},
				},
			},
			uncompressedPayload: `{
				"config_table": {
					"plugins": [
						{
							"id": "29b2b210-3a3a-4344-8208-36c698aa9f5a",
							"name": "plugin_1",
							"config": {
								"plugin_1_field_1": "element",
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
			dataPlaneVersion: "2.7.0",
			expectedPayload: `{
				"config_table": {
					"plugins": [
						{
							"id": "29b2b210-3a3a-4344-8208-36c698aa9f5a",
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
			expectedChanges: TrackedChanges{
				ChangeDetails: []ChangeDetail{
					{
						ID: "T101",
						Resources: []ResourceInfo{
							{
								Type: "plugin",
								ID:   "29b2b210-3a3a-4344-8208-36c698aa9f5a",
							},
						},
					},
				},
			},
		},
		{
			name: "no array removal",
			configTableUpdates: map[string][]ConfigTableUpdates{
				"< 2.8.0": {
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
						ChangeID: "T101",
					},
				},
			},
			uncompressedPayload: `{
				"config_table": {
					"plugins": [
						{
							"id": "29b2b210-3a3a-4344-8208-36c698aa9f5a",
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
			dataPlaneVersion: "2.7.0",
			expectedPayload: `{
				"config_table": {
					"plugins": [
						{
							"id": "29b2b210-3a3a-4344-8208-36c698aa9f5a",
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
			expectedChanges: TrackedChanges{
				ChangeDetails: []ChangeDetail{
					{
						ID: "T101",
						Resources: []ResourceInfo{
							{
								Type: "plugin",
								ID:   "29b2b210-3a3a-4344-8208-36c698aa9f5a",
							},
						},
					},
				},
			},
		},
		{
			name: "no array removal with multiple versions defined",
			configTableUpdates: map[string][]ConfigTableUpdates{
				"< 2.8.0": {
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
						ChangeID: "T101",
					},
				},
				"< 2.7.0": {
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
						ChangeID: "T102",
					},
				},
			},
			uncompressedPayload: `{
				"config_table": {
					"plugins": [
						{
							"id": "29b2b210-3a3a-4344-8208-36c698aa9f5a",
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
			dataPlaneVersion: "2.6.0",
			expectedPayload: `{
				"config_table": {
					"plugins": [
						{
							"id": "29b2b210-3a3a-4344-8208-36c698aa9f5a",
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
			expectedChanges: TrackedChanges{
				ChangeDetails: []ChangeDetail{
					{
						ID: "T101",
						Resources: []ResourceInfo{
							{
								Type: "plugin",
								ID:   "29b2b210-3a3a-4344-8208-36c698aa9f5a",
							},
						},
					},
					{
						ID: "T102",
						Resources: []ResourceInfo{
							{
								Type: "plugin",
								ID:   "29b2b210-3a3a-4344-8208-36c698aa9f5a",
							},
						},
					},
				},
			},
		},
		{
			name: "single field update with single item",
			configTableUpdates: map[string][]ConfigTableUpdates{
				"< 2.8.0": {
					{
						Name: "plugin_1",
						Type: Plugin,
						FieldUpdates: []ConfigTableFieldCondition{
							{
								Field:     "plugin_field_1",
								Condition: "plugin_field_1=condition",
								Updates: []ConfigTableFieldUpdate{
									{
										Field: "plugin_field_1",
										Value: "value_updated",
									},
								},
							},
						},
						ChangeID: "T101",
					},
				},
			},
			uncompressedPayload: `{
				"config_table": {
					"plugins": [
						{
							"id": "c50e912c-873c-45da-9f7c-f12a19fd56d1",
							"name": "plugin_1",
							"config": {
								"plugin_field_1": "condition"
							}
						}
					]
				}
			}`,
			dataPlaneVersion: "2.7.0",
			expectedPayload: `{
				"config_table": {
					"plugins": [
						{
							"id": "c50e912c-873c-45da-9f7c-f12a19fd56d1",
							"name": "plugin_1",
							"config": {
								"plugin_field_1": "value_updated"
							}
						}
					]
				}
			}`,
			expectedChanges: TrackedChanges{
				ChangeDetails: []ChangeDetail{
					{
						ID: "T101",
						Resources: []ResourceInfo{
							{
								Type: "plugin",
								ID:   "c50e912c-873c-45da-9f7c-f12a19fd56d1",
							},
						},
					},
				},
			},
		},
		{
			name: "field updates with multiple data types",
			configTableUpdates: map[string][]ConfigTableUpdates{
				"< 2.8.0": {
					{
						Name: "plugin_1",
						Type: Plugin,
						FieldUpdates: []ConfigTableFieldCondition{
							{
								Field:     "plugin_field_string",
								Condition: "plugin_field_string=old",
								Updates: []ConfigTableFieldUpdate{
									{
										Field: "plugin_field_string",
										Value: "new",
									},
								},
							},
							{
								Field:     "plugin_field_number",
								Condition: "plugin_field_number=9",
								Updates: []ConfigTableFieldUpdate{
									{
										Field: "plugin_field_number",
										Value: 28,
									},
								},
							},
							{
								Field:     "plugin_field_bool",
								Condition: "plugin_field_bool=true",
								Updates: []ConfigTableFieldUpdate{
									{
										Field: "plugin_field_bool",
										Value: false,
									},
								},
							},
						},
						ChangeID: "T101",
					},
				},
			},
			uncompressedPayload: `{
				"config_table": {
					"plugins": [
						{
							"id": "47e46c41-e781-49d1-b4b8-d02e419b7ef5",
							"name": "plugin_1",
							"config": {
								"plugin_field_string": "old",
								"plugin_field_number": 9,
								"plugin_field_bool": true
							}
						}
					]
				}
			}`,
			dataPlaneVersion: "2.7.0",
			expectedPayload: `{
				"config_table": {
					"plugins": [
						{
							"id": "47e46c41-e781-49d1-b4b8-d02e419b7ef5",
							"name": "plugin_1",
							"config": {
								"plugin_field_string": "new",
								"plugin_field_number": 28,
								"plugin_field_bool": false
							}
						}
					]
				}
			}`,
			expectedChanges: TrackedChanges{
				ChangeDetails: []ChangeDetail{
					{
						ID: "T101",
						Resources: []ResourceInfo{
							{
								Type: "plugin",
								ID:   "47e46c41-e781-49d1-b4b8-d02e419b7ef5",
							},
						},
					},
				},
			},
		},
		{
			name: "field update with multiple value updates",
			configTableUpdates: map[string][]ConfigTableUpdates{
				"< 2.8.0": {
					{
						Name: "plugin_1",
						Type: Plugin,
						FieldUpdates: []ConfigTableFieldCondition{
							{
								Field:     "plugin_field_1",
								Condition: "plugin_field_1=condition",
								Updates: []ConfigTableFieldUpdate{
									{
										Field: "plugin_field_1",
										Value: "value_updated",
									},
									{
										Field: "plugin_field_3",
										Value: "value_updated",
									},
								},
							},
						},
						ChangeID: "T101",
					},
				},
			},
			uncompressedPayload: `{
				"config_table": {
					"plugins": [
						{
							"id": "47e46c41-e781-49d1-b4b8-d02e419b7ef5",
							"name": "plugin_1",
							"config": {
								"plugin_field_1": "condition",
								"plugin_field_2": "value",
								"plugin_field_3": "value",
								"plugin_field_4": "value"
							}
						}
					]
				}
			}`,
			dataPlaneVersion: "2.7.0",
			expectedPayload: `{
				"config_table": {
					"plugins": [
						{
							"id": "47e46c41-e781-49d1-b4b8-d02e419b7ef5",
							"name": "plugin_1",
							"config": {
								"plugin_field_1": "value_updated",
								"plugin_field_2": "value",
								"plugin_field_3": "value_updated",
								"plugin_field_4": "value"
							}
						}
					]
				}
			}`,
			expectedChanges: TrackedChanges{
				ChangeDetails: []ChangeDetail{
					{
						ID: "T101",
						Resources: []ResourceInfo{
							{
								Type: "plugin",
								ID:   "47e46c41-e781-49d1-b4b8-d02e419b7ef5",
							},
						},
					},
				},
			},
		},
		{
			name: "nested field update",
			configTableUpdates: map[string][]ConfigTableUpdates{
				"< 2.8.0": {
					{
						Name: "plugin_1",
						Type: Plugin,
						FieldUpdates: []ConfigTableFieldCondition{
							{
								Field:     "plugin_field_1.nested_field_1",
								Condition: "plugin_field_1.nested_field_1=condition",
								Updates: []ConfigTableFieldUpdate{
									{
										Field: "plugin_field_1.nested_field_1",
										Value: "value_updated",
									},
								},
							},
						},
						ChangeID: "T101",
					},
				},
			},
			uncompressedPayload: `{
				"config_table": {
					"plugins": [
						{
							"id": "47e46c41-e781-49d1-b4b8-d02e419b7ef5",
							"name": "plugin_1",
							"config": {
								"plugin_field_1": {
									"nested_field_1": "condition"
								}
							}
						}
					]
				}
			}`,
			dataPlaneVersion: "2.7.0",
			expectedPayload: `{
				"config_table": {
					"plugins": [
						{
							"id": "47e46c41-e781-49d1-b4b8-d02e419b7ef5",
							"name": "plugin_1",
							"config": {
								"plugin_field_1": {
									"nested_field_1": "value_updated"
								}
							}
						}
					]
				}
			}`,
			expectedChanges: TrackedChanges{
				ChangeDetails: []ChangeDetail{
					{
						ID: "T101",
						Resources: []ResourceInfo{
							{
								Type: "plugin",
								ID:   "47e46c41-e781-49d1-b4b8-d02e419b7ef5",
							},
						},
					},
				},
			},
		},
		{
			name: "field update with additional nested field update",
			configTableUpdates: map[string][]ConfigTableUpdates{
				"< 2.8.0": {
					{
						Name: "plugin_1",
						Type: Plugin,
						FieldUpdates: []ConfigTableFieldCondition{
							{
								Field:     "plugin_field_1",
								Condition: "plugin_field_1=condition",
								Updates: []ConfigTableFieldUpdate{
									{
										Field: "plugin_field_1",
										Value: "value_updated",
									},
									{
										Field: "plugin_field_2.nested_field_1",
										Value: "value_updated",
									},
								},
							},
						},
						ChangeID: "T101",
					},
				},
			},
			uncompressedPayload: `{
				"config_table": {
					"plugins": [
						{
							"id": "47e46c41-e781-49d1-b4b8-d02e419b7ef5",
							"name": "plugin_1",
							"config": {
								"plugin_field_1": "condition",
								"plugin_field_2": {
									"nested_field_1": "value"
								}
							}
						}
					]
				}
			}`,
			dataPlaneVersion: "2.7.0",
			expectedPayload: `{
				"config_table": {
					"plugins": [
						{
							"id": "47e46c41-e781-49d1-b4b8-d02e419b7ef5",
							"name": "plugin_1",
							"config": {
								"plugin_field_1": "value_updated",
								"plugin_field_2": {
									"nested_field_1": "value_updated"
								}
							}
						}
					]
				}
			}`,
			expectedChanges: TrackedChanges{
				ChangeDetails: []ChangeDetail{
					{
						ID: "T101",
						Resources: []ResourceInfo{
							{
								Type: "plugin",
								ID:   "47e46c41-e781-49d1-b4b8-d02e419b7ef5",
							},
						},
					},
				},
			},
		},
		{
			name: "no field updates",
			configTableUpdates: map[string][]ConfigTableUpdates{
				"< 2.8.0": {
					{
						Name: "plugin_1",
						Type: Plugin,
						FieldUpdates: []ConfigTableFieldCondition{
							{
								Field:     "plugin_field_1",
								Condition: "plugin_field_1=condition",
								Updates: []ConfigTableFieldUpdate{
									{
										Field: "plugin_field_1",
										Value: "value_updated",
									},
									{
										Field: "plugin_field_2",
										Value: "value_updated",
									},
								},
							},
						},
						ChangeID: "T101",
					},
				},
			},
			uncompressedPayload: `{
				"config_table": {
					"plugins": [
						{
							"id": "47e46c41-e781-49d1-b4b8-d02e419b7ef5",
							"name": "plugin_1",
							"config": {
								"plugin_field_1": "value",
								"plugin_field_2": "condition",
								"plugin_field_3": "condition",
								"plugin_field_4": "value"
							}
						}
					]
				}
			}`,
			dataPlaneVersion: "2.7.0",
			expectedPayload: `{
				"config_table": {
					"plugins": [
						{
							"id": "47e46c41-e781-49d1-b4b8-d02e419b7ef5",
							"name": "plugin_1",
							"config": {
								"plugin_field_1": "value",
								"plugin_field_2": "condition",
								"plugin_field_3": "condition",
								"plugin_field_4": "value"
							}
						}
					]
				}
			}`,
			expectedChanges: TrackedChanges{},
		},
		{
			name: "field, array removal, and field update",
			configTableUpdates: map[string][]ConfigTableUpdates{
				"< 2.8.0": {
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
						FieldUpdates: []ConfigTableFieldCondition{
							{
								Field:     "plugin_1_field_3",
								Condition: "plugin_1_field_3=condition",
								Updates: []ConfigTableFieldUpdate{
									{
										Field: "plugin_1_field_3",
										Value: "value_updated",
									},
								},
							},
						},
						ChangeID: "T101",
					},
				},
			},
			uncompressedPayload: `{
				"config_table": {
					"plugins": [
						{
							"id": "47e46c41-e781-49d1-b4b8-d02e419b7ef5",
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
								"plugin_1_field_3": "condition",
								"plugin_1_field_4": "value"
							}
						}
					]
				}
			}`,
			dataPlaneVersion: "2.7.0",
			expectedPayload: `{
				"config_table": {
					"plugins": [
						{
							"id": "47e46c41-e781-49d1-b4b8-d02e419b7ef5",
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
								"plugin_1_field_3": "value_updated",
								"plugin_1_field_4": "value"
							}
						}
					]
				}
			}`,
			expectedChanges: TrackedChanges{
				ChangeDetails: []ChangeDetail{
					{
						ID: "T101",
						Resources: []ResourceInfo{
							{
								Type: "plugin",
								ID:   "47e46c41-e781-49d1-b4b8-d02e419b7ef5",
							},
						},
					},
				},
			},
		},
		{
			name: "field value create based on other field and delete",
			configTableUpdates: map[string][]ConfigTableUpdates{
				"< 2.8.0": {
					{
						Name: "plugin_1",
						Type: Plugin,
						FieldUpdates: []ConfigTableFieldCondition{
							{
								Field:     "plugin_field_1",
								Condition: "plugin_field_1",
								Updates: []ConfigTableFieldUpdate{
									{
										Field:          "plugin_field_2",
										ValueFromField: "plugin_field_1",
									},
									{
										Field: "plugin_field_1",
									},
								},
							},
						},
						ChangeID: "T101",
					},
				},
			},
			uncompressedPayload: `{
				"config_table": {
					"plugins": [
						{
							"id": "47e46c41-e781-49d1-b4b8-d02e419b7ef5",
							"name": "plugin_1",
							"config": {
								"plugin_field_1": "value"
							}
						}
					]
				}
			}`,
			dataPlaneVersion: "2.7.0",
			expectedPayload: `{
				"config_table": {
					"plugins": [
						{
							"id": "47e46c41-e781-49d1-b4b8-d02e419b7ef5",
							"name": "plugin_1",
							"config": {
								"plugin_field_2": "value"
							}
						}
					]
				}
			}`,
			expectedChanges: TrackedChanges{
				ChangeDetails: []ChangeDetail{
					{
						ID: "T101",
						Resources: []ResourceInfo{
							{
								Type: "plugin",
								ID:   "47e46c41-e781-49d1-b4b8-d02e419b7ef5",
							},
						},
					},
				},
			},
		},
		{
			name: "field value update based on other field and delete",
			configTableUpdates: map[string][]ConfigTableUpdates{
				"< 2.8.0": {
					{
						Name: "plugin_1",
						Type: Plugin,
						FieldUpdates: []ConfigTableFieldCondition{
							{
								Field:     "plugin_field_1",
								Condition: "plugin_field_1",
								Updates: []ConfigTableFieldUpdate{
									{
										Field:          "plugin_field_2",
										ValueFromField: "plugin_field_1",
									},
									{
										Field: "plugin_field_1",
									},
								},
							},
						},
						ChangeID: "T101",
					},
				},
			},
			uncompressedPayload: `{
				"config_table": {
					"plugins": [
						{
							"id": "47e46c41-e781-49d1-b4b8-d02e419b7ef5",
							"name": "plugin_1",
							"config": {
								"plugin_field_1": "plugin_field_1_value"
								"plugin_field_2": "value"
							}
						}
					]
				}
			}`,
			dataPlaneVersion: "2.7.0",
			expectedPayload: `{
				"config_table": {
					"plugins": [
						{
							"id": "47e46c41-e781-49d1-b4b8-d02e419b7ef5",
							"name": "plugin_1",
							"config": {
								"plugin_field_2": "plugin_field_1_value"
							}
						}
					]
				}
			}`,
			expectedChanges: TrackedChanges{
				ChangeDetails: []ChangeDetail{
					{
						ID: "T101",
						Resources: []ResourceInfo{
							{
								Type: "plugin",
								ID:   "47e46c41-e781-49d1-b4b8-d02e419b7ef5",
							},
						},
					},
				},
			},
		},
		{
			name: "field value based on non-existing field; ensure no change",
			configTableUpdates: map[string][]ConfigTableUpdates{
				"< 2.8.0": {
					{
						Name: "plugin_1",
						Type: Plugin,
						FieldUpdates: []ConfigTableFieldCondition{
							{
								Field:     "plugin_field_1",
								Condition: "plugin_field_1",
								Updates: []ConfigTableFieldUpdate{
									{
										Field:          "plugin_field_2",
										ValueFromField: "plugin_field_2",
									},
								},
							},
						},
						ChangeID: "T101",
					},
				},
			},
			uncompressedPayload: `{
				"config_table": {
					"plugins": [
						{
							"id": "47e46c41-e781-49d1-b4b8-d02e419b7ef5",
							"name": "plugin_1",
							"config": {
								"plugin_field_1": "value"
							}
						}
					]
				}
			}`,
			dataPlaneVersion: "2.7.0",
			expectedPayload: `{
				"config_table": {
					"plugins": [
						{
							"id": "47e46c41-e781-49d1-b4b8-d02e419b7ef5",
							"name": "plugin_1",
							"config": {
								"plugin_field_1": "value"
							}
						}
					]
				}
			}`,
			expectedChanges: TrackedChanges{},
		},
		{
			name: "existing plugin to be removed is removed",
			configTableUpdates: map[string][]ConfigTableUpdates{
				"< 3.0.0": {
					{
						Name:     "plugin_1",
						Type:     Plugin,
						Remove:   true,
						ChangeID: "T101",
					},
				},
			},
			uncompressedPayload: `{
				"config_table": {
					"plugins": [
						{
							"id": "2f303641-37dd-4757-a189-bdebe357fd23",
							"name": "plugin_1",
							"config": {
								"plugin_field_1": "value"
							}
						},
						{
							"id": "ab3b5a6d-923e-4e71-83b4-77e4b68d3e55",
							"name": "plugin_2",
							"config": {
								"plugin_field_2": "value"
							}
						}
					]
				}
			}`,
			dataPlaneVersion: "2.7.0",
			expectedPayload: `{
				"config_table": {
					"plugins": [
						{
							"id": "ab3b5a6d-923e-4e71-83b4-77e4b68d3e55",
							"name": "plugin_2",
							"config": {
								"plugin_field_2": "value"
							}
						}
					]
				}
			}`,
			expectedChanges: TrackedChanges{
				ChangeDetails: []ChangeDetail{
					{
						ID: "T101",
						Resources: []ResourceInfo{
							{
								Type: "plugin",
								ID:   "2f303641-37dd-4757-a189-bdebe357fd23",
							},
						},
					},
				},
			},
		},
		{
			name: "multiple existing plugins to be removed are removed",
			configTableUpdates: map[string][]ConfigTableUpdates{
				"< 3.0.0": {
					{
						Name:     "plugin_1",
						Type:     Plugin,
						Remove:   true,
						ChangeID: "T101",
					},
					{
						Name:     "plugin_2",
						Type:     Plugin,
						Remove:   true,
						ChangeID: "T102",
					},
				},
			},
			uncompressedPayload: `{
				"config_table": {
					"plugins": [
						{
							"id": "5441b100-f441-4d4b-bcc2-3bb153e2bd40",
							"name": "plugin_1",
							"config": {
								"plugin_field_1": "value"
							}
						},
						{
							"id": "5441b100-f441-4d4b-bcc2-3bb153e2bd41",
							"name": "plugin_2",
							"config": {
								"plugin_field_2": "value"
							}
						},
						{
							"id": "5441b100-f441-4d4b-bcc2-3bb153e2bd43",
							"name": "plugin_3",
							"config": {
								"plugin_field_3": "value"
							}
						}
					]
				}
			}`,
			dataPlaneVersion: "2.7.0",
			expectedPayload: `{
				"config_table": {
					"plugins": [
						{
							"id": "5441b100-f441-4d4b-bcc2-3bb153e2bd43",
							"name": "plugin_3",
							"config": {
								"plugin_field_3": "value"
							}
						}
					]
				}
			}`,
			expectedChanges: TrackedChanges{
				ChangeDetails: []ChangeDetail{
					{
						ID: "T101",
						Resources: []ResourceInfo{
							{
								Type: "plugin",
								ID:   "5441b100-f441-4d4b-bcc2-3bb153e2bd40",
							},
						},
					},
					{
						ID: "T102",
						Resources: []ResourceInfo{
							{
								Type: "plugin",
								ID:   "5441b100-f441-4d4b-bcc2-3bb153e2bd41",
							},
						},
					},
				},
			},
		},
		{
			name: "all existing plugins to be removed are removed",
			configTableUpdates: map[string][]ConfigTableUpdates{
				"< 3.0.0": {
					{
						Name:     "plugin_1",
						Type:     Plugin,
						Remove:   true,
						ChangeID: "T101",
					},
					{
						Name:     "plugin_2",
						Type:     Plugin,
						Remove:   true,
						ChangeID: "T102",
					},
				},
			},
			uncompressedPayload: `{
				"config_table": {
					"plugins": [
						{
							"id": "5441b100-f441-4d4b-bcc2-3bb153e2bd40",
							"name": "plugin_1",
							"config": {
								"plugin_field_1": "value"
							}
						},
						{
							"id": "5441b100-f441-4d4b-bcc2-3bb153e2bd41",
							"name": "plugin_2",
							"config": {
								"plugin_field_2": "value"
							}
						}
					]
				}
			}`,
			dataPlaneVersion: "2.7.0",
			expectedPayload: `{
				"config_table": {
					"plugins": []
				}
			}`,
			expectedChanges: TrackedChanges{
				ChangeDetails: []ChangeDetail{
					{
						ID: "T101",
						Resources: []ResourceInfo{
							{
								Type: "plugin",
								ID:   "5441b100-f441-4d4b-bcc2-3bb153e2bd40",
							},
						},
					},
					{
						ID: "T102",
						Resources: []ResourceInfo{
							{
								Type: "plugin",
								ID:   "5441b100-f441-4d4b-bcc2-3bb153e2bd41",
							},
						},
					},
				},
			},
		},
		{
			name: "existing plugin not to be removed is not removed",
			configTableUpdates: map[string][]ConfigTableUpdates{
				"< 3.0.0": {
					{
						Name:     "plugin_1",
						Type:     Plugin,
						Remove:   true,
						ChangeID: "T101",
					},
				},
			},
			uncompressedPayload: `{
				"config_table": {
					"plugins": [
						{
							"id": "5441b100-f441-4d4b-bcc2-3bb153e2bd41",
							"name": "plugin_1",
							"config": {
								"plugin_field_1": "value"
							}
						}
					]
				}
			}`,
			dataPlaneVersion: "3.0.0",
			expectedPayload: `{
				"config_table": {
					"plugins": [
						{
							"id": "5441b100-f441-4d4b-bcc2-3bb153e2bd41",
							"name": "plugin_1",
							"config": {
								"plugin_field_1": "value"
							}
						}
					]
				}
			}`,
			expectedChanges: TrackedChanges{},
		},
		{
			name: "ensure multiple plugins are removed and process field updates occur for multiple configured plugins",
			configTableUpdates: map[string][]ConfigTableUpdates{
				"< 3.0.0": {
					{
						Name:     "plugin_1",
						Type:     Plugin,
						Remove:   true,
						ChangeID: "T101",
					},
					{
						Name: "plugin_2",
						Type: Plugin,
						FieldUpdates: []ConfigTableFieldCondition{
							{
								Field:     "plugin_2_field_1",
								Condition: "plugin_2_field_1",
								Updates: []ConfigTableFieldUpdate{
									{
										Field:          "plugin_2_field_2",
										ValueFromField: "plugin_2_field_1",
									},
								},
							},
						},
						ChangeID: "T102",
					},
				},
			},
			uncompressedPayload: `{
				"config_table": {
					"plugins": [
						{
							"id": "5441b100-f441-4d4b-bcc2-3bb153e2bd41",
							"name": "plugin_2",
							"config": {
								"plugin_2_field_1": "plugin_2_field_1_value",
								"plugin_2_field_2": "value"
							}
						},
						{
							"id": "5441b100-f441-4d4b-bcc2-3bb153e2bd42",
							"name": "plugin_1",
							"config": {
								"plugin_1_field_1": "element"
							}
						},
						{
							"id": "5441b100-f441-4d4b-bcc2-3bb153e2bd43",
							"name": "plugin_3",
							"config": {
								"plugin_3_field_1": "element"
							}
						},
						{
							"id": "5441b100-f441-4d4b-bcc2-3bb153e2bd44",
							"name": "plugin_1",
							"config": {
								"plugin_3_field_1": "element"
							}
						},
						{
							"id": "5441b100-f441-4d4b-bcc2-3bb153e2bd45",
							"name": "plugin_3",
							"config": {
								"plugin_3_field_1": "element"
							}
						}
					]
				}
			}`,
			dataPlaneVersion: "2.7.0",
			expectedPayload: `{
				"config_table": {
					"plugins": [
						{
							"id": "5441b100-f441-4d4b-bcc2-3bb153e2bd41",
							"name": "plugin_2",
							"config": {
								"plugin_2_field_1": "plugin_2_field_1_value",
								"plugin_2_field_2": "plugin_2_field_1_value"
							}
						},
						{
							"id": "5441b100-f441-4d4b-bcc2-3bb153e2bd43",
							"name": "plugin_3",
							"config": {
								"plugin_3_field_1": "element"
							}
						},
						{
							"id": "5441b100-f441-4d4b-bcc2-3bb153e2bd45",
							"name": "plugin_3",
							"config": {
								"plugin_3_field_1": "element"
							}
						}
					]
				}
			}`,
			expectedChanges: TrackedChanges{
				ChangeDetails: []ChangeDetail{
					{
						ID: "T101",
						Resources: []ResourceInfo{
							{
								Type: "plugin",
								ID:   "5441b100-f441-4d4b-bcc2-3bb153e2bd42",
							},
							{
								Type: "plugin",
								ID:   "5441b100-f441-4d4b-bcc2-3bb153e2bd44",
							},
						},
					},
					{
						ID: "T102",
						Resources: []ResourceInfo{
							{
								Type: "plugin",
								ID:   "5441b100-f441-4d4b-bcc2-3bb153e2bd41",
							},
						},
					},
				},
			},
		},
		{
			name: "drop single service field",
			configTableUpdates: map[string][]ConfigTableUpdates{
				"< 3.0.0": {
					{
						Name: Service.String(),
						Type: Service,
						RemoveFields: []string{
							"service_field_1",
						},
						ChangeID: "T101",
					},
				},
			},
			uncompressedPayload: `{
				"config_table": {
					"services": [
						{
							"id": "5441b100-f441-4d4b-bcc2-3bb153e2bd41",
							"name": "service_1",
							"service_field_1": "element",
							"service_field_2": "element"
						}
					]
				}
			}`,
			dataPlaneVersion: "2.7.0",
			expectedPayload: `{
				"config_table": {
					"services": [
						{
							"id": "5441b100-f441-4d4b-bcc2-3bb153e2bd41",
							"name": "service_1",
							"service_field_2": "element"
						}
					]
				}
			}`,
			expectedChanges: TrackedChanges{
				ChangeDetails: []ChangeDetail{
					{
						ID: "T101",
						Resources: []ResourceInfo{
							{
								Type: "service",
								ID:   "5441b100-f441-4d4b-bcc2-3bb153e2bd41",
							},
						},
					},
				},
			},
		},
		{
			name: "drop multiple service fields",
			configTableUpdates: map[string][]ConfigTableUpdates{
				"< 3.0.0": {
					{
						Name: Service.String(),
						Type: Service,
						RemoveFields: []string{
							"service_field_1",
							"service_field_2",
						},
						ChangeID: "T101",
					},
				},
			},
			uncompressedPayload: `{
				"config_table": {
					"services": [
						{

							"id": "f0a3858b-e411-4b56-b415-b8018ac92369",
							"name": "service_1",
							"service_field_1": "element",
							"service_field_2": "element",
							"service_field_3": "element"
						}
					]
				}
			}`,
			dataPlaneVersion: "2.7.0",
			expectedPayload: `{
				"config_table": {
					"services": [
						{
							"id": "f0a3858b-e411-4b56-b415-b8018ac92369",
							"name": "service_1",
							"service_field_3": "element"
						}
					]
				}
			}`,
			expectedChanges: TrackedChanges{
				ChangeDetails: []ChangeDetail{
					{
						ID: "T101",
						Resources: []ResourceInfo{
							{
								Type: "service",
								ID:   "f0a3858b-e411-4b56-b415-b8018ac92369",
							},
						},
					},
				},
			},
		},
		{
			name: "drop multiple service fields from multiple services",
			configTableUpdates: map[string][]ConfigTableUpdates{
				"< 3.0.0": {
					{
						Name: Service.String(),
						Type: Service,
						RemoveFields: []string{
							"service_field_1",
							"service_field_2",
							"service_field_4",
						},
						ChangeID: "T101",
					},
				},
			},
			uncompressedPayload: `{
				"config_table": {
					"services": [
						{
							"id": "47e46c41-e781-49d1-b4b8-d02e419b7ef5",
							"name": "service_1",
							"service_field_1": "element",
							"service_field_2": "element",
							"service_field_3": "element"
						},
						{
							"id": "47e46c41-e781-49d1-b4b8-d02e419b7ef6",
							"name": "service_2",
							"service_field_1": "element",
							"service_field_3": "element",
							"service_field_4": "element"
						}
					]
				}
			}`,
			dataPlaneVersion: "2.7.0",
			expectedPayload: `{
				"config_table": {
					"services": [
						{
							"id": "47e46c41-e781-49d1-b4b8-d02e419b7ef5",
							"name": "service_1",
							"service_field_3": "element"
						},
						{
							"id": "47e46c41-e781-49d1-b4b8-d02e419b7ef6",
							"name": "service_2",
							"service_field_3": "element"
						}
					]
				}
			}`,
			expectedChanges: TrackedChanges{
				ChangeDetails: []ChangeDetail{
					{
						ID: "T101",
						Resources: []ResourceInfo{
							{
								Type: "service",
								ID:   "47e46c41-e781-49d1-b4b8-d02e419b7ef5",
							},
							{
								Type: "service",
								ID:   "47e46c41-e781-49d1-b4b8-d02e419b7ef6",
							},
						},
					},
				},
			},
		},
		{
			name: "drop single plugin core field",
			configTableUpdates: map[string][]ConfigTableUpdates{
				"< 3.0.0": {
					{
						Name: CorePlugin.String(),
						Type: CorePlugin,
						RemoveFields: []string{
							"plugin_field_1",
						},
						ChangeID: "T101",
					},
				},
			},
			uncompressedPayload: `{
				"config_table": {
					"plugins": [
						{
							"id": "5441b100-f441-4d4b-bcc2-3bb153e2bd41",
							"name": "plugin_1",
							"plugin_field_1": "element",
							"plugin_field_2": "element"
						}
					]
				}
			}`,
			dataPlaneVersion: "2.7.0",
			expectedPayload: `{
				"config_table": {
					"plugins": [
						{
							"id": "5441b100-f441-4d4b-bcc2-3bb153e2bd41",
							"name": "plugin_1",
							"plugin_field_2": "element"
						}
					]
				}
			}`,
			expectedChanges: TrackedChanges{
				ChangeDetails: []ChangeDetail{
					{
						ID: "T101",
						Resources: []ResourceInfo{
							{
								Type: "plugin",
								ID:   "5441b100-f441-4d4b-bcc2-3bb153e2bd41",
							},
						},
					},
				},
			},
		},
		{
			name: "drop multiple plugin core fields",
			configTableUpdates: map[string][]ConfigTableUpdates{
				"< 3.0.0": {
					{
						Name: CorePlugin.String(),
						Type: CorePlugin,
						RemoveFields: []string{
							"plugin_field_1",
							"plugin_field_2",
						},
						ChangeID: "T101",
					},
				},
			},
			uncompressedPayload: `{
				"config_table": {
					"plugins": [
						{

							"id": "f0a3858b-e411-4b56-b415-b8018ac92369",
							"name": "plugin_1",
							"plugin_field_1": "element",
							"plugin_field_2": "element",
							"plugin_field_3": "element"
						}
					]
				}
			}`,
			dataPlaneVersion: "2.7.0",
			expectedPayload: `{
				"config_table": {
					"plugins": [
						{
							"id": "f0a3858b-e411-4b56-b415-b8018ac92369",
							"name": "plugin_1",
							"plugin_field_3": "element"
						}
					]
				}
			}`,
			expectedChanges: TrackedChanges{
				ChangeDetails: []ChangeDetail{
					{
						ID: "T101",
						Resources: []ResourceInfo{
							{
								Type: "plugin",
								ID:   "f0a3858b-e411-4b56-b415-b8018ac92369",
							},
						},
					},
				},
			},
		},
		{
			name: "drop multiple plugins core fields from multiple plugins",
			configTableUpdates: map[string][]ConfigTableUpdates{
				"< 3.0.0": {
					{
						Name: CorePlugin.String(),
						Type: CorePlugin,
						RemoveFields: []string{
							"plugin_field_1",
							"plugin_field_2",
							"plugin_field_4",
						},
						ChangeID: "T101",
					},
				},
			},
			uncompressedPayload: `{
				"config_table": {
					"plugins": [
						{
							"id": "47e46c41-e781-49d1-b4b8-d02e419b7ef5",
							"name": "plugin_1",
							"plugin_field_1": "element",
							"plugin_field_2": "element",
							"plugin_field_3": "element"
						},
						{
							"id": "47e46c41-e781-49d1-b4b8-d02e419b7ef6",
							"name": "plugin_2",
							"plugin_field_1": "element",
							"plugin_field_3": "element",
							"plugin_field_4": "element"
						}
					]
				}
			}`,
			dataPlaneVersion: "2.7.0",
			expectedPayload: `{
				"config_table": {
					"plugins": [
						{
							"id": "47e46c41-e781-49d1-b4b8-d02e419b7ef5",
							"name": "plugin_1",
							"plugin_field_3": "element"
						},
						{
							"id": "47e46c41-e781-49d1-b4b8-d02e419b7ef6",
							"name": "plugin_2",
							"plugin_field_3": "element"
						}
					]
				}
			}`,
			expectedChanges: TrackedChanges{
				ChangeDetails: []ChangeDetail{
					{
						ID: "T101",
						Resources: []ResourceInfo{
							{
								Type: "plugin",
								ID:   "47e46c41-e781-49d1-b4b8-d02e419b7ef5",
							},
							{
								Type: "plugin",
								ID:   "47e46c41-e781-49d1-b4b8-d02e419b7ef6",
							},
						},
					},
				},
			},
		},
		{
			name: "drop single route field",
			configTableUpdates: map[string][]ConfigTableUpdates{
				"< 3.0.0": {
					{
						Name: Route.String(),
						Type: Route,
						RemoveFields: []string{
							"route_field_1",
						},
						ChangeID: "T101",
					},
				},
			},
			uncompressedPayload: `{
				"config_table": {
					"routes": [
						{
							"id": "5441b100-f441-4d4b-bcc2-3bb153e2bd41",
							"name": "route_1",
							"route_field_1": "element",
							"route_field_2": "element"
						}
					]
				}
			}`,
			dataPlaneVersion: "2.7.0",
			expectedPayload: `{
				"config_table": {
					"routes": [
						{
							"id": "5441b100-f441-4d4b-bcc2-3bb153e2bd41",
							"name": "route_1",
							"route_field_2": "element"
						}
					]
				}
			}`,
			expectedChanges: TrackedChanges{
				ChangeDetails: []ChangeDetail{
					{
						ID: "T101",
						Resources: []ResourceInfo{
							{
								Type: "route",
								ID:   "5441b100-f441-4d4b-bcc2-3bb153e2bd41",
							},
						},
					},
				},
			},
		},
		{
			name: "drop multiple route fields",
			configTableUpdates: map[string][]ConfigTableUpdates{
				"< 3.0.0": {
					{
						Name: Route.String(),
						Type: Route,
						RemoveFields: []string{
							"route_field_1",
							"route_field_2",
						},
						ChangeID: "T101",
					},
				},
			},
			uncompressedPayload: `{
				"config_table": {
					"routes": [
						{

							"id": "f0a3858b-e411-4b56-b415-b8018ac92369",
							"name": "route_1",
							"route_field_1": "element",
							"route_field_2": "element",
							"route_field_3": "element"
						}
					]
				}
			}`,
			dataPlaneVersion: "2.7.0",
			expectedPayload: `{
				"config_table": {
					"routes": [
						{
							"id": "f0a3858b-e411-4b56-b415-b8018ac92369",
							"name": "route_1",
							"route_field_3": "element"
						}
					]
				}
			}`,
			expectedChanges: TrackedChanges{
				ChangeDetails: []ChangeDetail{
					{
						ID: "T101",
						Resources: []ResourceInfo{
							{
								Type: "route",
								ID:   "f0a3858b-e411-4b56-b415-b8018ac92369",
							},
						},
					},
				},
			},
		},
		{
			name: "drop multiple route fields from multiple routes",
			configTableUpdates: map[string][]ConfigTableUpdates{
				"< 3.0.0": {
					{
						Name: Route.String(),
						Type: Route,
						RemoveFields: []string{
							"route_field_1",
							"route_field_2",
							"route_field_4",
						},
						ChangeID: "T101",
					},
				},
			},
			uncompressedPayload: `{
				"config_table": {
					"routes": [
						{
							"id": "47e46c41-e781-49d1-b4b8-d02e419b7ef5",
							"name": "route_1",
							"route_field_1": "element",
							"route_field_2": "element",
							"route_field_3": "element"
						},
						{
							"id": "47e46c41-e781-49d1-b4b8-d02e419b7ef6",
							"name": "route_2",
							"route_field_1": "element",
							"route_field_3": "element",
							"route_field_4": "element"
						}
					]
				}
			}`,
			dataPlaneVersion: "2.7.0",
			expectedPayload: `{
				"config_table": {
					"routes": [
						{
							"id": "47e46c41-e781-49d1-b4b8-d02e419b7ef5",
							"name": "route_1",
							"route_field_3": "element"
						},
						{
							"id": "47e46c41-e781-49d1-b4b8-d02e419b7ef6",
							"name": "route_2",
							"route_field_3": "element"
						}
					]
				}
			}`,
			expectedChanges: TrackedChanges{
				ChangeDetails: []ChangeDetail{
					{
						ID: "T101",
						Resources: []ResourceInfo{
							{
								Type: "route",
								ID:   "47e46c41-e781-49d1-b4b8-d02e419b7ef5",
							},
							{
								Type: "route",
								ID:   "47e46c41-e781-49d1-b4b8-d02e419b7ef6",
							},
						},
					},
				},
			},
		},
		{
			name: "drop services, routes and plugins' fields",
			configTableUpdates: map[string][]ConfigTableUpdates{
				"< 3.0.0": {
					{
						Name: Service.String(),
						Type: Service,
						RemoveFields: []string{
							"service_field_1",
							"service_field_2",
							"service_field_4",
						},
						ChangeID: "T101",
					},
					{
						Name:     "plugin_1",
						Type:     Plugin,
						Remove:   true,
						ChangeID: "T102",
					},
					{
						Name: CorePlugin.String(),
						Type: CorePlugin,
						RemoveFields: []string{
							"core_plugin_field_1",
						},
						ChangeID: "T103",
					},
					{
						Name: Route.String(),
						Type: Route,
						RemoveFields: []string{
							"route_field_1",
							"route_field_2",
							"route_field_4",
						},
						ChangeID: "T104",
					},
				},
			},
			uncompressedPayload: `{
				"config_table": {
					"plugins": [
						{
							"id": "47e46c41-e781-49d1-b4b8-d02e419b7ef6",
							"name": "plugin_2",
							"config": {
								"plugin_2_field_1": "element"
							},
							"core_plugin_field_1": "value"
						},
						{
							"id": "5441b100-f441-4d4b-bcc2-3bb153e2bd40",
							"name": "plugin_1",
							"config": {
								"plugin_1_field_1": "element"
							},
							"core_plugin_field_1": "value"
						},
						{
							"id": "2f303641-37dd-4757-a189-bdebe357fd23",
							"name": "plugin_3",
							"config": {
								"plugin_3_field_1": "element"
							}
						},
						{
							"id": "ab3b5a6d-923e-4e71-83b4-77e4b68d3e55",
							"name": "plugin_1",
							"config": {
								"plugin_1_field_1": "element"
							}
						}
					],
					"services": [
						{
							"id": "af1ab105-c9a3-42c8-9b12-442e3fd87a7f",
							"name": "service_1",
							"service_field_1": "element",
							"service_field_2": "element",
							"service_field_3": "element"
						},
						{
							"id": "c50e912c-873c-45da-9f7c-f12a19fd56d1",
							"name": "service_2",
							"service_field_1": "element",
							"service_field_3": "element",
							"service_field_4": "element"
						}
					],
					"routes": [
						{
							"id": "bbbba698-1fae-11ed-861d-0242ac120002",
							"name": "route_1",
							"route_field_1": "element",
							"route_field_2": "element",
							"route_field_3": "element"
						},
						{
							"id": "c1460cc0-1fae-11ed-861d-0242ac120002",
							"name": "route_2",
							"route_field_1": "element",
							"route_field_3": "element",
							"route_field_4": "element"
						}
					]
				}
			}`,
			dataPlaneVersion: "2.7.0",
			expectedPayload: `{
				"config_table": {
					"plugins": [
						{
							"id": "47e46c41-e781-49d1-b4b8-d02e419b7ef6",
							"name": "plugin_2",
							"config": {
								"plugin_2_field_1": "element"
							}
						},
						{
							"id": "2f303641-37dd-4757-a189-bdebe357fd23",
							"name": "plugin_3",
							"config": {
								"plugin_3_field_1": "element"
							}
						}
					],
					"services": [
						{
							"id": "af1ab105-c9a3-42c8-9b12-442e3fd87a7f",
							"name": "service_1",
							"service_field_3": "element"
						},
						{
							"id": "c50e912c-873c-45da-9f7c-f12a19fd56d1",
							"name": "service_2",
							"service_field_3": "element"
						}
					],
					"routes": [
						{
							"id": "bbbba698-1fae-11ed-861d-0242ac120002",
							"name": "route_1",
							"route_field_3": "element"
						},
						{
							"id": "c1460cc0-1fae-11ed-861d-0242ac120002",
							"name": "route_2",
							"route_field_3": "element"
						}
					]
				}
			}`,
			expectedChanges: TrackedChanges{
				ChangeDetails: []ChangeDetail{
					{
						ID: "T101",
						Resources: []ResourceInfo{
							{
								Type: "service",
								ID:   "af1ab105-c9a3-42c8-9b12-442e3fd87a7f",
							},
							{
								Type: "service",
								ID:   "c50e912c-873c-45da-9f7c-f12a19fd56d1",
							},
						},
					},
					{
						ID: "T102",
						Resources: []ResourceInfo{
							{
								Type: "plugin",
								ID:   "5441b100-f441-4d4b-bcc2-3bb153e2bd40",
							},
							{
								Type: "plugin",
								ID:   "ab3b5a6d-923e-4e71-83b4-77e4b68d3e55",
							},
						},
					},
					{
						ID: "T103",
						Resources: []ResourceInfo{
							{
								Type: "plugin",
								ID:   "47e46c41-e781-49d1-b4b8-d02e419b7ef6",
							},
						},
					},
					{
						ID: "T104",
						Resources: []ResourceInfo{
							{
								Type: "route",
								ID:   "bbbba698-1fae-11ed-861d-0242ac120002",
							},
							{
								Type: "route",
								ID:   "c1460cc0-1fae-11ed-861d-0242ac120002",
							},
						},
					},
				},
			},
		},
		{
			name: "ensure plugin field is removed because of newer version",
			configTableUpdates: map[string][]ConfigTableUpdates{
				">= 3.0.0": {
					{
						Name: "plugin_1",
						Type: Plugin,
						RemoveFields: []string{
							"plugin_1_field_1",
						},
						ChangeID: "T101",
					},
				},
			},
			uncompressedPayload: `{
				"config_table": {
					"plugins": [
						{
							"id": "c50e912c-873c-45da-9f7c-f12a19fd56d1",
							"name": "plugin_2",
							"config": {
								"plugin_2_field_1": "element"
							}
						},
						{
							"id": "f239e435-b1fa-4f0f-8d0c-6c317b8ec606",
							"name": "plugin_1",
							"config": {
								"plugin_1_field_1": "element",
								"plugin_1_field_2": "element"
							}
						},
						{
							"id": "b79f9024-e2b5-43b0-b16b-28d14c3bca90y",
							"name": "plugin_3",
							"config": {
								"plugin_3_field_1": "element"
							}
						}
					]
				}
			}`,
			dataPlaneVersion: "3.0.0",
			expectedPayload: `{
				"config_table": {
					"plugins": [
						{
							"id": "c50e912c-873c-45da-9f7c-f12a19fd56d1",
							"name": "plugin_2",
							"config": {
								"plugin_2_field_1": "element"
							}
						},
						{
							"id": "f239e435-b1fa-4f0f-8d0c-6c317b8ec606",
							"name": "plugin_1",
							"config": {
								"plugin_1_field_2": "element"
							}
						},
						{
							"id": "b79f9024-e2b5-43b0-b16b-28d14c3bca90y",
							"name": "plugin_3",
							"config": {
								"plugin_3_field_1": "element"
							}
						}
					]
				}
			}`,
			expectedChanges: TrackedChanges{
				ChangeDetails: []ChangeDetail{
					{
						ID: "T101",
						Resources: []ResourceInfo{
							{
								Type: "plugin",
								ID:   "f239e435-b1fa-4f0f-8d0c-6c317b8ec606",
							},
						},
					},
				},
			},
		},
		{
			name: "ensure plugin field is removed because of newer version (enterprise format)",
			configTableUpdates: map[string][]ConfigTableUpdates{
				">= 3.0.0.0": {
					{
						Name: "plugin_1",
						Type: Plugin,
						RemoveFields: []string{
							"plugin_1_field_1",
						},
						ChangeID: "T102",
					},
				},
			},
			uncompressedPayload: `{
				"config_table": {
					"plugins": [
						{
							"id": "f239e435-b1fa-4f0f-8d0c-6c317b8ec606",
							"name": "plugin_2",
							"config": {
								"plugin_2_field_1": "element"
							}
						},
						{
							"id": "b79f9024-e2b5-43b0-b16b-28d14c3bca90y",
							"name": "plugin_1",
							"config": {
								"plugin_1_field_1": "element",
								"plugin_1_field_2": "element"
							}
						},
						{
							"id": "c50e912c-873c-45da-9f7c-f12a19fd56d1",
							"name": "plugin_3",
							"config": {
								"plugin_3_field_1": "element"
							}
						}
					]
				}
			}`,
			dataPlaneVersion: "3.0.0.0",
			expectedPayload: `{
				"config_table": {
					"plugins": [
						{
							"id": "f239e435-b1fa-4f0f-8d0c-6c317b8ec606",
							"name": "plugin_2",
							"config": {
								"plugin_2_field_1": "element"
							}
						},
						{
							"id": "b79f9024-e2b5-43b0-b16b-28d14c3bca90y",
							"name": "plugin_1",
							"config": {
								"plugin_1_field_2": "element"
							}
						},
						{
							"id": "c50e912c-873c-45da-9f7c-f12a19fd56d1",
							"name": "plugin_3",
							"config": {
								"plugin_3_field_1": "element"
							}
						}
					]
				}
			}`,
			expectedChanges: TrackedChanges{
				ChangeDetails: []ChangeDetail{
					{
						ID: "T102",
						Resources: []ResourceInfo{
							{
								Type: "plugin",
								ID:   "b79f9024-e2b5-43b0-b16b-28d14c3bca90y",
							},
						},
					},
				},
			},
		},
		{
			name: "ensure plugin field is removed because of older version (enterprise format)",
			configTableUpdates: map[string][]ConfigTableUpdates{
				"< 3.0.0.0": {
					{
						Name: "plugin_1",
						Type: Plugin,
						RemoveFields: []string{
							"plugin_1_field_1",
						},
						ChangeID: "T103",
					},
				},
			},
			uncompressedPayload: `{
				"config_table": {
					"plugins": [
						{
							"id": "f239e435-b1fa-4f0f-8d0c-6c317b8ec606",
							"name": "plugin_2",
							"config": {
								"plugin_2_field_1": "element"
							}
						},
						{
							"id": "b79f9024-e2b5-43b0-b16b-28d14c3bca90y",
							"name": "plugin_1",
							"config": {
								"plugin_1_field_1": "element",
								"plugin_1_field_2": "element"
							}
						},
						{
							"id": "c50e912c-873c-45da-9f7c-f12a19fd56d1",
							"name": "plugin_3",
							"config": {
								"plugin_3_field_1": "element"
							}
						}
					]
				}
			}`,
			dataPlaneVersion: "2.8.0.0",
			expectedPayload: `{
				"config_table": {
					"plugins": [
						{
							"id": "f239e435-b1fa-4f0f-8d0c-6c317b8ec606",
							"name": "plugin_2",
							"config": {
								"plugin_2_field_1": "element"
							}
						},
						{
							"id": "b79f9024-e2b5-43b0-b16b-28d14c3bca90y",
							"name": "plugin_1",
							"config": {
								"plugin_1_field_2": "element"
							}
						},
						{
							"id": "c50e912c-873c-45da-9f7c-f12a19fd56d1",
							"name": "plugin_3",
							"config": {
								"plugin_3_field_1": "element"
							}
						}
					]
				}
			}`,
			expectedChanges: TrackedChanges{
				ChangeDetails: []ChangeDetail{
					{
						ID: "T103",
						Resources: []ResourceInfo{
							{
								Type: "plugin",
								ID:   "b79f9024-e2b5-43b0-b16b-28d14c3bca90y",
							},
						},
					},
				},
			},
		},
		{
			name: "ensure changes are not tracked if disabled",
			configTableUpdates: map[string][]ConfigTableUpdates{
				"< 3.0.0": {
					{
						Name: "plugin_1",
						Type: Plugin,
						RemoveFields: []string{
							"plugin_1_field_1",
						},
						ChangeID: "T103",
						DisableChangeTracking: func(_ string) bool {
							return true
						},
					},
				},
			},
			uncompressedPayload: `{
				"config_table": {
					"plugins": [
						{
							"id": "b79f9024-e2b5-43b0-b16b-28d14c3bca90y",
							"name": "plugin_1",
							"config": {
								"plugin_1_field_1": "element",
								"plugin_1_field_2": "element"
							}
						}
					]
				}
			}`,
			dataPlaneVersion: "2.8.0",
			expectedPayload: `{
				"config_table": {
					"plugins": [
						{
							"id": "b79f9024-e2b5-43b0-b16b-28d14c3bca90y",
							"name": "plugin_1",
							"config": {
								"plugin_1_field_2": "element"
							}
						}
					]
				}
			}`,
			expectedChanges: TrackedChanges{},
		},
		{
			name: "single field update with single item",
			configTableUpdates: map[string][]ConfigTableUpdates{
				"< 2.8.0": {
					{
						Name: "service_1",
						Type: Service,
						FieldUpdates: []ConfigTableFieldCondition{
							{
								Field:     "service_field_1",
								Condition: "service_field_1=condition",
								Updates: []ConfigTableFieldUpdate{
									{
										Field: "service_field_1",
										Value: "value_updated",
									},
								},
							},
						},
						ChangeID: "T101",
					},
				},
			},
			uncompressedPayload: `{
				"config_table": {
					"services": [
						{
							"id": "c50e912c-873c-45da-9f7c-f12a19fd56d1",
							"name": "service_1",
							"service_field_1": "condition"
						}
					]
				}
			}`,
			dataPlaneVersion: "2.7.0",
			expectedPayload: `{
				"config_table": {
					"services": [
						{
							"id": "c50e912c-873c-45da-9f7c-f12a19fd56d1",
							"name": "service_1",
							"service_field_1": "value_updated"
						}
					]
				}
			}`,
			expectedChanges: TrackedChanges{
				ChangeDetails: []ChangeDetail{
					{
						ID: "T101",
						Resources: []ResourceInfo{
							{
								Type: "service",
								ID:   "c50e912c-873c-45da-9f7c-f12a19fd56d1",
							},
						},
					},
				},
			},
		},
		{
			name: "field updates with multiple data types",
			configTableUpdates: map[string][]ConfigTableUpdates{
				"< 2.8.0": {
					{
						Name: "service_1",
						Type: Service,
						FieldUpdates: []ConfigTableFieldCondition{
							{
								Field:     "service_field_string",
								Condition: "service_field_string=old",
								Updates: []ConfigTableFieldUpdate{
									{
										Field: "service_field_string",
										Value: "new",
									},
								},
							},
							{
								Field:     "service_field_number",
								Condition: "service_field_number=9",
								Updates: []ConfigTableFieldUpdate{
									{
										Field: "service_field_number",
										Value: 28,
									},
								},
							},
							{
								Field:     "service_field_bool",
								Condition: "service_field_bool=true",
								Updates: []ConfigTableFieldUpdate{
									{
										Field: "service_field_bool",
										Value: false,
									},
								},
							},
						},
						ChangeID: "T101",
					},
				},
			},
			uncompressedPayload: `{
				"config_table": {
					"services": [
						{
							"id": "47e46c41-e781-49d1-b4b8-d02e419b7ef5",
							"name": "service_1",
							"service_field_string": "old",
							"service_field_number": 9,
							"service_field_bool": true
						}
					]
				}
			}`,
			dataPlaneVersion: "2.7.0",
			expectedPayload: `{
				"config_table": {
					"services": [
						{
							"id": "47e46c41-e781-49d1-b4b8-d02e419b7ef5",
							"name": "service_1",
							"service_field_string": "new",
							"service_field_number": 28,
							"service_field_bool": false
						}
					]
				}
			}`,
			expectedChanges: TrackedChanges{
				ChangeDetails: []ChangeDetail{
					{
						ID: "T101",
						Resources: []ResourceInfo{
							{
								Type: "service",
								ID:   "47e46c41-e781-49d1-b4b8-d02e419b7ef5",
							},
						},
					},
				},
			},
		},
		{
			name: "field update with multiple value updates",
			configTableUpdates: map[string][]ConfigTableUpdates{
				"< 2.8.0": {
					{
						Name: "service_1",
						Type: Service,
						FieldUpdates: []ConfigTableFieldCondition{
							{
								Field:     "service_field_1",
								Condition: "service_field_1=condition",
								Updates: []ConfigTableFieldUpdate{
									{
										Field: "service_field_1",
										Value: "value_updated",
									},
									{
										Field: "service_field_3",
										Value: "value_updated",
									},
								},
							},
						},
						ChangeID: "T101",
					},
				},
			},
			uncompressedPayload: `{
				"config_table": {
					"services": [
						{
							"id": "47e46c41-e781-49d1-b4b8-d02e419b7ef5",
							"name": "service_1",
							"service_field_1": "condition",
							"service_field_2": "value",
							"service_field_3": "value",
							"service_field_4": "value"
						}
					]
				}
			}`,
			dataPlaneVersion: "2.7.0",
			expectedPayload: `{
				"config_table": {
					"services": [
						{
							"id": "47e46c41-e781-49d1-b4b8-d02e419b7ef5",
							"name": "service_1",
							"service_field_1": "value_updated",
							"service_field_2": "value",
							"service_field_3": "value_updated",
							"service_field_4": "value"
						}
					]
				}
			}`,
			expectedChanges: TrackedChanges{
				ChangeDetails: []ChangeDetail{
					{
						ID: "T101",
						Resources: []ResourceInfo{
							{
								Type: "service",
								ID:   "47e46c41-e781-49d1-b4b8-d02e419b7ef5",
							},
						},
					},
				},
			},
		},
		{
			name: "nested field update",
			configTableUpdates: map[string][]ConfigTableUpdates{
				"< 2.8.0": {
					{
						Name: "service_1",
						Type: Service,
						FieldUpdates: []ConfigTableFieldCondition{
							{
								Field:     "service_field_1.nested_field_1",
								Condition: "service_field_1.nested_field_1=condition",
								Updates: []ConfigTableFieldUpdate{
									{
										Field: "service_field_1.nested_field_1",
										Value: "value_updated",
									},
								},
							},
						},
						ChangeID: "T101",
					},
				},
			},
			uncompressedPayload: `{
				"config_table": {
					"services": [
						{
							"id": "47e46c41-e781-49d1-b4b8-d02e419b7ef5",
							"name": "service_1",
							"service_field_1": {
								"nested_field_1": "condition"
							}
						}
					]
				}
			}`,
			dataPlaneVersion: "2.7.0",
			expectedPayload: `{
				"config_table": {
					"services": [
						{
							"id": "47e46c41-e781-49d1-b4b8-d02e419b7ef5",
							"name": "service_1",
							"service_field_1": {
								"nested_field_1": "value_updated"
							}
						}
					]
				}
			}`,
			expectedChanges: TrackedChanges{
				ChangeDetails: []ChangeDetail{
					{
						ID: "T101",
						Resources: []ResourceInfo{
							{
								Type: "service",
								ID:   "47e46c41-e781-49d1-b4b8-d02e419b7ef5",
							},
						},
					},
				},
			},
		},
		{
			name: "field update with additional nested field update",
			configTableUpdates: map[string][]ConfigTableUpdates{
				"< 2.8.0": {
					{
						Name: "service_1",
						Type: Service,
						FieldUpdates: []ConfigTableFieldCondition{
							{
								Field:     "service_field_1",
								Condition: "service_field_1=condition",
								Updates: []ConfigTableFieldUpdate{
									{
										Field: "service_field_1",
										Value: "value_updated",
									},
									{
										Field: "service_field_2.nested_field_1",
										Value: "value_updated",
									},
								},
							},
						},
						ChangeID: "T101",
					},
				},
			},
			uncompressedPayload: `{
				"config_table": {
					"services": [
						{
							"id": "47e46c41-e781-49d1-b4b8-d02e419b7ef5",
							"name": "service_1",
							"service_field_1": "condition",
							"service_field_2": {
								"nested_field_1": "value"
							}
						}
					]
				}
			}`,
			dataPlaneVersion: "2.7.0",
			expectedPayload: `{
				"config_table": {
					"services": [
						{
							"id": "47e46c41-e781-49d1-b4b8-d02e419b7ef5",
							"name": "service_1",
							"service_field_1": "value_updated",
							"service_field_2": {
								"nested_field_1": "value_updated"
							}
						}
					]
				}
			}`,
			expectedChanges: TrackedChanges{
				ChangeDetails: []ChangeDetail{
					{
						ID: "T101",
						Resources: []ResourceInfo{
							{
								Type: "service",
								ID:   "47e46c41-e781-49d1-b4b8-d02e419b7ef5",
							},
						},
					},
				},
			},
		},
		{
			name: "no field updates",
			configTableUpdates: map[string][]ConfigTableUpdates{
				"< 2.8.0": {
					{
						Name: "service_1",
						Type: Service,
						FieldUpdates: []ConfigTableFieldCondition{
							{
								Field:     "service_field_1",
								Condition: "service_field_1=condition",
								Updates: []ConfigTableFieldUpdate{
									{
										Field: "service_field_1",
										Value: "value_updated",
									},
									{
										Field: "service_field_2",
										Value: "value_updated",
									},
								},
							},
						},
						ChangeID: "T101",
					},
				},
			},
			uncompressedPayload: `{
				"config_table": {
					"services": [
						{
							"id": "47e46c41-e781-49d1-b4b8-d02e419b7ef5",
							"name": "service_1",
							"service_field_1": "value",
							"service_field_2": "condition",
							"service_field_3": "condition",
							"service_field_4": "value"
						}
					]
				}
			}`,
			dataPlaneVersion: "2.7.0",
			expectedPayload: `{
				"config_table": {
					"services": [
						{
							"id": "47e46c41-e781-49d1-b4b8-d02e419b7ef5",
							"name": "service_1",
							"service_field_1": "value",
							"service_field_2": "condition",
							"service_field_3": "condition",
							"service_field_4": "value"
						}
					]
				}
			}`,
			expectedChanges: TrackedChanges{},
		},
		{
			name: "field removal and field update",
			configTableUpdates: map[string][]ConfigTableUpdates{
				"< 2.8.0": {
					{
						Name: "service_1",
						Type: Service,
						RemoveFields: []string{
							"service_1_field_1",
						},
						FieldUpdates: []ConfigTableFieldCondition{
							{
								Field:     "service_1_field_3",
								Condition: "service_1_field_3=condition",
								Updates: []ConfigTableFieldUpdate{
									{
										Field: "service_1_field_3",
										Value: "value_updated",
									},
								},
							},
						},
						ChangeID: "T101",
					},
				},
			},
			uncompressedPayload: `{
				"config_table": {
					"services": [
						{
							"id": "47e46c41-e781-49d1-b4b8-d02e419b7ef5",
							"name": "service_1",
							"service_1_field_1": "element",
							"service_1_field_3": "condition",
							"service_1_field_4": "value"
						}
					]
				}
			}`,
			dataPlaneVersion: "2.7.0",
			expectedPayload: `{
				"config_table": {
					"services": [
						{
							"id": "47e46c41-e781-49d1-b4b8-d02e419b7ef5",
							"name": "service_1",
							"service_1_field_3": "value_updated",
							"service_1_field_4": "value"
						}
					]
				}
			}`,
			expectedChanges: TrackedChanges{
				ChangeDetails: []ChangeDetail{
					{
						ID: "T101",
						Resources: []ResourceInfo{
							{
								Type: "service",
								ID:   "47e46c41-e781-49d1-b4b8-d02e419b7ef5",
							},
						},
					},
				},
			},
		},
		{
			name: "service field value create based on other field and delete",
			configTableUpdates: map[string][]ConfigTableUpdates{
				"< 2.8.0": {
					{
						Name: "service_1",
						Type: Service,
						FieldUpdates: []ConfigTableFieldCondition{
							{
								Field:     "service_field_1",
								Condition: "service_field_1",
								Updates: []ConfigTableFieldUpdate{
									{
										Field:          "service_field_2",
										ValueFromField: "service_field_1",
									},
									{
										Field: "service_field_1",
									},
								},
							},
						},
						ChangeID: "T101",
					},
				},
			},
			uncompressedPayload: `{
				"config_table": {
					"services": [
						{
							"id": "47e46c41-e781-49d1-b4b8-d02e419b7ef5",
							"name": "service_1",
							"service_field_1": "value"
						}
					]
				}
			}`,
			dataPlaneVersion: "2.7.0",
			expectedPayload: `{
				"config_table": {
					"services": [
						{
							"id": "47e46c41-e781-49d1-b4b8-d02e419b7ef5",
							"name": "service_1",
							"service_field_2": "value"
						}
					]
				}
			}`,
			expectedChanges: TrackedChanges{
				ChangeDetails: []ChangeDetail{
					{
						ID: "T101",
						Resources: []ResourceInfo{
							{
								Type: "service",
								ID:   "47e46c41-e781-49d1-b4b8-d02e419b7ef5",
							},
						},
					},
				},
			},
		},
		{
			name: "field value update based on other field and delete",
			configTableUpdates: map[string][]ConfigTableUpdates{
				"< 2.8.0": {
					{
						Name: "service_1",
						Type: Service,
						FieldUpdates: []ConfigTableFieldCondition{
							{
								Field:     "service_field_1",
								Condition: "service_field_1",
								Updates: []ConfigTableFieldUpdate{
									{
										Field:          "service_field_2",
										ValueFromField: "service_field_1",
									},
									{
										Field: "service_field_1",
									},
								},
							},
						},
						ChangeID: "T101",
					},
				},
			},
			uncompressedPayload: `{
				"config_table": {
					"services": [
						{
							"id": "47e46c41-e781-49d1-b4b8-d02e419b7ef5",
							"name": "service_1",
							"service_field_1": "service_field_1_value",
							"service_field_2": "value"
						}
					]
				}
			}`,
			dataPlaneVersion: "2.7.0",
			expectedPayload: `{
				"config_table": {
					"services": [
						{
							"id": "47e46c41-e781-49d1-b4b8-d02e419b7ef5",
							"name": "service_1",
							"service_field_2": "service_field_1_value"
						}
					]
				}
			}`,
			expectedChanges: TrackedChanges{
				ChangeDetails: []ChangeDetail{
					{
						ID: "T101",
						Resources: []ResourceInfo{
							{
								Type: "service",
								ID:   "47e46c41-e781-49d1-b4b8-d02e419b7ef5",
							},
						},
					},
				},
			},
		},
		{
			name: "field value based on non-existing field; ensure no change",
			configTableUpdates: map[string][]ConfigTableUpdates{
				"< 2.8.0": {
					{
						Name: "service_1",
						Type: Service,
						FieldUpdates: []ConfigTableFieldCondition{
							{
								Field:     "service_field_1",
								Condition: "service_field_1",
								Updates: []ConfigTableFieldUpdate{
									{
										Field:          "service_field_2",
										ValueFromField: "service_field_2",
									},
								},
							},
						},
						ChangeID: "T101",
					},
				},
			},
			uncompressedPayload: `{
				"config_table": {
					"services": [
						{
							"id": "47e46c41-e781-49d1-b4b8-d02e419b7ef5",
							"name": "service_1",
							"service_field_1": "value"
						}
					]
				}
			}`,
			dataPlaneVersion: "2.7.0",
			expectedPayload: `{
				"config_table": {
					"services": [
						{
							"id": "47e46c41-e781-49d1-b4b8-d02e419b7ef5",
							"name": "service_1",
							"service_field_1": "value"
						}
					]
				}
			}`,
			expectedChanges: TrackedChanges{},
		},

		{
			name: "single entity field array removal with single item in array",
			configTableUpdates: map[string][]ConfigTableUpdates{
				"< 2.8.0": {
					{
						Name: "route_1",
						Type: Route,
						RemoveElementsFromArray: []ConfigTableFieldCondition{
							{
								Field:     "route_field_1_array_1",
								Condition: "array_element_1=condition",
							},
						},
						ChangeID: "T101",
					},
				},
			},
			uncompressedPayload: `{
				"config_table": {
					"routes": [
						{
							"id": "ab3b5a6d-923e-4e71-83b4-77e4b68d3e55",
							"name": "route_1",
							"route_field_1_array_1": [
								{
									"array_element_1": "condition"
								}
							]
						}
					]
				}
			}`,
			dataPlaneVersion: "2.7.0",
			expectedPayload: `{
				"config_table": {
					"routes": [
						{
							"id": "ab3b5a6d-923e-4e71-83b4-77e4b68d3e55",
							"name": "route_1",
							"route_field_1_array_1": []
						}
					]
				}
			}`,
			expectedChanges: TrackedChanges{
				ChangeDetails: []ChangeDetail{
					{
						ID: "T101",
						Resources: []ResourceInfo{
							{
								Type: "route",
								ID:   "ab3b5a6d-923e-4e71-83b4-77e4b68d3e55",
							},
						},
					},
				},
			},
		},
		{
			name: "single nested entity field array removal with single item in array",
			configTableUpdates: map[string][]ConfigTableUpdates{
				"< 2.8.0": {
					{
						Name: "route_1",
						Type: Route,
						RemoveElementsFromArray: []ConfigTableFieldCondition{
							{
								Field:     "route_field_1_1.route_field_1_array_1",
								Condition: "array_element_1=condition",
							},
						},
						ChangeID: "T101",
					},
				},
			},
			uncompressedPayload: `{
				"config_table": {
					"routes": [
						{
							"id": "29b2b210-3a3a-4344-8208-36c698aa9f5a",
							"name": "route_1",
							"route_field_1_1": {
								"route_field_1_array_1": [
									{
										"array_element_1": "condition"
									}
								]
							}
						}
					]
				}
			}`,
			dataPlaneVersion: "2.7.0",
			expectedPayload: `{
				"config_table": {
					"routes": [
						{
							"id": "29b2b210-3a3a-4344-8208-36c698aa9f5a",
							"name": "route_1",
							"route_field_1_1": {
								"route_field_1_array_1": []
							}
						}
					]
				}
			}`,
			expectedChanges: TrackedChanges{
				ChangeDetails: []ChangeDetail{
					{
						ID: "T101",
						Resources: []ResourceInfo{
							{
								Type: "route",
								ID:   "29b2b210-3a3a-4344-8208-36c698aa9f5a",
							},
						},
					},
				},
			},
		},
		{
			name: "single entity field array removal with multiple items in array",
			configTableUpdates: map[string][]ConfigTableUpdates{
				"< 2.8.0": {
					{
						Name: "route_1",
						Type: Route,
						RemoveElementsFromArray: []ConfigTableFieldCondition{
							{
								Field:     "route_field_1_array_1",
								Condition: "array_element_1=condition",
							},
						},
						ChangeID: "T101",
					},
				},
			},
			uncompressedPayload: `{
				"config_table": {
					"routes": [
						{
							"id": "29b2b210-3a3a-4344-8208-36c698aa9f5a",
							"name": "route_1",
							"route_field_1_array_1": [
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
					]
				}
			}`,
			dataPlaneVersion: "2.7.0",
			expectedPayload: `{
				"config_table": {
					"routes": [
						{
							"id": "29b2b210-3a3a-4344-8208-36c698aa9f5a",
							"name": "route_1",
							"route_field_1_array_1": [
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
					]
				}
			}`,
			expectedChanges: TrackedChanges{
				ChangeDetails: []ChangeDetail{
					{
						ID: "T101",
						Resources: []ResourceInfo{
							{
								Type: "route",
								ID:   "29b2b210-3a3a-4344-8208-36c698aa9f5a",
							},
						},
					},
				},
			},
		},
		{
			name: "entity field and array removal with multiple array removals",
			configTableUpdates: map[string][]ConfigTableUpdates{
				"< 2.8.0": {
					{
						Name: "route_1",
						Type: Route,
						RemoveFields: []string{
							"route_1_field_1",
						},
						RemoveElementsFromArray: []ConfigTableFieldCondition{
							{
								Field:     "route_field_1_array_2",
								Condition: "array_element_3=condition",
							},
						},
						ChangeID: "T101",
					},
				},
			},
			uncompressedPayload: `{
				"config_table": {
					"routes": [
						{
							"id": "29b2b210-3a3a-4344-8208-36c698aa9f5a",
							"name": "route_1",
							"route_1_field_1": "element",
							"route_field_1_array_2": [
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
							"route_1_field_3": "element"
						}
					]
				}
			}`,
			dataPlaneVersion: "2.7.0",
			expectedPayload: `{
				"config_table": {
					"routes": [
						{
							"id": "29b2b210-3a3a-4344-8208-36c698aa9f5a",
							"name": "route_1",
							"route_field_1_array_2": [
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
							"route_1_field_3": "element"
						}
					]
				}
			}`,
			expectedChanges: TrackedChanges{
				ChangeDetails: []ChangeDetail{
					{
						ID: "T101",
						Resources: []ResourceInfo{
							{
								Type: "route",
								ID:   "29b2b210-3a3a-4344-8208-36c698aa9f5a",
							},
						},
					},
				},
			},
		},
		{
			name: "no array removal for entity ",
			configTableUpdates: map[string][]ConfigTableUpdates{
				"< 2.8.0": {
					{
						Name: "route_1",
						Type: Route,
						RemoveFields: []string{
							"route_1_field_1",
						},
						RemoveElementsFromArray: []ConfigTableFieldCondition{
							{
								Field:     "route_field_1_array_2",
								Condition: "array_element_3=condition",
							},
						},
						ChangeID: "T101",
					},
				},
			},
			uncompressedPayload: `{
				"config_table": {
					"routes": [
						{
							"id": "29b2b210-3a3a-4344-8208-36c698aa9f5a",
							"name": "route_1",
							"route_1_field_1": "element",
							"route_field_1_array_2": [
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
							"route_1_field_3": "element"
						}
					]
				}
			}`,
			dataPlaneVersion: "2.7.0",
			expectedPayload: `{
				"config_table": {
					"routes": [
						{
							"id": "29b2b210-3a3a-4344-8208-36c698aa9f5a",
							"name": "route_1",
							"route_field_1_array_2": [
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
							"route_1_field_3": "element"
						}
					]
				}
			}`,
			expectedChanges: TrackedChanges{
				ChangeDetails: []ChangeDetail{
					{
						ID: "T101",
						Resources: []ResourceInfo{
							{
								Type: "route",
								ID:   "29b2b210-3a3a-4344-8208-36c698aa9f5a",
							},
						},
					},
				},
			},
		},
		{
			name: "no entity array removal with multiple versions defined",
			configTableUpdates: map[string][]ConfigTableUpdates{
				"< 2.8.0": {
					{
						Name: "route_1",
						Type: Route,
						RemoveFields: []string{
							"route_1_field_1",
						},
						RemoveElementsFromArray: []ConfigTableFieldCondition{
							{
								Field:     "route_field_1_array_2",
								Condition: "array_element_3=condition",
							},
						},
						ChangeID: "T101",
					},
				},
				"< 2.7.0": {
					{
						Name: "route_1",
						Type: Route,
						RemoveFields: []string{
							"route_1_field_4",
						},
						RemoveElementsFromArray: []ConfigTableFieldCondition{
							{
								Field:     "route_field_1_array_2",
								Condition: "array_element_2=condition",
							},
							{
								Field:     "route_field_1_array_5",
								Condition: "array_element_1=condition",
							},
						},
						ChangeID: "T102",
					},
				},
			},
			uncompressedPayload: `{
				"config_table": {
					"routes": [
						{
							"id": "29b2b210-3a3a-4344-8208-36c698aa9f5a",
							"name": "route_1",
							"route_1_field_1": "element",
							"route_field_1_array_2": [
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
							"route_1_field_3": "element",
							"route_1_field_4": "element",
							"route_field_1_array_5": [
								{
									"array_element_1": "value_index_1"
								}
							]
						}
					]
				}
			}`,
			dataPlaneVersion: "2.6.0",
			expectedPayload: `{
				"config_table": {
					"routes": [
						{
							"id": "29b2b210-3a3a-4344-8208-36c698aa9f5a",
							"name": "route_1",
							"route_field_1_array_2": [
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
							"route_1_field_3": "element",
							"route_field_1_array_5": [
								{
									"array_element_1": "value_index_1"
								}
							]
						}
					]
				}
			}`,
			expectedChanges: TrackedChanges{
				ChangeDetails: []ChangeDetail{
					{
						ID: "T101",
						Resources: []ResourceInfo{
							{
								Type: "route",
								ID:   "29b2b210-3a3a-4344-8208-36c698aa9f5a",
							},
						},
					},
					{
						ID: "T102",
						Resources: []ResourceInfo{
							{
								Type: "route",
								ID:   "29b2b210-3a3a-4344-8208-36c698aa9f5a",
							},
						},
					},
				},
			},
		},
		{
			name: "drop single upstream field",
			configTableUpdates: map[string][]ConfigTableUpdates{
				"< 3.0.0": {
					{
						Type: Upstream,
						RemoveFields: []string{
							"upstream_field_1",
						},
						ChangeID: "P042",
					},
				},
			},
			uncompressedPayload: `{
				"config_table": {
					"upstreams": [
						{
							"name": "upstream_1",
							"upstream_field_1": "element",
							"upstream_field_2": "element"
						}
					]
				}
			}`,
			dataPlaneVersion: "2.7.0",
			expectedPayload: `{
				"config_table": {
					"upstreams": [
						{
							"name": "upstream_1",
							"upstream_field_2": "element"
						}
					]
				}
			}`,
		},
		{
			name: "drop multiple upstream fields",
			configTableUpdates: map[string][]ConfigTableUpdates{
				"< 3.0.0": {
					{
						Type: Upstream,
						RemoveFields: []string{
							"upstream_field_1",
							"upstream_field_2",
						},
						ChangeID: "P042",
					},
				},
			},
			uncompressedPayload: `{
				"config_table": {
					"upstreams": [
						{
							"name": "upstream_1",
							"upstream_field_1": "element",
							"upstream_field_2": "element",
							"upstream_field_3": "element"
						}
					]
				}
			}`,
			dataPlaneVersion: "2.7.0",
			expectedPayload: `{
				"config_table": {
					"upstreams": [
						{
							"name": "upstream_1",
							"upstream_field_3": "element"
						}
					]
				}
			}`,
		},
		{
			name: "drop multiple upstream fields from multiple upstreams",
			configTableUpdates: map[string][]ConfigTableUpdates{
				"< 3.0.0": {
					{
						Type: Upstream,
						RemoveFields: []string{
							"upstream_field_1",
							"upstream_field_2",
							"upstream_field_4",
						},
						ChangeID: "P042",
					},
				},
			},
			uncompressedPayload: `{
				"config_table": {
					"upstreams": [
						{
							"name": "upstream_1",
							"upstream_field_1": "element",
							"upstream_field_2": "element",
							"upstream_field_3": "element"
						},
						{
							"name": "upstream_2",
							"upstream_field_1": "element",
							"upstream_field_3": "element",
							"upstream_field_4": "element"
						}
					]
				}
			}`,
			dataPlaneVersion: "2.7.0",
			expectedPayload: `{
				"config_table": {
					"upstreams": [
						{
							"name": "upstream_1",
							"upstream_field_3": "element"
						},
						{
							"name": "upstream_2",
							"upstream_field_3": "element"
						}
					]
				}
			}`,
		},
		{
			name: "drop upstreams and plugins' fields",
			configTableUpdates: map[string][]ConfigTableUpdates{
				"< 3.0.0": {
					{
						Type: Upstream,
						RemoveFields: []string{
							"upstream_field_1",
							"upstream_field_2",
							"upstream_field_4",
						},
						ChangeID: "P042",
					},
					{
						Name:     "plugin_1",
						Type:     Plugin,
						Remove:   true,
						ChangeID: "P043",
					},
				},
			},
			uncompressedPayload: `{
				"config_table": {
					"plugins": [
						{
							"name": "plugin_2",
							"config": {
								"plugin_2_field_1": "element"
							}
						},
						{
							"name": "plugin_1",
							"config": {
								"plugin_1_field_1": "element"
							}
						},
						{
							"name": "plugin_3",
							"config": {
								"plugin_3_field_1": "element"
							}
						},
						{
							"name": "plugin_1",
							"config": {
								"plugin_1_field_1": "element"
							}
						}
					],
					"upstreams": [
						{
							"name": "upstream_1",
							"upstream_field_1": "element",
							"upstream_field_2": "element",
							"upstream_field_3": "element"
						},
						{
							"name": "upstream_2",
							"upstream_field_1": "element",
							"upstream_field_3": "element",
							"upstream_field_4": "element"
						}
					]
				}
			}`,
			dataPlaneVersion: "2.7.0",
			expectedPayload: `{
				"config_table": {
					"plugins": [
						{
							"name": "plugin_2",
							"config": {
								"plugin_2_field_1": "element"
							}
						},
						{
							"name": "plugin_3",
							"config": {
								"plugin_3_field_1": "element"
							}
						}
					],
					"upstreams": [
						{
							"name": "upstream_1",
							"upstream_field_3": "element"
						},
						{
							"name": "upstream_2",
							"upstream_field_3": "element"
						}
					]
				}
			}`,
		},
		{
			name: "ensure unsupported upstreams fields values are changed to 'none'",
			configTableUpdates: map[string][]ConfigTableUpdates{
				"< 3.0.0": {
					{
						Type: Upstream,
						FieldUpdates: []ConfigTableFieldCondition{
							{
								Field:     "hash_on",
								Condition: "hash_on=path",
								Updates: []ConfigTableFieldUpdate{
									{
										Field: "hash_on",
										Value: "none",
									},
								},
							},
							{
								Field:     "hash_on",
								Condition: "hash_on=query_arg",
								Updates: []ConfigTableFieldUpdate{
									{
										Field: "hash_on",
										Value: "none",
									},
								},
							},
							{
								Field:     "hash_on",
								Condition: "hash_on=uri_capture",
								Updates: []ConfigTableFieldUpdate{
									{
										Field: "hash_on",
										Value: "none",
									},
								},
							},
							{
								Field:     "hash_fallback",
								Condition: "hash_fallback=path",
								Updates: []ConfigTableFieldUpdate{
									{
										Field: "hash_fallback",
										Value: "none",
									},
								},
							},
							{
								Field:     "hash_fallback",
								Condition: "hash_fallback=query_arg",
								Updates: []ConfigTableFieldUpdate{
									{
										Field: "hash_fallback",
										Value: "none",
									},
								},
							},
							{
								Field:     "hash_fallback",
								Condition: "hash_fallback=uri_capture",
								Updates: []ConfigTableFieldUpdate{
									{
										Field: "hash_fallback",
										Value: "none",
									},
								},
							},
						},
						ChangeID: "P042",
					},
				},
			},
			uncompressedPayload: `{
				"config_table": {
					"upstreams": [
						{
							"name": "upstream_1",
							"hash_on": "path",
							"hash_fallback": "path"
						},
						{
							"name": "upstream_2",
							"hash_on": "query_arg",
							"hash_fallback": "query_arg"
						},
						{
							"name": "upstream_3",
							"hash_on": "uri_capture",
							"hash_fallback": "uri_capture"
						}
					]
				}
			}`,
			dataPlaneVersion: "2.7.0",
			expectedPayload: `{
				"config_table": {
					"upstreams": [
						{
							"name": "upstream_1",
							"hash_on": "none",
							"hash_fallback": "none"
						},
						{
							"name": "upstream_2",
							"hash_on": "none",
							"hash_fallback": "none"
						},
						{
							"name": "upstream_3",
							"hash_on": "none",
							"hash_fallback": "none"
						}
					]
				}
			}`,
		},
		{
			name: "ensure unsupported upstreams fields are replaced and dropped",
			configTableUpdates: map[string][]ConfigTableUpdates{
				"< 3.0.0": {
					{
						Type: Upstream,
						FieldUpdates: []ConfigTableFieldCondition{
							{
								Field:     "upstream_unsupported_1",
								Condition: "upstream_unsupported_1",
								Updates: []ConfigTableFieldUpdate{
									{
										Field:          "upstream_supported_2",
										ValueFromField: "upstream_unsupported_1",
									},
									{
										Field: "upstream_unsupported_1",
									},
								},
							},
						},
						ChangeID: "P042",
					},
				},
			},
			uncompressedPayload: `{
				"config_table": {
					"upstreams": [
						{
							"name": "upstream_1",
							"upstream_unsupported_1": "foo"
						}
					]
				}
			}`,
			dataPlaneVersion: "2.7.0",
			expectedPayload: `{
				"config_table": {
					"upstreams": [
						{
							"name": "upstream_1",
							"upstream_supported_2": "foo"
						}
					]
				}
			}`,
		},
		{
			name: "ensure missing upstreams ValueFromField doesn't cause any change",
			configTableUpdates: map[string][]ConfigTableUpdates{
				"< 3.0.0": {
					{
						Type: Upstream,
						FieldUpdates: []ConfigTableFieldCondition{
							{
								Field:     "upstream_unsupported_1",
								Condition: "upstream_unsupported_1",
								Updates: []ConfigTableFieldUpdate{
									{
										Field:          "upstream_supported_1",
										ValueFromField: "upstream_unsupported_2",
									},
									{
										Field: "upstream_unsupported_1",
									},
								},
							},
						},
						ChangeID: "P042",
					},
				},
			},
			uncompressedPayload: `{
				"config_table": {
					"upstreams": [
						{
							"name": "upstream_1",
							"upstream_unsupported_1": "foo"
						}
					]
				}
			}`,
			dataPlaneVersion: "2.7.0",
			expectedPayload: `{
				"config_table": {
					"upstreams": [
						{
							"name": "upstream_1",
							"upstream_unsupported_1": "foo"
						}
					]
				}
			}`,
		},
		{
			name: "field updates should overwrite existing empty values",
			configTableUpdates: map[string][]ConfigTableUpdates{
				">= 3.0.0": {
					{
						Name: "plugin_1",
						Type: Plugin,
						FieldUpdates: []ConfigTableFieldCondition{
							{
								Field:     "plugin_field_1",
								Condition: "plugin_field_1",
								Updates: []ConfigTableFieldUpdate{
									{
										Field:            "plugin_field_2",
										ValueFromField:   "plugin_field_1",
										FieldMustBeEmpty: true,
									},
									{
										Field: "plugin_field_1",
									},
								},
							},
						},
						ChangeID: "T101",
					},
					{
						Name: "plugin_1",
						Type: Plugin,
						RemoveFields: []string{
							"plugin_field_1",
						},
						ChangeID: "T102",
					},
				},
			},
			uncompressedPayload: `{
				"config_table": {
					"plugins": [
						{
							"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							"name": "plugin_1",
							"config": {
								"plugin_field_1": ["kong.log.err('Hello Koko!')"],
								"plugin_field_2": []
							}
						}
					]
				}
			}`,
			dataPlaneVersion: "3.0.0",
			expectedPayload: `{
				"config_table": {
					"plugins": [
						{
							"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							"name": "plugin_1",
							"config": {
								"plugin_field_2": ["kong.log.err('Hello Koko!')"]
							}
						}
					]
				}
			}`,
			expectedChanges: TrackedChanges{
				ChangeDetails: []ChangeDetail{
					{
						ID: "T101",
						Resources: []ResourceInfo{
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
			name: "field updates do not overwrite existing values",
			configTableUpdates: map[string][]ConfigTableUpdates{
				">= 3.0.0": {
					{
						Name: "plugin_1",
						Type: Plugin,
						FieldUpdates: []ConfigTableFieldCondition{
							{
								Field:     "plugin_field_1",
								Condition: "plugin_field_1",
								Updates: []ConfigTableFieldUpdate{
									{
										Field:            "plugin_field_2",
										ValueFromField:   "plugin_field_1",
										FieldMustBeEmpty: true,
									},
									{
										Field: "plugin_field_1",
									},
								},
							},
						},
						ChangeID: "T101",
					},
				},
			},
			uncompressedPayload: `{
				"config_table": {
					"plugins": [
						{
							"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							"name": "plugin_1",
							"config": {
								"plugin_field_1": ["kong.log.err('Should not overwrite')"],
								"plugin_field_2": ["kong.log.err('Hello Koko!')"]
							}
						}
					]
				}
			}`,
			dataPlaneVersion: "3.0.0",
			expectedPayload: `{
				"config_table": {
					"plugins": [
						{
							"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							"name": "plugin_1",
							"config": {
								"plugin_field_2": ["kong.log.err('Hello Koko!')"]
							}
						}
					]
				}
			}`,
			expectedChanges: TrackedChanges{
				ChangeDetails: []ChangeDetail{
					{
						ID: "T101",
						Resources: []ResourceInfo{
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
			name: "field updates conditionally ignore empty arrays",
			configTableUpdates: map[string][]ConfigTableUpdates{
				">= 3.0.0": {
					{
						Name: "plugin_1",
						Type: Plugin,
						FieldUpdates: []ConfigTableFieldCondition{
							{
								Field:     "plugin_field_1",
								Condition: "plugin_field_1",
								Updates: []ConfigTableFieldUpdate{
									{
										Field:            "plugin_field_2",
										ValueFromField:   "plugin_field_1",
										FieldMustBeEmpty: true,
									},
									{
										Field: "plugin_field_1",
									},
								},
							},
						},
						ChangeID: "T101",
					},
				},
			},
			uncompressedPayload: `{
				"config_table": {
					"plugins": [
						{
							"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							"name": "plugin_1",
							"config": {
								"plugin_field_1": [],
								"plugin_field_2": ["kong.log.err('Hello Koko!')"]
							}
						}
					]
				}
			}`,
			dataPlaneVersion: "3.0.0",
			expectedPayload: `{
				"config_table": {
					"plugins": [
						{
							"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							"name": "plugin_1",
							"config": {
								"plugin_field_2": ["kong.log.err('Hello Koko!')"]
							}
						}
					]
				}
			}`,
			expectedChanges: TrackedChanges{
				ChangeDetails: []ChangeDetail{
					{
						ID: "T101",
						Resources: []ResourceInfo{
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
			name: "field updates conditionally ignore empty objects",
			configTableUpdates: map[string][]ConfigTableUpdates{
				">= 3.0.0": {
					{
						Name: "plugin_1",
						Type: Plugin,
						FieldUpdates: []ConfigTableFieldCondition{
							{
								Field:     "plugin_field_1",
								Condition: "plugin_field_1",
								Updates: []ConfigTableFieldUpdate{
									{
										Field:            "plugin_field_2",
										ValueFromField:   "plugin_field_1",
										FieldMustBeEmpty: true,
									},
									{
										Field: "plugin_field_1",
									},
								},
							},
						},
						ChangeID: "T101",
					},
				},
			},
			uncompressedPayload: `{
				"config_table": {
					"plugins": [
						{
							"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							"name": "plugin_1",
							"config": {
								"plugin_field_1": {},
								"plugin_field_2": { "foo": "bar" }
							}
						}
					]
				}
			}`,
			dataPlaneVersion: "3.0.0",
			expectedPayload: `{
				"config_table": {
					"plugins": [
						{
							"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							"name": "plugin_1",
							"config": {
								"plugin_field_2": { "foo": "bar" }
							}
						}
					]
				}
			}`,
			expectedChanges: TrackedChanges{
				ChangeDetails: []ChangeDetail{
					{
						ID: "T101",
						Resources: []ResourceInfo{
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
			name: "field updates conditionally ignore empty strings",
			configTableUpdates: map[string][]ConfigTableUpdates{
				">= 3.0.0": {
					{
						Name: "plugin_1",
						Type: Plugin,
						FieldUpdates: []ConfigTableFieldCondition{
							{
								Field:     "plugin_field_1",
								Condition: "plugin_field_1",
								Updates: []ConfigTableFieldUpdate{
									{
										Field:            "plugin_field_2",
										ValueFromField:   "plugin_field_1",
										FieldMustBeEmpty: true,
									},
									{
										Field: "plugin_field_1",
									},
								},
							},
						},
						ChangeID: "T101",
					},
				},
			},
			uncompressedPayload: `{
				"config_table": {
					"plugins": [
						{
							"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							"name": "plugin_1",
							"config": {
								"plugin_field_1": "",
								"plugin_field_2": "foo"
							}
						}
					]
				}
			}`,
			dataPlaneVersion: "3.0.0",
			expectedPayload: `{
				"config_table": {
					"plugins": [
						{
							"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							"name": "plugin_1",
							"config": {
								"plugin_field_2": "foo"
							}
						}
					]
				}
			}`,
			expectedChanges: TrackedChanges{
				ChangeDetails: []ChangeDetail{
					{
						ID: "T101",
						Resources: []ResourceInfo{
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
			name: "field updates conditionally ignore nil values",
			configTableUpdates: map[string][]ConfigTableUpdates{
				">= 3.0.0": {
					{
						Name: "plugin_1",
						Type: Plugin,
						FieldUpdates: []ConfigTableFieldCondition{
							{
								Field:     "plugin_field_1",
								Condition: "plugin_field_1",
								Updates: []ConfigTableFieldUpdate{
									{
										Field:            "plugin_field_2",
										ValueFromField:   "plugin_field_1",
										FieldMustBeEmpty: true,
									},
									{
										Field: "plugin_field_1",
									},
								},
							},
						},
						ChangeID: "T101",
					},
				},
			},
			uncompressedPayload: `{
				"config_table": {
					"plugins": [
						{
							"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							"name": "plugin_1",
							"config": {
								"plugin_field_1": null,
								"plugin_field_2": "foo"
							}
						}
					]
				}
			}`,
			dataPlaneVersion: "3.0.0",
			expectedPayload: `{
				"config_table": {
					"plugins": [
						{
							"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							"name": "plugin_1",
							"config": {
								"plugin_field_2": "foo"
							}
						}
					]
				}
			}`,
			expectedChanges: TrackedChanges{
				ChangeDetails: []ChangeDetail{
					{
						ID: "T101",
						Resources: []ResourceInfo{
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
			name: "field updates do not create new fields when ignoring empty array values",
			configTableUpdates: map[string][]ConfigTableUpdates{
				">= 3.0.0": {
					{
						Name: "plugin_1",
						Type: Plugin,
						FieldUpdates: []ConfigTableFieldCondition{
							{
								Field:     "plugin_field_1",
								Condition: "plugin_field_1",
								Updates: []ConfigTableFieldUpdate{
									{
										Field:            "plugin_field_2",
										ValueFromField:   "plugin_field_1",
										FieldMustBeEmpty: true,
									},
									{
										Field: "plugin_field_1",
									},
								},
							},
						},
						ChangeID: "T101",
					},
					{
						Name: "plugin_1",
						Type: Plugin,
						RemoveFields: []string{
							"plugin_field_1",
						},
						DisableChangeTracking: func(rawJSON string) bool {
							// do not emit change if functions is set to default value (empty array)
							plugin := gjson.Parse(rawJSON)
							return len(plugin.Get("config.plugin_field_1").Array()) == 0
						},
						ChangeID: "T102",
					},
				},
			},
			uncompressedPayload: `{
				"config_table": {
					"plugins": [
						{
							"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							"name": "plugin_1",
							"config": {
								"plugin_field_1": []
							}
						}
					]
				}
			}`,
			dataPlaneVersion: "3.0.0",
			expectedPayload: `{
				"config_table": {
					"plugins": [
						{
							"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							"name": "plugin_1",
							"config": {}
						}
					]
				}
			}`,
			expectedChanges: TrackedChanges{},
		},
		{
			name: "field updates conditionally ignore empty arrays",
			configTableUpdates: map[string][]ConfigTableUpdates{
				">= 3.0.0": {
					{
						Name: "plugin_1",
						Type: CorePlugin,
						FieldUpdates: []ConfigTableFieldCondition{
							{
								Field:     "plugin_field_1",
								Condition: "plugin_field_1",
								Updates: []ConfigTableFieldUpdate{
									{
										Field:            "plugin_field_2",
										ValueFromField:   "plugin_field_1",
										FieldMustBeEmpty: true,
									},
									{
										Field: "plugin_field_1",
									},
								},
							},
						},
						ChangeID: "T101",
					},
					{
						Name: "plugin_1",
						Type: CorePlugin,
						RemoveFields: []string{
							"plugin_field_1",
						},
						ChangeID: "T102",
					},
				},
			},
			uncompressedPayload: `{
				"config_table": {
					"plugins": [
						{
							"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							"name": "plugin_1",
							"plugin_field_1": [],
							"plugin_field_2": ["kong.log.err('Hello Koko!')"]
						}
					]
				}
			}`,
			dataPlaneVersion: "3.0.0",
			expectedPayload: `{
				"config_table": {
					"plugins": [
						{
							"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							"name": "plugin_1",
							"plugin_field_2": ["kong.log.err('Hello Koko!')"]
						}
					]
				}
			}`,
			expectedChanges: TrackedChanges{
				ChangeDetails: []ChangeDetail{
					{
						ID: "T102",
						Resources: []ResourceInfo{
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
			name: "field updates conditionally ignore empty objects",
			configTableUpdates: map[string][]ConfigTableUpdates{
				">= 3.0.0": {
					{
						Name: "plugin_1",
						Type: CorePlugin,
						FieldUpdates: []ConfigTableFieldCondition{
							{
								Field:     "plugin_field_1",
								Condition: "plugin_field_1",
								Updates: []ConfigTableFieldUpdate{
									{
										Field:            "plugin_field_2",
										ValueFromField:   "plugin_field_1",
										FieldMustBeEmpty: true,
									},
									{
										Field: "plugin_field_1",
									},
								},
							},
						},
						ChangeID: "T101",
					},
					{
						Name: "plugin_1",
						Type: CorePlugin,
						RemoveFields: []string{
							"plugin_field_1",
						},
						ChangeID: "T102",
					},
				},
			},
			uncompressedPayload: `{
				"config_table": {
					"plugins": [
						{
							"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							"name": "plugin_1",
							"plugin_field_1": {},
							"plugin_field_2": { "foo": "bar" }
						}
					]
				}
			}`,
			dataPlaneVersion: "3.0.0",
			expectedPayload: `{
				"config_table": {
					"plugins": [
						{
							"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							"name": "plugin_1",
							"plugin_field_2": { "foo": "bar" }
						}
					]
				}
			}`,
			expectedChanges: TrackedChanges{
				ChangeDetails: []ChangeDetail{
					{
						ID: "T102",
						Resources: []ResourceInfo{
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
			name: "field updates conditionally ignore empty strings",
			configTableUpdates: map[string][]ConfigTableUpdates{
				">= 3.0.0": {
					{
						Name: "plugin_1",
						Type: CorePlugin,
						FieldUpdates: []ConfigTableFieldCondition{
							{
								Field:     "plugin_field_1",
								Condition: "plugin_field_1",
								Updates: []ConfigTableFieldUpdate{
									{
										Field:            "plugin_field_2",
										ValueFromField:   "plugin_field_1",
										FieldMustBeEmpty: true,
									},
									{
										Field: "plugin_field_1",
									},
								},
							},
						},
						ChangeID: "T101",
					},
					{
						Name: "plugin_1",
						Type: CorePlugin,
						RemoveFields: []string{
							"plugin_field_1",
						},
						ChangeID: "T102",
					},
				},
			},
			uncompressedPayload: `{
				"config_table": {
					"plugins": [
						{
							"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							"name": "plugin_1",
							"plugin_field_1": "",
							"plugin_field_2": "foo"
						}
					]
				}
			}`,
			dataPlaneVersion: "3.0.0",
			expectedPayload: `{
				"config_table": {
					"plugins": [
						{
							"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							"name": "plugin_1",
							"plugin_field_2": "foo"
						}
					]
				}
			}`,
			expectedChanges: TrackedChanges{
				ChangeDetails: []ChangeDetail{
					{
						ID: "T102",
						Resources: []ResourceInfo{
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
			name: "field updates conditionally ignore nil values",
			configTableUpdates: map[string][]ConfigTableUpdates{
				">= 3.0.0": {
					{
						Name: "plugin_1",
						Type: CorePlugin,
						FieldUpdates: []ConfigTableFieldCondition{
							{
								Field:     "plugin_field_1",
								Condition: "plugin_field_1",
								Updates: []ConfigTableFieldUpdate{
									{
										Field:            "plugin_field_2",
										ValueFromField:   "plugin_field_1",
										FieldMustBeEmpty: true,
									},
									{
										Field: "plugin_field_1",
									},
								},
							},
						},
						ChangeID: "T101",
					},
					{
						Name: "plugin_1",
						Type: CorePlugin,
						RemoveFields: []string{
							"plugin_field_1",
						},
						ChangeID: "T102",
					},
				},
			},
			uncompressedPayload: `{
				"config_table": {
					"plugins": [
						{
							"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							"name": "plugin_1",
							"plugin_field_1": null,
							"plugin_field_2": "foo"
						}
					]
				}
			}`,
			dataPlaneVersion: "3.0.0",
			expectedPayload: `{
				"config_table": {
					"plugins": [
						{
							"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							"name": "plugin_1",
							"plugin_field_2": "foo"
						}
					]
				}
			}`,
			expectedChanges: TrackedChanges{
				ChangeDetails: []ChangeDetail{
					{
						ID: "T102",
						Resources: []ResourceInfo{
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
			name: "field updates do not create new fields when ignoring empty array values",
			configTableUpdates: map[string][]ConfigTableUpdates{
				">= 3.0.0": {
					{
						Name: "plugin_1",
						Type: CorePlugin,
						FieldUpdates: []ConfigTableFieldCondition{
							{
								Field:     "plugin_field_1",
								Condition: "plugin_field_1",
								Updates: []ConfigTableFieldUpdate{
									{
										Field:            "plugin_field_2",
										ValueFromField:   "plugin_field_1",
										FieldMustBeEmpty: true,
									},
									{
										Field: "plugin_field_1",
									},
								},
							},
						},
						ChangeID: "T101",
					},
					{
						Name: "plugin_1",
						Type: CorePlugin,
						RemoveFields: []string{
							"plugin_field_1",
						},
						DisableChangeTracking: func(rawJSON string) bool {
							// do not emit change if functions is set to default value (empty array)
							plugin := gjson.Parse(rawJSON)
							return len(plugin.Get("config.plugin_field_1").Array()) == 0
						},
						ChangeID: "T102",
					},
				},
			},
			uncompressedPayload: `{
				"config_table": {
					"plugins": [
						{
							"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							"name": "plugin_1",
							"plugin_field_1": []
						}
					]
				}
			}`,
			dataPlaneVersion: "3.0.0",
			expectedPayload: `{
				"config_table": {
					"plugins": [
						{
							"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							"name": "plugin_1"
						}
					]
				}
			}`,
			expectedChanges: TrackedChanges{},
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("entity update for %s", test.name), func(t *testing.T) {
			wsvc, err := NewVersionCompatibilityProcessor(VersionCompatibilityOpts{
				Logger:        log.Logger,
				KongCPVersion: "2.8.0",
			})
			require.NoError(t, err)
			err = wsvc.AddConfigTableUpdates(test.configTableUpdates)
			require.NoError(t, err)

			tracker := NewChangeTracker()
			dataPlaneVersion := versioning.MustNewVersion(test.dataPlaneVersion)
			processedPayload, err := wsvc.processConfigTableUpdates(test.uncompressedPayload,
				dataPlaneVersion, tracker)
			require.Nil(t, err)
			require.JSONEq(t, test.expectedPayload, processedPayload)

			require.Equal(t, test.expectedChanges, tracker.Get())
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
		processedPayloadCompressed, tracker, err := wsvc.ProcessConfigTableUpdates("2.8.0", expectedPayload)
		require.Nil(t, err)
		require.NotNil(t, tracker)

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
				ExtraProcessor: func(uncompressedPayload string, dataPlaneVersion versioning.Version,
					tracker *ChangeTracker, logger *zap.Logger,
				) (string, error) {
					if test.wantsErr {
						return "", fmt.Errorf("extra processing error")
					}
					if test.wantsInvalidJSON {
						return "invalid JSON", nil
					}
					if dataPlaneVersion.IsKongGatewayEnterprise() {
						return `{"extra_processing": "enterprise-edition"}`, nil
					}
					return `{"extra_processing": "oss"}`, nil
				},
			})
			require.Nil(t, err)

			dataPlaneVersion := versioning.MustNewVersion("2.8.0")
			if test.isEnterprise {
				dataPlaneVersion = versioning.MustNewVersion("2.8.0.0")
			}
			processedPayload, err := wsvc.performExtraProcessing("{}", dataPlaneVersion, nil)
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
			ExtraProcessor: func(uncompressedPayload string, dataPlaneVersion versioning.Version,
				tracker *ChangeTracker, logger *zap.Logger,
			) (string, error) {
				return sjson.Set(uncompressedPayload, "config_table.extra_processing", "processed")
			},
		})
		require.Nil(t, err)

		payload := `{"config_table": {"extra_processing": "unprocessed"}, "type": "reconfigure"}`
		compressedPayload, err := CompressPayload([]byte(payload))
		require.Nil(t, err)
		processedPayloadCompressed, tracker, err := wsvc.ProcessConfigTableUpdates("2.8.0", compressedPayload)
		require.Nil(t, err)
		require.NotNil(t, tracker)
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

func TestVersionCompatibility_ValueIsEmpty(t *testing.T) {
	tests := []struct {
		name          string
		rawJSON       string
		expectedValue bool
	}{
		{
			name: "object is empty",
			rawJSON: `{
				"value": {}
			}`,
			expectedValue: true,
		},
		{
			name: "object is not empty",
			rawJSON: `{
				"value": { "foo": "bar" }
			}`,
			expectedValue: false,
		},
		{
			name: "string is empty",
			rawJSON: `{
				"value": ""
			}`,
			expectedValue: true,
		},
		{
			name: "string is not empty",
			rawJSON: `{
				"value": "foo"
			}`,
			expectedValue: false,
		},
		{
			name: "array is empty",
			rawJSON: `{
				"value": []
			}`,
			expectedValue: true,
		},
		{
			name: "array is not empty",
			rawJSON: `{
				"value": ["foo"]
			}`,
			expectedValue: false,
		},
		{
			name: "value is null",
			rawJSON: `{
				"value": null
			}`,
			expectedValue: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.expectedValue, valueIsEmpty(gjson.Get(test.rawJSON, "value")))
		})
	}
}
