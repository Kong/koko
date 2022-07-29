package config

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/kong/go-kong/kong"
	"github.com/kong/koko/internal/json"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"go.uber.org/zap"
)

const (
	KongGatewayCompatibilityVersion = "3.0.0"

	buildVersionPattern = `(?P<build_version>^[0-9]+)[a-zA-Z\-]*`
	invalidVersionOctet = 1000
	majorVersionBase    = 1000000000
	minorVersionBase    = majorVersionBase / 1000
	patchVersionBase    = minorVersionBase / 1000
	base                = 10
	bitSize             = 64
)

var buildVersionRegex = regexp.MustCompile(buildVersionPattern)

type VersionedConfigUpdates map[uint64][]ConfigTableUpdates

type VersionCompatibility interface {
	AddConfigTableUpdates(configTableUpdates VersionedConfigUpdates) error
	ProcessConfigTableUpdates(dataPlaneVersionStr string, compressedPayload []byte) ([]byte, error)
}

type VersionCompatibilityOpts struct {
	Logger         *zap.Logger
	KongCPVersion  string
	ExtraProcessor Processor
}

type UpdateType uint8

const (
	Plugin UpdateType = iota
	Service
)

func (u UpdateType) String() string {
	return [...]string{"plugins", "services"}[u]
}

//nolint: revive
type ConfigTableFieldUpdate struct {
	// Field to perform update or delete operation on; if Value is nil or
	// ValueFromField is empty the field will be removed.
	Field string
	// Value when specified is the value applied to the key referenced in the
	// member Field.
	Value interface{}
	// ValueFromField when specified represents the name of the key whose value is
	// retrieved and applied to the key referenced in the member Field.
	ValueFromField string
}

//nolint: revive
type ConfigTableFieldCondition struct {
	// Field is a top-level or nested field; use dot notation for nested fields.
	Field string
	// Condition is an expression for matching criteria.
	// uses gjson path syntax; https://github.com/tidwall/gjson#path-syntax
	Condition string
	// Updates is an array of updates to perform based on the matched criteria.
	Updates []ConfigTableFieldUpdate
}

//nolint: revive
type ConfigTableUpdates struct {
	// Name is the name of the configuration or field depending on UpdateType.
	Name string
	// UpdateType is the type of update being performed; currently plugins only.
	Type UpdateType
	// RemoveFields will remove fields from the configuration table.
	//
	// Values contained within this array are associated with a field name inside
	// the configuration table. The value can be a top-level field or a nested
	// field which is separated using the dot notation.
	RemoveFields []string
	// RemoveElementsFromArray will remove an array element from a field in the
	// configuration table.
	//
	// Values contained within this array are associated with a field and a
	// condition. If the condition matches for the given field, the array in which
	// the match occurred (e.g. index) will be removed from  the configuration
	// table.
	RemoveElementsFromArray []ConfigTableFieldCondition
	// FieldUpdates will create/update a field to a new value in the configuration
	// table.
	//
	// Values contained within this array are associated with a field, a
	// condition, and an array of updates to perform. New fields can be created,
	// original fields can be updated, and specific fields can be removed
	// depending on the version compatibility requirement of the data plane
	// version being targeted.
	FieldUpdates []ConfigTableFieldCondition
	// Remove indicates whether the whole entity should be removed or not.
	Remove bool
}

type Processor func(uncompressedPayload string, dataPlaneVersion uint64,
	isEnterprise bool, logger *zap.Logger) (string, error)

type WSVersionCompatibility struct {
	logger             *zap.Logger
	kongCPVersion      uint64
	configTableUpdates map[uint64][]ConfigTableUpdates
	extraProcessor     Processor
}

func NewVersionCompatibilityProcessor(opts VersionCompatibilityOpts) (*WSVersionCompatibility, error) {
	if opts.Logger == nil {
		return nil, fmt.Errorf("opts.Logger required")
	}
	if len(strings.TrimSpace(opts.KongCPVersion)) == 0 {
		return nil, fmt.Errorf("opts.KongCPVersion required")
	}
	controlPlaneVersion, err := parseSemanticVersion(opts.KongCPVersion)
	if err != nil {
		return nil, fmt.Errorf("unable to parse opts.KongCPVersion %v", err)
	}

	return &WSVersionCompatibility{
		logger:             opts.Logger,
		kongCPVersion:      controlPlaneVersion,
		configTableUpdates: make(map[uint64][]ConfigTableUpdates),
		extraProcessor:     opts.ExtraProcessor,
	}, nil
}

func (vc *WSVersionCompatibility) AddConfigTableUpdates(payloadUpdates VersionedConfigUpdates) error {
	for version, updates := range payloadUpdates {
		// Handle restriction for FieldUpdates
		for _, update := range updates {
			for _, fieldUpdates := range update.FieldUpdates {
				for _, fieldUpdate := range fieldUpdates.Updates {
					if fieldUpdate.Value != nil && len(fieldUpdate.ValueFromField) > 0 {
						return fmt.Errorf("'Value' and 'ValueFromField' are mutually exclusive")
					}
				}
			}
		}

		vc.configTableUpdates[version] = append(vc.configTableUpdates[version], updates...)
	}
	return nil
}

func (vc *WSVersionCompatibility) ProcessConfigTableUpdates(dataPlaneVersionStr string,
	compressedPayload []byte,
) ([]byte, error) {
	dataPlaneVersion, err := parseSemanticVersion(dataPlaneVersionStr)
	if err != nil {
		return nil, fmt.Errorf("unable to parse data plane version: %v", err)
	}

	// Short circuit if possible (extra processing cannot be skipped)
	if vc.kongCPVersion == dataPlaneVersion && vc.extraProcessor == nil {
		return compressedPayload, nil
	}

	uncompressedPayloadBytes, err := UncompressPayload(compressedPayload)
	if err != nil {
		return nil, fmt.Errorf("unable to uncompress payload: %v", err)
	}
	// TODO(fero) perf use bytes
	uncompressedPayload := string(uncompressedPayloadBytes)

	processedPayload, err := vc.processConfigTableUpdates(
		uncompressedPayload, dataPlaneVersion)
	if err != nil {
		return nil, err
	}
	isEnterprise := strings.Contains(dataPlaneVersionStr, "enterprise")
	processedPayload, err = vc.performExtraProcessing(processedPayload,
		dataPlaneVersion, isEnterprise)
	if err != nil {
		return nil, err
	}

	return CompressPayload([]byte(processedPayload))
}

func (vc *WSVersionCompatibility) performExtraProcessing(uncompressedPayload string, dataPlaneVersion uint64,
	isEnterprise bool,
) (string, error) {
	if vc.extraProcessor != nil {
		processedPayload, err := vc.extraProcessor(uncompressedPayload, dataPlaneVersion,
			isEnterprise, vc.logger)
		if err != nil {
			return "", err
		}
		if !gjson.Valid(processedPayload) {
			return "", fmt.Errorf("processed payload is no longer valid JSON")
		}
		return processedPayload, nil
	}

	return uncompressedPayload, nil
}

func (vc *WSVersionCompatibility) getConfigTableUpdates(dataPlaneVersion uint64) []ConfigTableUpdates {
	configTableUpdates := []ConfigTableUpdates{}
	for versionNumber, updates := range vc.configTableUpdates {
		if dataPlaneVersion < versionNumber {
			configTableUpdates = append(configTableUpdates, updates...)
		}
	}
	return configTableUpdates
}

func (vc *WSVersionCompatibility) processConfigTableUpdates(uncompressedPayload string,
	dataPlaneVersion uint64,
) (string, error) {
	processedPayload := uncompressedPayload

	configTableUpdates := vc.getConfigTableUpdates(dataPlaneVersion)
	for _, configTableUpdate := range configTableUpdates {
		switch configTableUpdate.Type {
		case Plugin:
			processedPayload = vc.processPluginUpdates(processedPayload,
				configTableUpdate, dataPlaneVersion)
		case Service:
			processedPayload = vc.processCoreEntityUpdates(processedPayload,
				configTableUpdate, dataPlaneVersion)
		default:
			return "", fmt.Errorf("unsupported value type: %d", configTableUpdate.Type)
		}
	}

	if !gjson.Valid(processedPayload) {
		return "", fmt.Errorf("processed payload is no longer valid JSON")
	}

	return processedPayload, nil
}

func parseSemanticVersion(versionStr string) (uint64, error) {
	semVersion, err := kong.ParseSemanticVersion(versionStr)
	if err != nil {
		return 0, err
	}
	if semVersion.Minor >= invalidVersionOctet {
		return 0, fmt.Errorf("minor version must not be >= %d", invalidVersionOctet)
	}
	if semVersion.Patch >= invalidVersionOctet {
		return 0, fmt.Errorf("patch version must not be >= %d", invalidVersionOctet)
	}
	version := (majorVersionBase * semVersion.Major) +
		(minorVersionBase * semVersion.Minor) +
		(patchVersionBase * semVersion.Patch)

	if len(semVersion.Build) > 0 {
		buildVersion := semVersion.Build[0]
		if buildVersionRegex.MatchString(buildVersion) {
			tokens := buildVersionRegex.FindStringSubmatch(buildVersion)
			buildVersionStr := tokens[buildVersionRegex.SubexpIndex("build_version")]
			buildNum, err := strconv.ParseUint(buildVersionStr, base, bitSize)
			if err != nil {
				return 0, fmt.Errorf("unable to parse build version from version: %v", err)
			}
			if buildNum >= invalidVersionOctet {
				return 0, fmt.Errorf("build version must not be >= %d", invalidVersionOctet)
			}
			version += buildNum
		}
	}

	return version, nil
}

func (vc *WSVersionCompatibility) removePlugin(
	processedPayload string,
	pluginName string,
	dataPlaneVersion uint64,
) string {
	plugins := gjson.Get(processedPayload, "config_table.plugins")
	if plugins.IsArray() {
		removeCount := 0
		for i, res := range plugins.Array() {
			pluginCondition := fmt.Sprintf("..#(name=%s)", pluginName)
			if gjson.Get(res.Raw, pluginCondition).Exists() {
				var err error
				pluginDelete := fmt.Sprintf("config_table.plugins.%d", i-removeCount)
				if processedPayload, err = sjson.Delete(processedPayload, pluginDelete); err != nil {
					vc.logger.With(zap.String("plugin", pluginName)).
						With(zap.Uint64("data-plane", dataPlaneVersion)).
						With(zap.Error(err)).
						Error("plugin was not removed from configuration")
				} else {
					vc.logger.With(zap.String("plugin", pluginName)).
						With(zap.Uint64("data-plane", dataPlaneVersion)).
						Warn("removing plugin which is incompatible with data plane")
					removeCount++
				}
			}
		}
	}
	return processedPayload
}

func (vc *WSVersionCompatibility) processPluginUpdates(payload string,
	configTableUpdate ConfigTableUpdates,
	dataPlaneVersion uint64,
) string {
	pluginName := configTableUpdate.Name
	processedPayload := payload
	results := gjson.Get(processedPayload, fmt.Sprintf("config_table.plugins.#(name=%s)#", pluginName))
	if len(results.Indexes) > 0 {
		indexUpdate := 0
		for _, res := range results.Array() {
			updatedRaw := res.Raw
			var err error

			// Field removal
			for _, field := range configTableUpdate.RemoveFields {
				configField := fmt.Sprintf("config.%s", field)
				if gjson.Get(updatedRaw, configField).Exists() {
					if updatedRaw, err = sjson.Delete(updatedRaw, configField); err != nil {
						vc.logger.With(zap.String("plugin", pluginName)).
							With(zap.String("field", configField)).
							With(zap.Uint64("data-plane", dataPlaneVersion)).
							With(zap.Error(err)).
							Error("plugin configuration field was not removed from configuration")
					} else {
						vc.logger.With(zap.String("plugin", pluginName)).
							With(zap.String("field", configField)).
							With(zap.Uint64("data-plane", dataPlaneVersion)).
							Warn("removing plugin configuration field which is incompatible with data plane")
					}
				}
			}

			// Field element array removal
			for _, array := range configTableUpdate.RemoveElementsFromArray {
				configField := fmt.Sprintf("config.%s", array.Field)
				fieldArray := gjson.Get(updatedRaw, configField)
				if len(results.Indexes) > 0 {
					// Gather indexes to remove from array
					var arrayRemovalIndexes []int
					for i, arrayRes := range fieldArray.Array() {
						conditionField := fmt.Sprintf("..#(%s)", array.Condition)
						if gjson.Get(arrayRes.Raw, conditionField).Exists() {
							arrayRemovalIndexes = append(arrayRemovalIndexes, i)
						}
					}

					for i, arrayIndex := range arrayRemovalIndexes {
						fieldArrayWithIndex := fmt.Sprintf("config.%s.%d", array.Field, arrayIndex-i)
						var err error
						if updatedRaw, err = sjson.Delete(updatedRaw, fieldArrayWithIndex); err != nil {
							vc.logger.With(zap.String("plugin", pluginName)).
								With(zap.String("field", configField)).
								With(zap.String("condition", array.Condition)).
								With(zap.Int("index", arrayIndex)).
								With(zap.Uint64("data-plane", dataPlaneVersion)).
								With(zap.Error(err)).
								Error("plugin configuration array item was not removed from configuration")
						} else {
							vc.logger.With(zap.String("plugin", pluginName)).
								With(zap.String("field", configField)).
								With(zap.String("condition", array.Condition)).
								With(zap.Int("index", arrayIndex)).
								With(zap.Uint64("data-plane", dataPlaneVersion)).
								Warn("removing plugin configuration array item which is incompatible with data plane")
						}
					}
				}
			}

			// Field updates
			for _, update := range configTableUpdate.FieldUpdates {
				configField := fmt.Sprintf("config.%s", update.Field)
				if gjson.Get(updatedRaw, configField).Exists() {
					conditionField := fmt.Sprintf("[@this].#(config.%s)", update.Condition)
					if gjson.Get(updatedRaw, conditionField).Exists() {
						for _, fieldUpdate := range update.Updates {
							conditionUpdate := fmt.Sprintf("config.%s", fieldUpdate.Field)
							if fieldUpdate.Value == nil && len(fieldUpdate.ValueFromField) == 0 {
								// Handle field removal
								if updatedRaw, err = sjson.Delete(updatedRaw, conditionUpdate); err != nil {
									vc.logger.With(zap.String("plugin", pluginName)).
										With(zap.String("field", conditionUpdate)).
										With(zap.Uint64("data-plane", dataPlaneVersion)).
										With(zap.Error(err)).
										Error("plugin configuration item was not removed from configuration")
								} else {
									vc.logger.With(zap.String("plugin", pluginName)).
										With(zap.String("field", configField)).
										With(zap.String("condition", conditionUpdate)).
										With(zap.Uint64("data-plane", dataPlaneVersion)).
										Warn("removing plugin configuration item which is incompatible with data plane")
								}
							} else {
								// Get the field value if "Value" is a field
								var value interface{}
								if fieldUpdate.Value != nil {
									value = fieldUpdate.Value
								} else {
									valueFromField := fmt.Sprintf("config.%v", fieldUpdate.ValueFromField)
									res := gjson.Get(updatedRaw, valueFromField)
									if res.Exists() {
										value = res.Value()
									} else {
										vc.logger.With(zap.String("plugin", pluginName)).
											With(zap.String("field", configField)).
											With(zap.String("condition", update.Condition)).
											With(zap.Any("new-value", fieldUpdate.Value)).
											With(zap.Uint64("data-plane", dataPlaneVersion)).
											With(zap.Error(err)).
											Error("plugin configuration does not contain field value")
										break
									}
								}

								// Handle field update from value of value of field
								if updatedRaw, err = sjson.Set(updatedRaw, conditionUpdate, value); err != nil {
									vc.logger.With(zap.String("plugin", pluginName)).
										With(zap.String("field", configField)).
										With(zap.String("condition", update.Condition)).
										With(zap.Any("new-value", fieldUpdate.Value)).
										With(zap.Uint64("data-plane", dataPlaneVersion)).
										With(zap.Error(err)).
										Error("plugin configuration field was not updated int configuration")
								} else {
									vc.logger.With(zap.String("plugin", pluginName)).
										With(zap.String("field", configField)).
										With(zap.String("condition", update.Condition)).
										With(zap.Any("new-value", fieldUpdate.Value)).
										With(zap.Uint64("data-plane", dataPlaneVersion)).
										Warn("updating plugin configuration field which is incompatible with data plane")
								}
							}
						}
					}
				}
			}

			// Update the processed payload
			resIndex := res.Index - indexUpdate
			updatedPayload := processedPayload[:resIndex] + updatedRaw +
				processedPayload[resIndex+len(res.Raw):]
			indexUpdate = len(processedPayload) - len(updatedPayload)
			processedPayload = updatedPayload
		}
	}

	if configTableUpdate.Remove && results.Exists() {
		processedPayload = vc.removePlugin(processedPayload, pluginName,
			dataPlaneVersion)
	}
	return processedPayload
}

func (vc *WSVersionCompatibility) processCoreEntityUpdates(payload string,
	configTableUpdate ConfigTableUpdates,
	dataPlaneVersion uint64,
) string {
	entityName := configTableUpdate.Type.String()
	processedPayload := payload
	results := gjson.Get(processedPayload, fmt.Sprintf("config_table.%s", entityName))
	var (
		updates = []interface{}{}
		err     error
	)
	for _, res := range results.Array() {
		var entityJSON map[string]interface{}
		updatedRaw := res.Raw
		name := res.Get("name").Raw

		// Field removal
		for _, field := range configTableUpdate.RemoveFields {
			if gjson.Get(updatedRaw, field).Exists() {
				if updatedRaw, err = sjson.Delete(updatedRaw, field); err != nil {
					vc.logger.With(zap.String("entity", entityName)).
						With(zap.String("name", name)).
						With(zap.String("field", field)).
						With(zap.Uint64("data-plane", dataPlaneVersion)).
						With(zap.Error(err)).
						Error("entity field was not removed from configuration")
				} else {
					vc.logger.With(zap.String("entity", entityName)).
						With(zap.String("name", name)).
						With(zap.String("field", field)).
						With(zap.Uint64("data-plane", dataPlaneVersion)).
						Warn("removing entity field which is incompatible with data plane")
				}
			}
		}
		if err = json.Unmarshal([]byte(updatedRaw), &entityJSON); err != nil {
			vc.logger.With(zap.String("entity", entityName)).
				With(zap.String("name", name)).
				With(zap.String("config", updatedRaw)).
				With(zap.Uint64("data-plane", dataPlaneVersion)).
				With(zap.Error(err)).
				Error("couldn't unmarshal entity config")
		} else {
			updates = append(updates, entityJSON)
		}
	}
	if processedPayload, err = sjson.Set(
		processedPayload, fmt.Sprintf("config_table.%s", entityName), updates,
	); err != nil {
		vc.logger.With(zap.String("entity", entityName)).
			With(zap.Uint64("data-plane", dataPlaneVersion)).
			With(zap.Error(err)).
			Error("error while updating entities")
	}
	return processedPayload
}
