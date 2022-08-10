package config

import (
	"fmt"
	"strings"
	"testing"

	"github.com/blang/semver/v4"
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
		expectedVersion string
	}{
		{
			versionStr:      "0.33.3",
			expectedVersion: "0.33.3",
		},
		{
			versionStr:      "0.33.3-3-enterprise-edition",
			expectedVersion: "0.33.3.3",
		},
		{
			versionStr:      "0.33.3-3-enterprise",
			expectedVersion: "0.33.3.3",
		},
		{
			// go-kong won't parse build without suffix containing enterprise
			versionStr:      "0.33.3-3-build-will-not-be-parsed",
			expectedVersion: "0.33.3-3",
		},
		{
			versionStr:      "2.3.3-2",
			expectedVersion: "2.3.3-2",
		},
		{
			versionStr:      "2.3.2",
			expectedVersion: "2.3.2",
		},
		{
			versionStr:      "2.3.2-rc1",
			expectedVersion: "2.3.2",
		},
		{
			versionStr:      "2.3.3-alpha",
			expectedVersion: "2.3.3",
		},
		{
			versionStr:      "2.3.4-beta1",
			expectedVersion: "2.3.4",
		},
		{
			versionStr:      "2.3.3.2-enterprise-edition",
			expectedVersion: "2.3.3.2",
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
		version, err := ParseSemanticVersion(test.versionStr)
		if test.wantsErr {
			require.NotNil(t, err)
			require.EqualError(t, err, test.expectedErr)
			require.EqualValues(t, "", version)
		} else {
			require.Nil(t, err)
			require.Equal(t, test.expectedVersion, version)
		}
	}
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
			version, err := semver.Parse(test.dataPlaneVersion)
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
							"id": "5441b100-f441-4d4b-bcc2-3bb153e2bd40",
							"name": "plugin_1",
							"config": {
								"plugin_1_field_2": "element",
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
								ID:   "ab3b5a6d-923e-4e71-83b4-77e4b68d3e55",
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
			name: "ensure multiple plugins are removed from multiple configured plugins",
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
							"name": "plugin_2",
							"config": {
								"plugin_2_field_1": "element"
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
							"id": "5441b100-f441-4d4b-bcc2-3bb153e2bd41",
							"name": "plugin_2",
							"config": {
								"plugin_2_field_1": "element"
							}
						},
						{
							"id": "5441b100-f441-4d4b-bcc2-3bb153e2bd43",
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
			name: "drop services and plugins' fields",
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
							}
						},
						{
							"id": "5441b100-f441-4d4b-bcc2-3bb153e2bd40",
							"name": "plugin_1",
							"config": {
								"plugin_1_field_1": "element"
							}
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
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("plugin update for %s", test.name), func(t *testing.T) {
			wsvc, err := NewVersionCompatibilityProcessor(VersionCompatibilityOpts{
				Logger:        log.Logger,
				KongCPVersion: "2.8.0",
			})
			require.NoError(t, err)
			err = wsvc.AddConfigTableUpdates(test.configTableUpdates)
			require.NoError(t, err)

			tracker := NewChangeTracker()
			processedPayload, err := wsvc.processConfigTableUpdates(test.uncompressedPayload,
				test.dataPlaneVersion, tracker)
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
				ExtraProcessor: func(uncompressedPayload string, dataPlaneVersion string, isEnterprise bool,
					tracker *ChangeTracker, logger *zap.Logger,
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

			processedPayload, err := wsvc.performExtraProcessing("{}",
				"2.8.0", test.isEnterprise, nil)
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
			ExtraProcessor: func(uncompressedPayload string, dataPlaneVersion string, isEnterprise bool,
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
