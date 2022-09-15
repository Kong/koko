package compat

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/kong/koko/internal/json"
	"github.com/kong/koko/internal/model"
	"github.com/kong/koko/internal/resource"
	"github.com/kong/koko/internal/server/kong/ws/config"
	"github.com/kong/koko/internal/versioning"
	"github.com/samber/lo"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"go.uber.org/zap"
)

var (
	versionOlderThan300 = versioning.MustNewRange("< 3.0.0")
	version300OrNewer   = versioning.MustNewRange(">= 3.0.0")
)

var (
	statsdDefaultIdentifiers = map[string]string{
		"consumer_identifier":  "custom_id",
		"service_identifier":   "",
		"workspace_identifier": "",
	}

	unsupportedMetricsPre300 = []string{
		"shdict_usage",
		"status_count_per_user_per_route",
		"status_count_per_workspace",
	}

	defaultMetrics = map[string][]string{
		"kong_latency":          {},
		"latency":               {},
		"request_count":         {},
		"request_per_user":      {"consumer_identifier"},
		"request_size":          {},
		"response_size":         {},
		"status_count":          {},
		"status_count_per_user": {"consumer_identifier"},
		"unique_users":          {"consumer_identifier"},
		"upstream_latency":      {},
	}
)

const (
	awsLambdaExclusiveFieldChangeID          = "P121"
	pathRegexFieldChangeID                   = "P128"
	pathRegexFieldUnprefixedChangeID         = "P129"
	statsdUnsupportedMetricChangeID          = "P130"
	statsdUnsupportedMetricFieldChangeID     = "P131"
	statsdAddDefaultMetricFieldValueChangeID = "P132"
	dropSpacesInTagsChangeID                 = "P134"
)

func init() {
	for _, change := range []config.Change{
		{
			Metadata: config.ChangeMetadata{
				ID:       awsLambdaExclusiveFieldChangeID,
				Severity: config.ChangeSeverityError,
				Description: "For the 'aws-lambda' plugin, " +
					"'config.aws_region' and 'config.host' fields are set. " +
					"These fields were mutually exclusive for Kong gateway " +
					"versions < 3.0. The plugin configuration has been changed " +
					"to remove the 'config.host' field.",
				Resolution:       standardUpgradeMessage("3.0"),
				DocumentationURL: "",
			},
			SemverRange: versionsPre300,
			// none since the logic is hard-coded instead
			Update: config.ConfigTableUpdates{},
		},
		{
			Metadata: config.ChangeMetadata{
				ID:       pathRegexFieldChangeID,
				Severity: config.ChangeSeverityWarning,
				Description: "For the 'paths' field used in route a regex " +
					"pattern usage was detected. " +
					"This field has been back-ported by stripping the prefix '~'. " +
					"Routes which rely on regex parsing may not work as intended " +
					"on Kong gateway versions < 3.0. " +
					"If upgrading to version 3.0 is not possible, using paths " +
					"without the '~' prefix will avoid this warning.",
				Resolution: standardUpgradeMessage("3.0"),
			},
			SemverRange: versionsPre300,
			// none since the logic is hard-coded instead
			Update: config.ConfigTableUpdates{},
		},
		{
			Metadata: config.ChangeMetadata{
				ID:       dropSpacesInTagsChangeID,
				Severity: config.ChangeSeverityError,
				Description: "Tags that contain space characters (' ') are not supported on Kong gateway " +
					"versions < 3.0. As such, all tag values that contain space characters have been removed.",
				Resolution: standardUpgradeMessage("3.0"),
			},
			SemverRange: versionsPre300,
		},
		{
			Metadata: config.ChangeMetadata{
				ID:       pathRegexFieldUnprefixedChangeID,
				Severity: config.ChangeSeverityWarning,
				Description: "For the 'paths' field used in route a regex " +
					"pattern usage was detected. " +
					"Kong gateway versions 3.0 and above require that regular expressions " +
					"start with a '~' character to distinguish from simple prefix match.",
				Resolution: "To define a regular expression based path for routing, please " +
					"prefix the path with ~ character.",
			},
			SemverRange: versions300AndAbove,
		},
	} {
		if err := config.ChangeRegistry.Register(change); err != nil {
			panic(err)
		}
	}
}

// correctAWSLambdaMutuallyExclusiveFields handles 'aws_region' and 'host' fields, which were
// mutually exclusive until Kong version 2.8 but both are accepted in 3.x. If both are set
// with DPs < 3.x, the 'host' field will be dropped in order to prevent a failure in the DP.
func correctAWSLambdaMutuallyExclusiveFields(
	payload string,
	dataPlaneVersionStr string,
	tracker *config.ChangeTracker,
	logger *zap.Logger,
) string {
	pluginName := "aws-lambda"
	processedPayload := payload
	results := gjson.Get(processedPayload, "config_table.plugins.#(name=aws-lambda)#")
	indexUpdate := 0
	for _, res := range results.Array() {
		updatedRaw := res.Raw
		awsRegionResult := res.Get("config.aws_region")
		hostResult := res.Get("config.host")
		if awsRegionResult.Exists() && hostResult.Exists() {
			var (
				err      error
				pluginID = res.Get("id").String()
			)
			err = tracker.TrackForResource(awsLambdaExclusiveFieldChangeID,
				config.ResourceInfo{
					Type: string(resource.TypePlugin),
					ID:   pluginID,
				})
			if err != nil {
				logger.Error("failed to track version compatibility"+
					" change",
					zap.String("change-id", awsLambdaExclusiveFieldChangeID),
					zap.String("resource-type", "plugin"))
			}
			if updatedRaw, err = sjson.Delete(updatedRaw, "config.host"); err != nil {
				logger.With(zap.String("plugin", pluginName)).
					With(zap.String("field", "host")).
					With(zap.String("data-plane", dataPlaneVersionStr)).
					With(zap.Error(err)).
					Error("plugin configuration field was not removed from configuration")
			} else {
				logger.With(zap.String("plugin", pluginName)).
					With(zap.String("field", "host")).
					With(zap.String("data-plane", dataPlaneVersionStr)).
					Warn("removing plugin configuration field which is incompatible with data plane")
			}
		}

		// Update the processed payload
		resIndex := res.Index - indexUpdate
		updatedPayload := processedPayload[:resIndex] + updatedRaw +
			processedPayload[resIndex+len(res.Raw):]
		indexUpdate += len(processedPayload) - len(updatedPayload)
		processedPayload = updatedPayload
	}
	return processedPayload
}

func correctHTTPLogHeadersField(payload string) (string, error) {
	for i, plugin := range gjson.Get(payload, "config_table.plugins").Array() {
		if plugin.Get("name").Str != "http-log" {
			continue
		}

		for _, headers := range plugin.Get("config.headers").Array() {
			// When `headers.Type` is not a JSON object, no headers have been set, so we'll re-write it to be null.
			// This matches the behavior of the gateway when the headers field exists, but no value is defined.
			var newHeadersIface interface{}

			// Handle transforming the headers on each plugin from an object consisting
			// of a single string (`{"header-1": "value-1"}`), to an object consisting
			// of a single string value in an array (`{"header-1": ["value-1"]}`).
			if headers.Type == gjson.JSON {
				newHeaders := make(map[string][]string, len(headers.Indexes))
				newHeadersIface = newHeaders
				for key, values := range headers.Map() {
					// In <=2.8, while it is possible to set a header with an empty array of values,
					// the data plane won't send the header to the HTTP log endpoint with no value to
					// the defined HTTP endpoint. To match this behavior, we'll remove the header.
					if len(values.Str) == 0 {
						continue
					}

					newHeaders[key] = []string{values.Str}
				}
			}

			// Replace the headers for the http-log plugin.
			var err error
			if payload, err = sjson.SetOptions(
				payload,
				fmt.Sprintf("config_table.plugins.%d.config.headers", i),
				newHeadersIface,
				&sjson.Options{Optimistic: true},
			); err != nil {
				return "", err
			}
		}
	}

	return payload, nil
}

var regexPattern = regexp.MustCompile(`[^a-zA-Z0-9._~/%-]`)

func isRegexLike(path string) bool {
	return regexPattern.MatchString(path)
}

// migrateRoutesPathFieldPre300 changes any regex path matching pattern
// from the 3.0 style (where regex must start with '~/' and prefix paths
// start with '/') into the previous format (where all paths start with
// '/' and regexes are auto detected).
func migrateRoutesPathFieldPre300(payload string,
	dataPlaneVersion string,
	tracker *config.ChangeTracker,
	logger *zap.Logger,
) (string, error) {
	routes := gjson.Get(payload, "config_table.routes")

	processedPayload := payload
	logger = logger.With(zap.String("data-plane", dataPlaneVersion),
		zap.String("change-id", pathRegexFieldChangeID),
		zap.String("resource-type", "route"))

	for _, route := range routes.Array() {
		routeID := route.Get("id").Str
		pathsGJ := route.Get("paths")
		if !pathsGJ.Exists() || !pathsGJ.IsArray() {
			continue
		}

		modifiedRoute := false
		paths, ok := pathsGJ.Value().([]any)
		if !ok {
			continue
		}
		for j, pathIntf := range paths {
			path, ok := pathIntf.(string)
			if !ok {
				continue
			}
			if strings.HasPrefix(path, "~/") {
				path = strings.TrimPrefix(path, "~")
				if !modifiedRoute {
					err := tracker.TrackForResource(pathRegexFieldChangeID,
						config.ResourceInfo{
							Type: string(resource.TypeRoute),
							ID:   routeID,
						})
					if err != nil {
						logger.Error("failed to track version compatibility change", zap.Error(err),
							zap.String("route-id", routeID),
							zap.String("path", path))
						continue
					}
					modifiedRoute = true
				}
				paths[j] = path
			}
		}

		if modifiedRoute {
			var err error
			processedPayload, err = sjson.Set(processedPayload, route.Path(payload)+".paths", paths)
			if err != nil {
				logger.Error("failed to set processed paths", zap.Error(err),
					zap.String("route-id", routeID))
				return "", err
			}
		}
	}

	return processedPayload, nil
}

func checkRoutePaths300AndAbove(paths []interface{},
	routeID string,
	tracker *config.ChangeTracker,
) error {
	for i, pathIntf := range paths {
		path, ok := pathIntf.(string)
		if !ok {
			return fmt.Errorf("path #%d, expected string, found %T", i, pathIntf)
		}
		if strings.HasPrefix(path, "~/") || !isRegexLike(path) {
			continue
		}

		err := tracker.TrackForResource(pathRegexFieldUnprefixedChangeID,
			config.ResourceInfo{
				Type: string(resource.TypeRoute),
				ID:   routeID,
			})
		if err != nil {
			return fmt.Errorf("failed to track version compatibility change: %w", err)
		}
		return fmt.Errorf("found non-prefixed regex-like path")
	}
	return nil
}

func checkRoutesPathFieldPost300(payload string,
	dataPlaneVersion string,
	tracker *config.ChangeTracker,
	logger *zap.Logger,
) string {
	routes := gjson.Get(payload, "config_table.routes")

	logger = logger.With(zap.String("data-plane", dataPlaneVersion),
		zap.String("change-id", pathRegexFieldUnprefixedChangeID),
		zap.String("resource-type", "route"))

	for _, route := range routes.Array() {
		routeID := route.Get("id").Str
		pathsGJ := route.Get("paths")
		if !pathsGJ.Exists() || !pathsGJ.IsArray() {
			continue
		}

		paths, ok := pathsGJ.Value().([]interface{})
		if !ok {
			continue
		}

		err := checkRoutePaths300AndAbove(paths, routeID, tracker)
		if err != nil {
			logger.Error("verifying route paths for DP versions 3.0 and above", zap.Error(err),
				zap.String("route-id", routeID))
		}
	}

	return payload
}

func updateFormatVersion(payload string,
	dataPlaneVersion string,
	logger *zap.Logger,
) string {
	payload, err := sjson.Set(payload, "config_table._format_version", "3.0")
	if err != nil {
		logger.Error("failed to update \"_format_version\" parameter", zap.Error(err),
			zap.String("data-plane", dataPlaneVersion))
	}
	return payload
}

// correctStatsdMetrics addresses the needed statsd 3.0 schema changes for older DPs.
//
// The changes done in 3.0:
//   - remove some 'hard-coded' default values for metric identifiers and
//     introduces specific schema fields to define those defaults.
//   - remove some newly introduced metrics (listed in unsupportedMetricsPre300)
//
// This function ensures the new schema works for older DPs by setting the missing default
// values for the metric identifiers.
func correctStatsdMetrics(
	payload string,
	dataPlaneVersionStr string,
	tracker *config.ChangeTracker,
	logger *zap.Logger,
) string {
	log := logger.With(zap.String("plugin", "statsd")).
		With(zap.String("data-plane", dataPlaneVersionStr))

	processedPayload := payload
	indexUpdate := 0
	for _, res := range gjson.Get(processedPayload, "config_table.plugins.#(name=statsd)#").Array() {
		var (
			err         error
			metricsJSON []interface{}
			metrics     []string
			updatedRaw  = res.Raw
			pluginID    = res.Get("id").String()
		)
		for _, metric := range gjson.Get(updatedRaw, "config.metrics").Array() {
			metricRaw := metric.Raw
			name := metric.Get("name").String()
			// Skip unsupported metrics.
			if lo.Contains(unsupportedMetricsPre300, name) {
				err = tracker.TrackForResource(statsdUnsupportedMetricChangeID,
					config.ResourceInfo{
						Type: string(resource.TypePlugin),
						ID:   pluginID,
					})
				if err != nil {
					logger.Error("failed to track version compatibility change",
						zap.String("change-id", statsdUnsupportedMetricChangeID),
						zap.String("resource-type", "plugin"))
				}
				logger.With(zap.String("plugin", "statsd")).
					With(zap.String("metric", name)).
					With(zap.String("data-plane", dataPlaneVersionStr)).
					Warn("removing statsd plugin metric which is incompatible with data plane")
				continue
			}
			// Only evaluate default metrics.
			identifiers, ok := defaultMetrics[name]
			if !ok {
				metrics = append(metrics, metricRaw)
				continue
			}
			for key, defaultValue := range statsdDefaultIdentifiers {
				identifier := metric.Get(key)
				if lo.Contains(identifiers, key) {
					if !identifier.Exists() || identifier.String() == "" {
						log := log.With(zap.String("metric", name)).
							With(zap.String("field", key)).
							With(zap.String("condition", "missing value")).
							With(zap.String("new-value", defaultValue))

						err = tracker.TrackForResource(statsdAddDefaultMetricFieldValueChangeID,
							config.ResourceInfo{
								Type: string(resource.TypePlugin),
								ID:   pluginID,
							})
						if err != nil {
							logger.Error("failed to track version compatibility change",
								zap.String("change-id", statsdAddDefaultMetricFieldValueChangeID),
								zap.String("resource-type", "plugin"))
						}
						if metricRaw, err = sjson.Set(metricRaw, key, defaultValue); err != nil {
							log.With(zap.Error(err)).
								Error("statsd plugin metric was not updated in configuration")
						} else {
							log.Warn("updating statsd plugin metric which is incompatible with data plane")
						}
					}
				} else {
					if identifier.Exists() {
						log := log.With(zap.String("metric", name)).
							With(zap.String("field", key))

						err = tracker.TrackForResource(statsdUnsupportedMetricFieldChangeID,
							config.ResourceInfo{
								Type: string(resource.TypePlugin),
								ID:   pluginID,
							})
						if err != nil {
							logger.Error("failed to track version compatibility change",
								zap.String("change-id", statsdUnsupportedMetricFieldChangeID),
								zap.String("resource-type", "plugin"))
						}
						if metricRaw, err = sjson.Delete(metricRaw, key); err != nil {
							log.With(zap.Error(err)).
								Error("statsd plugin metric was not removed in configuration")
						} else {
							log.Warn("removing statsd plugin metric which is incompatible with data plane")
						}
					}
				}
			}
			metrics = append(metrics, metricRaw)
		}
		metricsBytes := []byte(fmt.Sprintf("[%v]", strings.Join(metrics, ",")))
		if err = json.Unmarshal(metricsBytes, &metricsJSON); err != nil {
			log.With(zap.Error(err)).Error("statsd plugin metrics were not updated in configuration")
			continue
		}
		if updatedRaw, err = sjson.Set(updatedRaw, "config.metrics", metricsJSON); err != nil {
			log.With(zap.Any("new-value", metricsJSON)).
				With(zap.Error(err)).
				Error("statsd plugin metrics were not updated in configuration")
		}

		// Update the processed payload.
		resIndex := res.Index - indexUpdate
		updatedPayload := processedPayload[:resIndex] + updatedRaw +
			processedPayload[resIndex+len(res.Raw):]
		indexUpdate += len(processedPayload) - len(updatedPayload)
		processedPayload = updatedPayload
	}

	return processedPayload
}

// dropTagsWithSpaces handles dropping tags with spaces for < 3.0 DPs, across all entities in the config table.
//
// In Kong 3.0, it introduced support for space characters (` `) in tags. For DPs < 3.0, we've decided to drop such
// tags that contain spaces, as we did not want to attempt to replace the space character with something else.
func dropTagsWithSpaces(
	payload string,
	dataPlaneVersionStr string,
	tracker *config.ChangeTracker,
	logger *zap.Logger,
) (string, error) {
	log := logger.With(zap.String("data-plane", dataPlaneVersionStr))

	for key, resources := range gjson.Get(payload, "config_table").Map() {
		typ := string(getTypeFromKongKeyName(key))
		log = log.With(zap.String("resource-type", typ))

		for rIdx, r := range resources.Array() {
			resourceID, tagsObj := r.Get("id").Str, r.Get("tags")
			log = log.With(zap.String("resource-id", resourceID))

			// This resource does not have tags, so we can skip it.
			if tagsObj.Type == gjson.Null {
				continue
			}

			var tags []string
			if err := json.Unmarshal([]byte(tagsObj.Raw), &tags); err != nil {
				return "", err
			}

			// Remove all tags that have any number of spaces.
			filteredTags := lo.Filter(tags, func(tag string, _ int) bool {
				return !strings.ContainsRune(tag, ' ')
			})

			// Nothing to do when no tags contain spaces.
			if len(tags) == len(filteredTags) {
				continue
			}

			tagsKey := fmt.Sprintf("config_table.%s.%d.tags", key, rIdx)
			var err error
			if len(filteredTags) == 0 {
				// All tags contained spaces, so we'll delete the tags key.
				payload, err = sjson.Delete(payload, tagsKey)
			} else {
				// Some tags contained spaces, so we'll keep the tags around that do not contain spaces.
				payload, err = sjson.Set(payload, tagsKey, filteredTags)
			}
			if err != nil {
				return "", err
			}

			if err := tracker.TrackForResource(
				dropSpacesInTagsChangeID,
				config.ResourceInfo{ID: resourceID, Type: typ},
			); err != nil {
				log.Error(
					"failed to track version compatibility change",
					zap.String("change-id", dropSpacesInTagsChangeID),
				)
			}
		}
	}

	return payload, nil
}

func VersionCompatibilityExtraProcessing(payload string, dataPlaneVersion versioning.Version,
	tracker *config.ChangeTracker, logger *zap.Logger,
) (string, error) {
	dataPlaneVersionStr := dataPlaneVersion.String()
	processedPayload := payload

	var err error
	if versionOlderThan300(dataPlaneVersion) {
		// 'aws_region' and 'host' are mutually exclusive for DP < 3.x
		processedPayload = correctAWSLambdaMutuallyExclusiveFields(
			processedPayload, dataPlaneVersionStr, tracker, logger)

		// The `headers` field on the `http-log` plugin changed from an array of strings to just a single string
		// for DP's >= 3.0. As such, we need to transform the headers back to a single string within an array.
		if processedPayload, err = correctHTTPLogHeadersField(processedPayload); err != nil {
			return "", err
		}

		// remove '~' prefix from paths
		processedPayload, err = migrateRoutesPathFieldPre300(processedPayload, dataPlaneVersionStr, tracker, logger)
		if err != nil {
			return "", err
		}

		// Correct default metrics identifier for statsd in 2.x.x.x.
		processedPayload = correctStatsdMetrics(processedPayload, dataPlaneVersionStr, tracker, logger)

		// Remove all tags that contain spaces, as it is only supported by DPs >= 3.0.
		if processedPayload, err = dropTagsWithSpaces(
			processedPayload,
			dataPlaneVersionStr,
			tracker,
			logger,
		); err != nil {
			return "", err
		}
	}

	if version300OrNewer(dataPlaneVersion) {
		processedPayload = updateFormatVersion(processedPayload, dataPlaneVersionStr, logger)
		processedPayload = checkRoutesPathFieldPost300(processedPayload, dataPlaneVersionStr, tracker, logger)
	}

	return processedPayload, nil
}

// getTypeFromKongKeyName translates Kong type names used in its config, like `plugins`,
// `routes`, `services`, etc., to Koko's own type name, which happens to be the
// singular version of Kong's type name, e.g.: `plugin`, `route`, `service`, etc.
func getTypeFromKongKeyName(key string) model.Type {
	// TODO(tjasko): We'll likely want an explicit mapping for this in the future, however, some care would
	//  need to be done to ensure such mapping is maintained in the event new resources are created.
	return model.Type(key[:len(key)-1])
}
