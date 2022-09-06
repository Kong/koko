package compat

import (
	"fmt"
	"strings"

	"github.com/kong/koko/internal/resource"
	"github.com/kong/koko/internal/server/kong/ws/config"
	"github.com/kong/koko/internal/versioning"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"go.uber.org/zap"
)

var versionOlderThan300 = versioning.MustNewRange("< 3.0.0")

const (
	awsLambdaExclusiveFieldChangeID = "P121"
	pathRegexFieldChangeID          = "P130"
)

func init() {
	if err := config.ChangeRegistry.Register(config.Change{
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
	}); err != nil {
		panic(err)
	}

	if err := config.ChangeRegistry.Register(config.Change{
		Metadata: config.ChangeMetadata{
			ID:       pathRegexFieldChangeID,
			Severity: config.ChangeSeverityWarning,
			Description: "For the 'paths' field used in route a regex " +
				"pattern usage was detected. This field has been de-normalized " +
				"replacing '%%' with '%25' and stripping the prefix '~'. " +
				"Routes which rely on regex parsing may not work as intended " +
				"on Kong gateway versions < 3.0. " +
				"If upgrading to version 3.0 is not possible, using paths " +
				"without the '~' prefix will avoid this warning.",
			Resolution:       standardUpgradeMessage("3.0"),
			DocumentationURL: "",
		},
		SemverRange: versionsPre300,
		// none since the logic is hard-coded instead
		Update: config.ConfigTableUpdates{},
	}); err != nil {
		panic(err)
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

// denormalizePath transforms a single path pattern into
// pre-3.0 format by removing the '~' for regexes
// and the minimal revertion of url-normalization.
func denormalizePath(path string) (string, error) {
	path = strings.TrimPrefix(path, "~")
	if !strings.HasPrefix(path, "/") {
		return path, fmt.Errorf("invalid path %v", path)
	}
	path = strings.ReplaceAll(path, "%%", "%25")
	return path, nil
}

// correctRoutesPathField changes any regex path matching pattern
// from the 3.0 style (where regex must start with '~/' and prefix paths
// start with '/') into the previous format (where all paths start with
// '/' and regexes are auto detected).
func correctRoutesPathField(payload string,
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
				path, err := denormalizePath(path)
				if err != nil {
					logger.Error("failed to denormalize route path", zap.Error(err),
						zap.String("route-id", routeID),
						zap.String("path", path))
					continue
				}
				if !modifiedRoute {
					err = tracker.TrackForResource(pathRegexFieldChangeID,
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

func VersionCompatibilityExtraProcessing(payload string, dataPlaneVersion versioning.Version,
	tracker *config.ChangeTracker, logger *zap.Logger,
) (string, error) {
	dataPlaneVersionStr := dataPlaneVersion.String()
	processedPayload := payload

	if versionOlderThan300(dataPlaneVersion) {
		// 'aws_region' and 'host' are mutually exclusive for DP < 3.x
		processedPayload = correctAWSLambdaMutuallyExclusiveFields(
			processedPayload, dataPlaneVersionStr, tracker, logger)

		// The `headers` field on the `http-log` plugin changed from an array of strings to just a single string
		// for DP's >= 3.0. As such, we need to transform the headers back to a single string within an array.
		var err error
		if processedPayload, err = correctHTTPLogHeadersField(processedPayload); err != nil {
			return "", err
		}

		// remove '~' prefix from paths
		processedPayload, err = correctRoutesPathField(processedPayload, dataPlaneVersionStr, tracker, logger)
		if err != nil {
			return "", err
		}
	}

	return processedPayload, nil
}
