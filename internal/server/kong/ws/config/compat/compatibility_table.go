package compat

import (
	"fmt"
	"strings"

	"github.com/kong/koko/internal/server/kong/ws/config"
)

func standardUpgradeMessage(version string) string {
	if version == "" {
		panic("no version provided")
	}
	return fmt.Sprintf("Please upgrade Kong Gateway to version '%s' "+
		"or above.", version)
}

func standardPluginFieldsMessage(
	pluginName string, fields []string, versionWithFeatureSupport string, isNewer bool,
) string {
	quotedFields := "'" + strings.Join(fields, "', '") + "'"
	olderOrNewer := "<"
	if isNewer {
		olderOrNewer = ">="
	}
	return fmt.Sprintf("For the '%s' plugin, "+
		"one or more of the following 'config' fields are set: %s "+
		"but Kong Gateway versions %s %s do not support these fields. "+
		"Plugin features that rely on these fields are not working as intended.",
		pluginName,
		quotedFields,
		olderOrNewer,
		versionWithFeatureSupport,
	)
}

func standardPluginNotAvailableMessage(pluginName string, versionWithFeatureSupport string) string {
	return fmt.Sprintf("Plugin '%s' is not available in Kong gateway versions "+
		"< %s. '",
		pluginName,
		versionWithFeatureSupport,
	)
}

const (
	versionsPre260      = "< 2.6.0"
	versionsPre270      = "< 2.7.0"
	versionsPre280      = "< 2.8.0"
	versionsPre300      = "< 3.0.0"
	versions300AndAbove = ">= 3.0.0"
)

var (
	acme25xFields = []string{
		"preferred_chain",
		"storage_config.vault.auth_method",
		"storage_config.vault.auth_path",
		"storage_config.vault.auth_role",
		"storage_config.vault.jwt_path",
	}
	zipkin30Fields = []string{
		"http_span_name",
		"connect_timeout",
		"send_timeout",
		"read_timeout",
	}
	prometheus30Fields = []string{
		"status_code_metrics",
		"latency_metrics",
		"bandwidth_metrics",
		"upstream_health_metrics",
	}

	changes = []config.Change{
		{
			Metadata: config.ChangeMetadata{
				ID:          config.ChangeID("P101"),
				Severity:    config.ChangeSeverityError,
				Description: standardPluginFieldsMessage("acme", acme25xFields, "2.6", false),
				Resolution:  standardUpgradeMessage("2.6"),
			},
			SemverRange: versionsPre260,
			Update: config.ConfigTableUpdates{
				Name:         "acme",
				Type:         config.Plugin,
				RemoveFields: acme25xFields,
			},
		},
		{
			Metadata: config.ChangeMetadata{
				ID:          config.ChangeID("P102"),
				Severity:    config.ChangeSeverityError,
				Description: standardPluginFieldsMessage("aws-lambda", []string{"base64_encode_body"}, "2.6", false),
				Resolution:  standardUpgradeMessage("2.6"),
			},
			SemverRange: versionsPre260,
			Update: config.ConfigTableUpdates{
				Name: "aws-lambda",
				Type: config.Plugin,
				RemoveFields: []string{
					"base64_encode_body",
				},
			},
		},
		{
			Metadata: config.ChangeMetadata{
				ID:       config.ChangeID("P103"),
				Severity: config.ChangeSeverityError,
				Description: standardPluginFieldsMessage("grpc-web",
					[]string{"allow_origin_header"}, "2.6", false),
				Resolution: standardUpgradeMessage("2.6"),
			},
			SemverRange: versionsPre260,
			Update: config.ConfigTableUpdates{
				Name: "grpc-web",
				Type: config.Plugin,
				RemoveFields: []string{
					"allow_origin_header",
				},
			},
		},
		{
			Metadata: config.ChangeMetadata{
				ID:       config.ChangeID("P104"),
				Severity: config.ChangeSeverityError,
				Description: standardPluginFieldsMessage("request-termination",
					[]string{"echo", "trigger"}, "2.6", false),
				Resolution: standardUpgradeMessage("2.6"),
			},
			SemverRange: versionsPre260,
			Update: config.ConfigTableUpdates{
				Name: "request-termination",
				Type: config.Plugin,
				RemoveFields: []string{
					"echo",
					"trigger",
				},
			},
		},
		{
			Metadata: config.ChangeMetadata{
				ID:       config.ChangeID("P105"),
				Severity: config.ChangeSeverityError,
				Description: standardPluginFieldsMessage("datadog",
					[]string{
						"service_name_tag",
						"status_tag",
						"consumer_tag",
					}, "2.7", false),
				Resolution: standardUpgradeMessage("2.7"),
			},
			SemverRange: versionsPre270,
			Update: config.ConfigTableUpdates{
				Name: "datadog",
				Type: config.Plugin,
				RemoveFields: []string{
					"service_name_tag",
					"status_tag",
					"consumer_tag",
				},
			},
		},
		{
			Metadata: config.ChangeMetadata{
				ID:       config.ChangeID("P106"),
				Severity: config.ChangeSeverityError,
				Description: "For the 'datadog' plugin, " +
					"distribution metric type is not support in Kong gateway " +
					"versions < 2.7. " +
					"Distribution metrics will not be emitted by the gateway.",
				Resolution: standardUpgradeMessage("2.7"),
			},
			SemverRange: versionsPre270,
			Update: config.ConfigTableUpdates{
				Name: "datadog",
				Type: config.Plugin,
				RemoveElementsFromArray: []config.ConfigTableFieldCondition{
					{
						Field:     "metrics",
						Condition: "stat_type=distribution",
					},
				},
			},
		},
		{
			Metadata: config.ChangeMetadata{
				ID:       config.ChangeID("P107"),
				Severity: config.ChangeSeverityError,
				Description: standardPluginFieldsMessage("ip-restriction",
					[]string{"status", "message"}, "2.7", false),
				Resolution: standardUpgradeMessage("2.7"),
			},
			SemverRange: versionsPre270,
			Update: config.ConfigTableUpdates{
				Name: "ip-restriction",
				Type: config.Plugin,
				RemoveFields: []string{
					"status",
					"message",
				},
			},
		},
		{
			Metadata: config.ChangeMetadata{
				ID:       config.ChangeID("P108"),
				Severity: config.ChangeSeverityError,
				Description: standardPluginFieldsMessage("rate-limiting",
					[]string{"redis_ssl", "redis_ssl_verify", "redis_server_name"}, "2.7", false),
				Resolution: standardUpgradeMessage("2.7"),
			},
			SemverRange: versionsPre270,
			Update: config.ConfigTableUpdates{
				Name: "rate-limiting",
				Type: config.Plugin,
				RemoveFields: []string{
					"redis_ssl",
					"redis_ssl_verify",
					"redis_server_name",
				},
			},
		},
		{
			Metadata: config.ChangeMetadata{
				ID:       config.ChangeID("P109"),
				Severity: config.ChangeSeverityError,
				Description: "For the 'zipkin' plugin, " +
					"'config.header_type' field has been set to 'ignore'. " +
					"This is not supported in Kong Gateway versions < 2.7. " +
					"The plugin configuration has been changed to 'config." +
					"header_type=preserve' in the data-plane.",
				Resolution: standardUpgradeMessage("2.7"),
			},
			SemverRange: versionsPre270,
			Update: config.ConfigTableUpdates{
				Name: "zipkin",
				Type: config.Plugin,
				FieldUpdates: []config.ConfigTableFieldCondition{
					{
						Field:     "header_type",
						Condition: "header_type=ignore",
						Updates: []config.ConfigTableFieldUpdate{
							{
								Field: "header_type",
								Value: "preserve",
							},
						},
					},
				},
			},
		},
		{
			Metadata: config.ChangeMetadata{
				ID:       config.ChangeID("P110"),
				Severity: config.ChangeSeverityError,
				Description: standardPluginFieldsMessage("zipkin",
					[]string{"local_service_name"}, "2.7", false),
				Resolution: standardUpgradeMessage("2.7"),
			},
			SemverRange: versionsPre270,
			Update: config.ConfigTableUpdates{
				Name: "zipkin",
				Type: config.Plugin,
				RemoveFields: []string{
					"local_service_name",
				},
			},
		},
		{
			Metadata: config.ChangeMetadata{
				ID:       config.ChangeID("P111"),
				Severity: config.ChangeSeverityError,
				Description: standardPluginFieldsMessage("acme",
					[]string{"rsa_key_size"}, "2.8", false),
				Resolution: standardUpgradeMessage("2.8"),
			},
			SemverRange: versionsPre280,
			Update: config.ConfigTableUpdates{
				Name: "acme",
				Type: config.Plugin,
				RemoveFields: []string{
					"rsa_key_size",
				},
			},
		},
		{
			Metadata: config.ChangeMetadata{
				ID:       config.ChangeID("P112"),
				Severity: config.ChangeSeverityError,
				Description: standardPluginFieldsMessage("rate-limiting",
					[]string{"redis_username"}, "2.8", false),
				Resolution: standardUpgradeMessage("2.8"),
			},
			SemverRange: versionsPre280,
			Update: config.ConfigTableUpdates{
				Name: "rate-limiting",
				Type: config.Plugin,
				RemoveFields: []string{
					"redis_username",
				},
			},
		},
		{
			Metadata: config.ChangeMetadata{
				ID:       config.ChangeID("P113"),
				Severity: config.ChangeSeverityError,
				Description: standardPluginFieldsMessage("response-ratelimiting",
					[]string{"redis_username"}, "2.8", false),
				Resolution: standardUpgradeMessage("2.8"),
			},
			SemverRange: versionsPre280,
			Update: config.ConfigTableUpdates{
				Name: "response-ratelimiting",
				Type: config.Plugin,
				RemoveFields: []string{
					"redis_username",
				},
			},
		},
		{
			Metadata: config.ChangeMetadata{
				ID:       config.ChangeID("P114"),
				Severity: config.ChangeSeverityError,
				Description: standardPluginFieldsMessage("response-ratelimiting",
					[]string{"redis_username"}, "2.8", false),
				Resolution: standardUpgradeMessage("3.0"),
			},
			SemverRange: versionsPre280,
			Update: config.ConfigTableUpdates{
				Name: "response-ratelimiting",
				Type: config.Plugin,
				RemoveFields: []string{
					"redis_username",
				},
			},
		},
		{
			Metadata: config.ChangeMetadata{
				ID:          config.ChangeID("P115"),
				Severity:    config.ChangeSeverityError,
				Description: standardPluginNotAvailableMessage("opentelemetry", "3.0"),
				Resolution:  standardUpgradeMessage("3.0"),
			},
			SemverRange: versionsPre300,
			Update: config.ConfigTableUpdates{
				Name:   "opentelemetry",
				Type:   config.Plugin,
				Remove: true,
			},
		},
		{
			Metadata: config.ChangeMetadata{
				ID:          config.ChangeID("P116"),
				Severity:    config.ChangeSeverityError,
				Description: standardPluginFieldsMessage("zipkin", zipkin30Fields, "3.0", false),
				Resolution:  standardUpgradeMessage("3.0"),
			},
			SemverRange: versionsPre300,
			Update: config.ConfigTableUpdates{
				Name:         "zipkin",
				Type:         config.Plugin,
				RemoveFields: zipkin30Fields,
			},
		},
		{
			Metadata: config.ChangeMetadata{
				ID:       config.ChangeID("P117"),
				Severity: config.ChangeSeverityError,
				Description: standardPluginFieldsMessage(
					"prometheus", prometheus30Fields, "3.0", false,
				),
				Resolution: standardUpgradeMessage("3.0"),
			},
			SemverRange: versionsPre300,
			Update: config.ConfigTableUpdates{
				Name:         "prometheus",
				Type:         config.Plugin,
				RemoveFields: prometheus30Fields,
			},
		},
		{
			Metadata: config.ChangeMetadata{
				ID:       config.ChangeID("P118"),
				Severity: config.ChangeSeverityError,
				Description: standardPluginFieldsMessage("acme",
					[]string{"allow_any_domain"},
					"3.0", false),
				Resolution: standardUpgradeMessage("3.0"),
			},
			SemverRange: versionsPre300,
			Update: config.ConfigTableUpdates{
				Name: "acme",
				Type: config.Plugin,
				RemoveFields: []string{
					"allow_any_domain",
				},
			},
		},
		{
			Metadata: config.ChangeMetadata{
				ID:       config.ChangeID("P119"),
				Severity: config.ChangeSeverityError,
				Description: "For the 'Service' entity, " +
					"'enabled' field has been set to 'true' but Kong " +
					"Gateway versions < 2.7 do not support this feature. " +
					"The Service has been left enabled in the Kong Gateway, " +
					"and the traffic for the Service is being routed " +
					"by Kong Gateway. " +
					"This is a critical error and may result in unwanted " +
					"traffic being routed to the upstream Services via Kong " +
					"Gateway.",
				Resolution: standardUpgradeMessage("2.7"),
			},
			SemverRange: versionsPre270,
			Update: config.ConfigTableUpdates{
				Name: config.Service.String(),
				Type: config.Service,
				RemoveFields: []string{
					"enabled",
				},
			},
		},
		{
			Metadata: config.ChangeMetadata{
				ID:       config.ChangeID("P120"),
				Severity: config.ChangeSeverityError,
				Description: standardPluginFieldsMessage("aws-lambda",
					[]string{"proxy_scheme"}, "3.0", true),
				Resolution: "Please use 'config.proxy_url' instead of " +
					"'config.proxy_scheme' field.",
			},
			SemverRange: versions300AndAbove,
			Update: config.ConfigTableUpdates{
				Name: "aws-lambda",
				Type: config.Plugin,
				RemoveFields: []string{
					"proxy_scheme",
				},
			},
		},
		{
			Metadata: config.ChangeMetadata{
				ID:       config.ChangeID("P122"),
				Severity: config.ChangeSeverityError,
				Description: "For the 'pre-function' plugin, " +
					"'config.functions' field has been used. " +
					"This is not supported in Kong Gateway versions >= 3.0. " +
					"The plugin configuration has been updated to rename " +
					"'config.functions' to 'config.access' in the data-plane.",
				Resolution: "Please update the configuration to use " +
					"'config.access' field instead of 'config.functions'.",
			},
			SemverRange: versions300AndAbove,
			Update: config.ConfigTableUpdates{
				Name: "pre-function",
				Type: config.Plugin,
				FieldUpdates: []config.ConfigTableFieldCondition{
					{
						Field:     "functions",
						Condition: "functions",
						Updates: []config.ConfigTableFieldUpdate{
							{
								Field:          "access",
								ValueFromField: "functions",
							},
							{
								Field: "functions",
							},
						},
					},
				},
			},
		},
		{
			Metadata: config.ChangeMetadata{
				ID:       config.ChangeID("P123"),
				Severity: config.ChangeSeverityError,
				Description: "For the 'post-function' plugin, " +
					"'config.functions' field has been used. " +
					"This is not supported in Kong Gateway versions >= 3.0. " +
					"The plugin configuration has been updated to rename " +
					"'config.functions' to 'config.access' in the data-plane.",
				Resolution: "Please update the configuration to use " +
					"'config.access' field instead of 'config.functions'.",
			},
			SemverRange: versions300AndAbove,
			Update: config.ConfigTableUpdates{
				Name: "post-function",
				Type: config.Plugin,
				FieldUpdates: []config.ConfigTableFieldCondition{
					{
						Field:     "functions",
						Condition: "functions",
						Updates: []config.ConfigTableFieldUpdate{
							{
								Field:          "access",
								ValueFromField: "functions",
							},
							{
								Field: "functions",
							},
						},
					},
				},
			},
		},
		{
			Metadata: config.ChangeMetadata{
				ID:       config.ChangeID("P124"),
				Severity: config.ChangeSeverityWarning,
				Description: "For the 'pre-function' plugin, " +
					"'config.functions' field has been used. " +
					"This field is deprecated and it is no longer supported " +
					"in Kong Gateway versions >= 3.0.",
				Resolution: "Please update the plugin configuration to use " +
					"'config.access' field in place of 'config.functions' field",
			},
			SemverRange: versionsPre300,
			Update: config.ConfigTableUpdates{
				Name: "pre-function",
				Type: config.Plugin,
				FieldUpdates: []config.ConfigTableFieldCondition{
					{
						Field:     "functions",
						Condition: "functions",
					},
				},
			},
		},
		{
			Metadata: config.ChangeMetadata{
				ID:       config.ChangeID("P125"),
				Severity: config.ChangeSeverityWarning,
				Description: "For the 'post-function' plugin, " +
					"'config.functions' field has been used. " +
					"This field is deprecated and it is no longer supported " +
					"in Kong Gateway versions >= 3.0.",
				Resolution: "Please update the plugin configuration to use " +
					"'config.access' field in place of 'config.functions' field",
			},
			SemverRange: versionsPre300,
			Update: config.ConfigTableUpdates{
				Name: "post-function",
				Type: config.Plugin,
				// TODO(hbagdi) figure out a mechanism to introduce warnings
				// without this wasteful update
				FieldUpdates: []config.ConfigTableFieldCondition{
					{
						Field:     "functions",
						Condition: "functions",
					},
				},
			},
		},
	}
)

func init() {
	for _, change := range changes {
		err := config.ChangeRegistry.Register(change)
		if err != nil {
			// ensures that changes are valid at compile time
			panic(err)
		}
	}
}
