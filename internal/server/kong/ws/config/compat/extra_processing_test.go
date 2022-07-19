package compat

import (
	"testing"

	"github.com/kong/koko/internal/log"
	"github.com/stretchr/testify/require"
)

func TestExtraProcessing_RemovePlugins(t *testing.T) {
	tests := []struct {
		name                string
		pluginName          string
		uncompressedPayload string
		expectedPayload     string
	}{
		{
			name:       "ensure single configured plugin is removed",
			pluginName: "plugin_1",
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
			expectedPayload: `{
				"config_table": {
					"plugins": []
				}
			}`,
		},
		{
			name:       "ensure multiple configured plugin are removed",
			pluginName: "plugin_1",
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
								"plugin_1_field_1": "element"
							}
						},
						{
							"name": "plugin_1",
							"config": {
								"plugin_1_field_1": "element"
							}
						}
					]
				}
			}`,
			expectedPayload: `{
				"config_table": {
					"plugins": []
				}
			}`,
		},
		{
			name:       "ensure single plugin is removed from multiple configured plugins",
			pluginName: "plugin_1",
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
						}
					]
				}
			}`,
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
					]
				}
			}`,
		},
		{
			name:       "ensure multiple plugins are removed from multiple configured plugins",
			pluginName: "plugin_1",
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
					]
				}
			}`,
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
					]
				}
			}`,
		},
		{
			name:       "ensure no plugins are removed from multiple configured plugins",
			pluginName: "plugin_0",
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
						},
						{
							"name": "plugin_4",
							"config": {
								"plugin_4_field_1": "element"
							}
						}
					]
				}
			}`,
			expectedPayload: `{
				"config_table": {
					"plugins": [
						{
							"name": "plugin_1",
							"config": {
								"plugin_1_field_1": "element"
							}
						},
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
						},
						{
							"name": "plugin_4",
							"config": {
								"plugin_4_field_1": "element"
							}
						}
					]
				}
			}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			processedPayload := removePlugin(test.uncompressedPayload, test.pluginName, 2006000000, log.Logger)
			require.JSONEq(t, test.expectedPayload, processedPayload)
		})
	}
}
