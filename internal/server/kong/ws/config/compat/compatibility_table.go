package compat

import (
	"fmt"
	"strings"

	"github.com/kong/koko/internal/server/kong/ws/config"
	"github.com/tidwall/gjson"
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
		"< %s.",
		pluginName,
		versionWithFeatureSupport,
	)
}

func standardCoreEntityMessage(entityName string, versionWithFeatureSupport string) string {
	return fmt.Sprintf("The '%s' entity is being used, "+
		"but Kong Gateway versions < %s do not support this entity. ",
		entityName,
		versionWithFeatureSupport,
	)
}

func standardCoreEntityFieldsMessage(entityName string, fields []string, versionWithFeatureSupport string) string {
	quotedFields := "'" + strings.Join(fields, "', '") + "'"
	return fmt.Sprintf("For the '%s' entity, "+
		"one or more of the following schema fields are set: %s "+
		"but Kong Gateway versions < %s do not support these fields. ",
		entityName,
		quotedFields,
		versionWithFeatureSupport,
	)
}

const (
	versionsPre260      = "< 2.6.0"
	versionsPre270      = "< 2.7.0"
	versionsPre280      = "< 2.8.0"
	versionsPre300      = "< 3.0.0"
	versionsPre310      = "< 3.1.0"
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
				DisableChangeTracking: func(rawJSON string) bool {
					// do not emit change if config.base64_encode_body is
					// set to the default of 'true'
					plugin := gjson.Parse(rawJSON)
					base64EncodeBody := plugin.Get("config.base64_encode_body")
					return base64EncodeBody.Bool()
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
				DisableChangeTracking: func(rawJSON string) bool {
					plugin := gjson.Parse(rawJSON)
					base64EncodeBody := plugin.Get("config.allow_origin_header")
					return base64EncodeBody.String() == "*"
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
				DisableChangeTracking: func(rawJSON string) bool {
					// do not emit change if echo is set to default value of false
					// and trigger is not set
					plugin := gjson.Parse(rawJSON)
					echo := plugin.Get("config.echo").Bool()
					trigger := plugin.Get("config.trigger").Type

					return !echo && trigger == gjson.Null
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
				DisableChangeTracking: func(rawJSON string) bool {
					// do not emit change if all are set to default value
					plugin := gjson.Parse(rawJSON)

					serviceNameTag := plugin.Get("config.service_name_tag")
					if serviceNameTag.String() != "name" {
						return false
					}

					statusTag := plugin.Get("config.status_tag")
					if statusTag.String() != "status" {
						return false
					}

					consumerTag := plugin.Get("config.consumer_tag")
					return consumerTag.String() == "consumer"
				},
			},
		},
		{
			Metadata: config.ChangeMetadata{
				ID:       config.ChangeID("P106"),
				Severity: config.ChangeSeverityError,
				Description: "For the 'datadog' plugin, " +
					"distribution metric type is not supported in Kong gateway " +
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
				DisableChangeTracking: func(rawJSON string) bool {
					plugin := gjson.Parse(rawJSON)

					redisSSl := plugin.Get("config.redis_ssl")
					if redisSSl.Bool() {
						// redis_ssl is set to non-default
						return false
					}

					redisSSLVerify := plugin.Get("config.redis_ssl_verify")
					if redisSSLVerify.Bool() {
						// redis_ssl_verify is set to non-default
						return false
					}

					redisServerName := plugin.Get("config.redis_server_name")
					if redisServerName.Exists() &&
						redisServerName.String() != "" {
						// redis_server_name is set
						return false
					}

					// all values are set to default, disable change tracking
					return true
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
				DisableChangeTracking: func(rawJSON string) bool {
					plugin := gjson.Parse(rawJSON)

					localServiceName := plugin.Get("config.local_service_name")
					return localServiceName.String() == "kong"
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
				DisableChangeTracking: func(rawJSON string) bool {
					plugin := gjson.Parse(rawJSON)

					rsaKeySize := plugin.Get("config.rsa_key_size")
					const defaultRSASize = 4096
					return rsaKeySize.Int() == defaultRSASize
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
		// P114 was removed due to duplication and inaccuracies for associated version upgrade
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
				DisableChangeTracking: func(rawJSON string) bool {
					plugin := gjson.Parse(rawJSON)

					spanName := plugin.Get("config.http_span_name")
					if spanName.Exists() &&
						spanName.String() != "method" {
						return false
					}

					connectTimeout := plugin.Get("config.connect_timeout")
					if connectTimeout.Exists() &&
						connectTimeout.Int() != 2000 {
						return false
					}

					sendTimeout := plugin.Get("config.send_timeout")
					if sendTimeout.Exists() &&
						sendTimeout.Int() != 5000 {
						return false
					}

					readTimeout := plugin.Get("config.read_timeout")
					if readTimeout.Exists() &&
						readTimeout.Int() != 5000 {
						return false
					}

					return true
				},
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
				DisableChangeTracking: func(_ string) bool {
					// always emit change because the default values are
					// breaking in their nature and are bound to surprise users
					return false
				},
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
				DisableChangeTracking: func(rawJSON string) bool {
					plugin := gjson.Parse(rawJSON)

					allowAnyDomain := plugin.Get("config.allow_any_domain")
					if allowAnyDomain.Exists() &&
						allowAnyDomain.Bool() {
						return false
					}
					return true
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
				DisableChangeTracking: func(rawJSON string) bool {
					service := gjson.Parse(rawJSON)

					enabled := service.Get("enabled")
					if enabled.Exists() &&
						!enabled.Bool() {
						return false
					}
					return true
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
				DisableChangeTracking: func(rawJSON string) bool {
					plugin := gjson.Parse(rawJSON)
					// do not emit change if 'proxy_scheme' is not set
					return config.ValueIsEmpty(plugin.Get("config.proxy_scheme"))
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
				DisableChangeTracking: func(rawJSON string) bool {
					// do not emit change if functions is set to default value (empty array)
					plugin := gjson.Parse(rawJSON)
					return config.ValueIsEmpty(plugin.Get("config.functions"))
				},
				FieldUpdates: []config.ConfigTableFieldCondition{
					{
						Field:     "functions",
						Condition: "functions",
						Updates: []config.ConfigTableFieldUpdate{
							{
								Field:            "access",
								ValueFromField:   "functions",
								FieldMustBeEmpty: true,
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
				DisableChangeTracking: func(rawJSON string) bool {
					// do not emit change if functions is set to default value (empty array)
					plugin := gjson.Parse(rawJSON)
					return config.ValueIsEmpty(plugin.Get("config.functions"))
				},
				FieldUpdates: []config.ConfigTableFieldCondition{
					{
						Field:     "functions",
						Condition: "functions",
						Updates: []config.ConfigTableFieldUpdate{
							{
								Field:            "access",
								ValueFromField:   "functions",
								FieldMustBeEmpty: true,
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
		{
			Metadata: config.ChangeMetadata{
				ID:       config.ChangeID("P126"),
				Severity: config.ChangeSeverityError,
				Description: standardCoreEntityFieldsMessage(config.Upstream.String(),
					[]string{
						"hash_on_query_arg",
						"hash_fallback_query_arg",
						"hash_on_uri_capture",
						"hash_fallback_uri_capture",
					},
					"3.0"),
				Resolution: standardUpgradeMessage("3.0"),
			},
			SemverRange: versionsPre300,
			Update: config.ConfigTableUpdates{
				Name: config.Upstream.String(),
				Type: config.Upstream,
				RemoveFields: []string{
					"hash_on_query_arg",
					"hash_fallback_query_arg",
					"hash_on_uri_capture",
					"hash_fallback_uri_capture",
				},
			},
		},
		{
			Metadata: config.ChangeMetadata{
				ID:       config.ChangeID("P127"),
				Severity: config.ChangeSeverityError,
				Description: fmt.Sprintf("For the 'upstreams' entity, "+
					"one or more of the '%s' schema fields are set with one "+
					"of the following values: %s, but Kong Gateway versions < 3.0 "+
					"do not support these values. Because of this, 'hash_on' and 'hash_fallback'"+
					"have been changed in the data-plane to 'none' and hashing is "+
					"not working as expected.",
					strings.Join([]string{"hash_on", "hash_fallback"}, ", "),
					strings.Join([]string{"path", "query_arg", "uri_capture"}, ", "),
				),
				Resolution: standardUpgradeMessage("3.0"),
			},
			SemverRange: versionsPre300,
			Update: config.ConfigTableUpdates{
				Name: config.Upstream.String(),
				Type: config.Upstream,
				FieldUpdates: []config.ConfigTableFieldCondition{
					{
						Field:     "hash_on",
						Condition: "hash_on=path",
						Updates: []config.ConfigTableFieldUpdate{
							{
								Field: "hash_on",
								Value: "none",
							},
						},
					},
					{
						Field:     "hash_on",
						Condition: "hash_on=query_arg",
						Updates: []config.ConfigTableFieldUpdate{
							{
								Field: "hash_on",
								Value: "none",
							},
						},
					},
					{
						Field:     "hash_on",
						Condition: "hash_on=uri_capture",
						Updates: []config.ConfigTableFieldUpdate{
							{
								Field: "hash_on",
								Value: "none",
							},
						},
					},
					{
						Field:     "hash_fallback",
						Condition: "hash_fallback=path",
						Updates: []config.ConfigTableFieldUpdate{
							{
								Field: "hash_fallback",
								Value: "none",
							},
						},
					},
					{
						Field:     "hash_fallback",
						Condition: "hash_fallback=query_arg",
						Updates: []config.ConfigTableFieldUpdate{
							{
								Field: "hash_fallback",
								Value: "none",
							},
						},
					},
					{
						Field:     "hash_fallback",
						Condition: "hash_fallback=uri_capture",
						Updates: []config.ConfigTableFieldUpdate{
							{
								Field: "hash_fallback",
								Value: "none",
							},
						},
					},
				},
			},
		},
		{
			Metadata: config.ChangeMetadata{
				ID:       config.ChangeID("P133"),
				Severity: config.ChangeSeverityError,
				Description: standardPluginFieldsMessage("statsd",
					[]string{
						"allow_status_codes",
						"udp_packet_size",
						"use_tcp",
						"hostname_in_prefix",
						"consumer_identifier_default",
						"service_identifier_default",
						"workspace_identifier_default",
					},
					"3.0", false),
				Resolution: standardUpgradeMessage("3.0"),
			},
			SemverRange: versionsPre300,
			Update: config.ConfigTableUpdates{
				Name: "statsd",
				Type: config.Plugin,
				RemoveFields: []string{
					"allow_status_codes",
					"udp_packet_size",
					"use_tcp",
					"hostname_in_prefix",
					"consumer_identifier_default",
					"service_identifier_default",
					"workspace_identifier_default",
				},
			},
		},
		{
			Metadata: config.ChangeMetadata{
				ID:          config.ChangeID("P135"),
				Severity:    config.ChangeSeverityError,
				Description: standardCoreEntityMessage("vault", "3.1"),
				Resolution:  standardUpgradeMessage("3.1"),
			},
			SemverRange: versionsPre300,
			Update: config.ConfigTableUpdates{
				Name:   config.Vault.String(),
				Type:   config.Vault,
				Remove: true,
			},
		},
		{
			Metadata: config.ChangeMetadata{
				ID:       config.ChangeID("P136"),
				Severity: config.ChangeSeverityError,
				Description: standardPluginFieldsMessage("zipkin",
					[]string{"http_response_header_for_traceid"},
					"3.1", false),
				Resolution: standardUpgradeMessage("3.1"),
			},
			SemverRange: versionsPre310,
			Update: config.ConfigTableUpdates{
				Name:         "zipkin",
				Type:         config.Plugin,
				RemoveFields: []string{"http_response_header_for_traceid"},
				DisableChangeTracking: func(rawJSON string) bool {
					plugin := gjson.Parse(rawJSON)
					traceHeader := plugin.Get("config.http_response_header_for_traceid")
					return traceHeader.Exists() && traceHeader.Type == gjson.Null
				},
			},
		},
		{
			Metadata: config.ChangeMetadata{
				ID:       config.ChangeID("P137"),
				Severity: config.ChangeSeverityError,
				Description: standardPluginFieldsMessage("rate-limiting",
					[]string{"error_code", "error_message"},
					"3.1", false),
				Resolution: standardUpgradeMessage("3.1"),
			},
			SemverRange: versionsPre310,
			Update: config.ConfigTableUpdates{
				Name:         "rate-limiting",
				Type:         config.Plugin,
				RemoveFields: []string{"error_code", "error_message"},
				DisableChangeTracking: func(rawJSON string) bool {
					plugin := gjson.Parse(rawJSON)

					providedErrorMessage := plugin.Get("config.error_message")
					if providedErrorMessage.Exists() &&
						!strings.Contains(providedErrorMessage.String(), "API rate limit exceeded") {
						return false
					}

					providedErrorCode := plugin.Get("config.error_code")
					if providedErrorCode.Exists() && providedErrorCode.Int() != 429 {
						return false
					}
					return true
				},
			},
		},
		{
			Metadata: config.ChangeMetadata{
				ID:       config.ChangeID("P138"),
				Severity: config.ChangeSeverityError,
				Description: standardPluginFieldsMessage("acme",
					[]string{
						"storage_config.redis.ssl",
						"storage_config.redis.ssl_verify",
						"storage_config.redis.ssl_server_name",
					},
					"3.1", false),
				Resolution: standardUpgradeMessage("3.1"),
			},
			SemverRange: versionsPre310,
			Update: config.ConfigTableUpdates{
				Name: "acme",
				Type: config.Plugin,
				RemoveFields: []string{
					"storage_config.redis.ssl",
					"storage_config.redis.ssl_verify",
					"storage_config.redis.ssl_server_name",
				},
				DisableChangeTracking: func(rawJSON string) bool {
					plugin := gjson.Parse(rawJSON)
					if plugin.Get("config.storage_config.redis.ssl").Bool() {
						return false
					}
					if plugin.Get("config.storage_config.redis.ssl_verify").Bool() {
						return false
					}
					sslServerName := plugin.Get("config.storage_config.redis.ssl_server_name")
					return sslServerName.Exists() && sslServerName.Type == gjson.Null
				},
			},
		},
		{
			Metadata: config.ChangeMetadata{
				ID:       config.ChangeID("P139"),
				Severity: config.ChangeSeverityError,
				Description: standardPluginFieldsMessage("response-ratelimiting",
					[]string{
						"redis_ssl",
						"redis_ssl_verify",
						"redis_server_name",
					},
					"3.1", false),
				Resolution: standardUpgradeMessage("3.1"),
			},
			SemverRange: versionsPre310,
			Update: config.ConfigTableUpdates{
				Name: "response-ratelimiting",
				Type: config.Plugin,
				RemoveFields: []string{
					"redis_ssl",
					"redis_ssl_verify",
					"redis_server_name",
				},
				DisableChangeTracking: func(rawJSON string) bool {
					plugin := gjson.Parse(rawJSON)
					if plugin.Get("config.redis_ssl").Bool() {
						return false
					}
					if plugin.Get("config.redis_ssl_verify").Bool() {
						return false
					}
					sslServerName := plugin.Get("config.redis_server_name")
					return sslServerName.Exists() && sslServerName.Type == gjson.Null
				},
			},
		},
		{
			Metadata: config.ChangeMetadata{
				ID:       config.ChangeID("P140"),
				Severity: config.ChangeSeverityError,
				Description: standardPluginFieldsMessage("aws-lambda",
					[]string{
						"aws_assume_role_arn",
						"aws_role_session_name",
					}, "3.0", false),
				Resolution: standardUpgradeMessage("3.0"),
			},
			SemverRange: versionsPre300,
			Update: config.ConfigTableUpdates{
				Name: "aws-lambda",
				Type: config.Plugin,
				RemoveFields: []string{
					"aws_assume_role_arn",
					"aws_role_session_name",
				},
				DisableChangeTracking: func(rawJSON string) bool {
					plugin := gjson.Parse(rawJSON)
					// do not emit change if 'aws_assume_role_arn' is not set and
					// 'aws_assume_role_arn' is set to its default value
					if !config.ValueIsEmpty(plugin.Get("config.aws_assume_role_arn")) {
						return false
					}
					awsRoleSessionName := plugin.Get("config.aws_role_session_name")
					if awsRoleSessionName.Exists() &&
						awsRoleSessionName.String() != "kong" {
						return false
					}
					return true
				},
			},
		},
		{
			Metadata: config.ChangeMetadata{
				ID:          config.ChangeID("P141"),
				Severity:    config.ChangeSeverityError,
				Description: standardCoreEntityMessage("key", "3.1"),
				Resolution:  standardUpgradeMessage("3.1"),
			},
			SemverRange: versionsPre310,
			Update: config.ConfigTableUpdates{
				Name:   config.Key.String(),
				Type:   config.Key,
				Remove: true,
			},
		},
		{
			Metadata: config.ChangeMetadata{
				ID:          config.ChangeID("P142"),
				Severity:    config.ChangeSeverityError,
				Description: standardCoreEntityMessage("key-set", "3.1"),
				Resolution:  standardUpgradeMessage("3.1"),
			},
			SemverRange: versionsPre310,
			Update: config.ConfigTableUpdates{
				Name:   config.KeySet.String(),
				Type:   config.KeySet,
				Remove: true,
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
