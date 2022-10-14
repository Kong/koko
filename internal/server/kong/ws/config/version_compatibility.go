package config

import (
	"fmt"
	"strings"

	"github.com/kong/koko/internal/json"
	_ "github.com/kong/koko/internal/resource" // ensure resources registered
	"github.com/kong/koko/internal/versioning"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"go.uber.org/zap"
)

const KongGatewayCompatibilityVersion = "3.0.0"

type VersionedConfigUpdates map[string][]ConfigTableUpdates

type VersionCompatibility interface {
	AddConfigTableUpdates(configTableUpdates VersionedConfigUpdates) error
	ProcessConfigTableUpdates(dataPlaneVersionStr string, compressedPayload []byte) ([]byte, TrackedChanges, error)
}

type VersionCompatibilityOpts struct {
	Logger         *zap.Logger
	KongCPVersion  string
	ExtraProcessor Processor
}

type UpdateType uint8

const (
	// Plugin is the UpdateType referring to plugins' config schema updates
	// e.g.: `$.config_table.plugins[?(@.name == '(PLUGIN_NAME)')].config`.
	Plugin UpdateType = iota
	// CorePlugin is the UpdateType referring to plugins's core schema updates
	// e.g.: `$.config_table.plugins[?(@.name == '(PLUGIN_NAME)')]`.
	CorePlugin

	Service
	Route
	Upstream
)

func (u UpdateType) String() string {
	return [...]string{"plugin", "plugin", "service", "route", "upstream"}[u]
}

func (u UpdateType) ConfigTableKey() string {
	return [...]string{"plugins", "plugins", "services", "routes", "upstreams"}[u]
}

//nolint:revive
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
	// FieldMustBeEmpty is a flag to indicate that the field update must only occur
	// if the Field is considered empty. If the Field is not empty the update is skipped.
	FieldMustBeEmpty bool
}

//nolint:revive
type ConfigTableFieldCondition struct {
	// Field is a top-level or nested field; use dot notation for nested fields.
	Field string
	// Condition is an expression for matching criteria.
	// uses gjson path syntax; https://github.com/tidwall/gjson#path-syntax
	Condition string
	// Updates is an array of updates to perform based on the matched criteria.
	Updates []ConfigTableFieldUpdate
}

//nolint:revive
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

	// ChangeID is a unique identifier for every schema update.
	ChangeID ChangeID

	// DisableChangeTracking takes in JSON encoded string representation of
	// UpdateType denoted by Type defined above.
	// This callback gives an opportunity to the change to dynamically disable
	// change tracking.
	//
	// An example use-case is when a newly added field with a backwards-compatible
	// default value is dropped from the configuration. In this case, the user
	// doesn't need to be notified of a benign change.
	// When unspecified, the change is always emitted.
	DisableChangeTracking func(rawJSON string) bool
}

type Processor func(uncompressedPayload string, dataPlaneVersion versioning.Version,
	tracker *ChangeTracker, logger *zap.Logger) (string, error)

type WSVersionCompatibility struct {
	logger             *zap.Logger
	kongCPVersion      string
	configTableUpdates map[string][]ConfigTableUpdates
	extraProcessor     Processor
}

func NewVersionCompatibilityProcessor(opts VersionCompatibilityOpts) (*WSVersionCompatibility, error) {
	if opts.Logger == nil {
		return nil, fmt.Errorf("opts.Logger required")
	}
	if len(strings.TrimSpace(opts.KongCPVersion)) == 0 {
		return nil, fmt.Errorf("opts.KongCPVersion required")
	}
	if _, err := versioning.NewVersion(opts.KongCPVersion); err != nil {
		return nil, fmt.Errorf("unable to parse opts.KongCPVersion %v", err)
	}

	return &WSVersionCompatibility{
		logger:             opts.Logger,
		kongCPVersion:      opts.KongCPVersion,
		configTableUpdates: make(map[string][]ConfigTableUpdates),
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

			if update.ChangeID == "" {
				return fmt.Errorf("invalid update with no change ID")
			}
		}
		vc.configTableUpdates[version] = append(vc.configTableUpdates[version], updates...)
	}
	return nil
}

func (vc *WSVersionCompatibility) ProcessConfigTableUpdates(dataPlaneVersionStr string,
	compressedPayload []byte,
) ([]byte, TrackedChanges, error) {
	dataPlaneVersion, err := versioning.NewVersion(dataPlaneVersionStr)
	if err != nil {
		return nil, TrackedChanges{}, fmt.Errorf("unable to parse data plane version: %w", err)
	}

	tracker := NewChangeTracker()

	uncompressedPayloadBytes, err := UncompressPayload(compressedPayload)
	if err != nil {
		return nil, TrackedChanges{},
			fmt.Errorf("unable to uncompress payload: %w", err)
	}
	// TODO(fero) perf use bytes
	uncompressedPayload := string(uncompressedPayloadBytes)

	processedPayload, err := vc.processConfigTableUpdates(uncompressedPayload, dataPlaneVersion, tracker)
	if err != nil {
		return nil, TrackedChanges{}, err
	}
	processedPayload, err = vc.performExtraProcessing(processedPayload, dataPlaneVersion, tracker)
	if err != nil {
		return nil, TrackedChanges{}, err
	}

	compatibleCompressedPayload, err := CompressPayload([]byte(processedPayload))
	if err != nil {
		return nil, TrackedChanges{}, err
	}
	return compatibleCompressedPayload, tracker.Get(), nil
}

func (vc *WSVersionCompatibility) performExtraProcessing(uncompressedPayload string,
	dataPlaneVersion versioning.Version,
	tracker *ChangeTracker,
) (string, error) {
	if vc.extraProcessor != nil {
		processedPayload, err := vc.extraProcessor(uncompressedPayload, dataPlaneVersion, tracker, vc.logger)
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

func (vc *WSVersionCompatibility) getConfigTableUpdates(dataPlaneVersion versioning.Version) []ConfigTableUpdates {
	configTableUpdates := []ConfigTableUpdates{}
	for versionRange, updates := range vc.configTableUpdates {
		versionRangeFunc := versioning.MustNewRange(versionRange)
		if versionRangeFunc(dataPlaneVersion) {
			configTableUpdates = append(configTableUpdates, updates...)
		}
	}
	return configTableUpdates
}

func (vc *WSVersionCompatibility) processConfigTableUpdates(uncompressedPayload string,
	dataPlaneVersion versioning.Version, tracker *ChangeTracker,
) (string, error) {
	dataPlaneVersionStr := dataPlaneVersion.String()
	processedPayload := uncompressedPayload
	configTableUpdates := vc.getConfigTableUpdates(dataPlaneVersion)
	for _, configTableUpdate := range configTableUpdates {
		switch configTableUpdate.Type {
		case Plugin:
			processedPayload = vc.processPluginUpdates(processedPayload,
				configTableUpdate, dataPlaneVersionStr, tracker)
		case Service, CorePlugin, Route, Upstream:
			processedPayload = vc.processCoreEntityUpdates(processedPayload,
				configTableUpdate, dataPlaneVersionStr, tracker)
		default:
			return "", fmt.Errorf("unsupported value type: %d", configTableUpdate.Type)
		}
	}

	if !gjson.Valid(processedPayload) {
		return "", fmt.Errorf("processed payload is no longer valid JSON")
	}

	return processedPayload, nil
}

func (vc *WSVersionCompatibility) removePlugin(
	processedPayload string,
	pluginName string,
	dataPlaneVersionStr string,
	changeID ChangeID,
	tracker *ChangeTracker,
) string {
	plugins := gjson.Get(processedPayload, "config_table.plugins")
	if plugins.IsArray() {
		removeCount := 0
		for i, res := range plugins.Array() {
			pluginCondition := fmt.Sprintf("..#(name=%s)", pluginName)
			if gjson.Get(res.Raw, pluginCondition).Exists() {
				var err error
				pluginID := res.Get("id").String()
				err = tracker.TrackForResource(changeID, ResourceInfo{
					Type: "plugin",
					ID:   pluginID,
				})
				if err != nil {
					vc.logger.Error("failed to track version compatibility change",
						zap.String("change-id", string(changeID)),
						zap.String("resource-type", "plugin"))
				}
				pluginDelete := fmt.Sprintf("config_table.plugins.%d", i-removeCount)
				if processedPayload, err = sjson.Delete(processedPayload, pluginDelete); err != nil {
					vc.logger.With(zap.String("plugin", pluginName)).
						With(zap.String("data-plane", dataPlaneVersionStr)).
						With(zap.Error(err)).
						Error("plugin was not removed from configuration")
				} else {
					vc.logger.With(zap.String("plugin", pluginName)).
						With(zap.String("data-plane", dataPlaneVersionStr)).
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
	dataPlaneVersionStr string, tracker *ChangeTracker,
) string {
	pluginName := configTableUpdate.Name
	processedPayload := payload
	results := gjson.Get(processedPayload, fmt.Sprintf("config_table.plugins.#(name=%s)#", pluginName))
	if len(results.Indexes) > 0 {
		indexUpdate := 0
		for _, res := range results.Array() {
			originalRaw := res.Raw
			updatedRaw := res.Raw
			var (
				err error
				// updated must be changed to true if there is a change to
				// configuration
				updated bool
			)

			// Field removal
			for _, field := range configTableUpdate.RemoveFields {
				configField := fmt.Sprintf("config.%s", field)
				if gjson.Get(updatedRaw, configField).Exists() {
					updated = true
					if updatedRaw, err = sjson.Delete(updatedRaw, configField); err != nil {
						vc.logger.With(zap.String("plugin", pluginName)).
							With(zap.String("field", configField)).
							With(zap.String("data-plane", dataPlaneVersionStr)).
							With(zap.Error(err)).
							Error("plugin configuration field was not removed from configuration")
					} else {
						vc.logger.With(zap.String("plugin", pluginName)).
							With(zap.String("field", configField)).
							With(zap.String("data-plane", dataPlaneVersionStr)).
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
						updated = true
						if updatedRaw, err = sjson.Delete(updatedRaw, fieldArrayWithIndex); err != nil {
							vc.logger.With(zap.String("plugin", pluginName)).
								With(zap.String("field", configField)).
								With(zap.String("condition", array.Condition)).
								With(zap.Int("index", arrayIndex)).
								With(zap.String("data-plane", dataPlaneVersionStr)).
								With(zap.Error(err)).
								Error("plugin configuration array item was not removed from configuration")
						} else {
							vc.logger.With(zap.String("plugin", pluginName)).
								With(zap.String("field", configField)).
								With(zap.String("condition", array.Condition)).
								With(zap.Int("index", arrayIndex)).
								With(zap.String("data-plane", dataPlaneVersionStr)).
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
							// Ensure the original field is not empty if specified; do not overwrite
							if fieldUpdate.FieldMustBeEmpty && !ValueIsEmpty(gjson.Get(updatedRaw, conditionUpdate)) {
								// Since this is a copy function from another field and the current field is already
								// configured the entire field update process should short circuit
								continue
							}

							if fieldUpdate.Value == nil && len(fieldUpdate.ValueFromField) == 0 {
								// Handle field removal
								updated = true
								if updatedRaw, err = sjson.Delete(updatedRaw, conditionUpdate); err != nil {
									vc.logger.With(zap.String("plugin", pluginName)).
										With(zap.String("field", conditionUpdate)).
										With(zap.String("data-plane", dataPlaneVersionStr)).
										With(zap.Error(err)).
										Error("plugin configuration item was not removed from configuration")
								} else {
									vc.logger.With(zap.String("plugin", pluginName)).
										With(zap.String("field", configField)).
										With(zap.String("condition", conditionUpdate)).
										With(zap.String("data-plane", dataPlaneVersionStr)).
										Warn("removing plugin configuration item which is incompatible with data plane")
								}
							} else {
								// Get the field value if "Value" is a field
								var value interface{}
								if fieldUpdate.Value != nil {
									value = fieldUpdate.Value
								} else {
									res := gjson.Get(updatedRaw, fmt.Sprintf("config.%v", fieldUpdate.ValueFromField))
									if !ValueIsEmpty(res) {
										value = res.Value()
									} else {
										vc.logger.With(zap.String("plugin", pluginName)).
											With(zap.String("field", configField)).
											With(zap.String("condition", update.Condition)).
											With(zap.Any("new-value", fieldUpdate.Value)).
											With(zap.String("data-plane", dataPlaneVersionStr)).
											With(zap.Error(err)).
											Error("plugin configuration does not contain field value")
										break
									}
								}

								// Handle field update from value of value of field
								updated = true
								if updatedRaw, err = sjson.Set(updatedRaw, conditionUpdate, value); err != nil {
									vc.logger.With(zap.String("plugin", pluginName)).
										With(zap.String("field", configField)).
										With(zap.String("condition", update.Condition)).
										With(zap.Any("new-value", fieldUpdate.Value)).
										With(zap.String("data-plane", dataPlaneVersionStr)).
										With(zap.Error(err)).
										Error("plugin configuration field was not updated in the configuration")
								} else {
									vc.logger.With(zap.String("plugin", pluginName)).
										With(zap.String("field", configField)).
										With(zap.String("condition", update.Condition)).
										With(zap.Any("new-value", fieldUpdate.Value)).
										With(zap.String("data-plane", dataPlaneVersionStr)).
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
			indexUpdate += len(processedPayload) - len(updatedPayload)
			processedPayload = updatedPayload

			if updated && shouldTrackChange(configTableUpdate, originalRaw) {
				pluginID := gjson.Get(updatedRaw, "id").String()
				err := tracker.TrackForResource(configTableUpdate.ChangeID, ResourceInfo{
					Type: "plugin",
					ID:   pluginID,
				})
				if err != nil {
					vc.logger.Error("failed to track version compatibility change",
						zap.String("change-id", string(configTableUpdate.ChangeID)),
						zap.String("resource-type", "plugin"))
				}
			}
		}
	}

	if configTableUpdate.Remove && results.Exists() {
		processedPayload = vc.removePlugin(processedPayload, pluginName,
			dataPlaneVersionStr, configTableUpdate.ChangeID, tracker)
	}
	return processedPayload
}

func (vc *WSVersionCompatibility) processCoreEntityUpdates(payload string,
	configTableUpdate ConfigTableUpdates,
	dataPlaneVersionStr string, tracker *ChangeTracker,
) string {
	entityType := configTableUpdate.Type.String()
	configTableKey := configTableUpdate.Type.ConfigTableKey()

	processedPayload := payload
	results := gjson.Get(processedPayload, fmt.Sprintf("config_table.%s", configTableKey))
	var (
		updates []interface{}
		err     error
	)
	if !results.Exists() {
		return processedPayload
	}
	for _, res := range results.Array() {
		var (
			entityJSON  map[string]interface{}
			originalRaw = res.Raw
			updatedRaw  = res.Raw
			name        = res.Get("name").Raw
			updated     bool
		)

		// Field removal
		for _, field := range configTableUpdate.RemoveFields {
			if gjson.Get(updatedRaw, field).Exists() {
				updated = true
				if updatedRaw, err = sjson.Delete(updatedRaw, field); err != nil {
					vc.logger.With(zap.String("entity", entityType)).
						With(zap.String("name", name)).
						With(zap.String("field", field)).
						With(zap.String("data-plane", dataPlaneVersionStr)).
						With(zap.Error(err)).
						Error("entity field was not removed from configuration")
				} else {
					vc.logger.With(zap.String("entity", entityType)).
						With(zap.String("name", name)).
						With(zap.String("field", field)).
						With(zap.String("data-plane", dataPlaneVersionStr)).
						Warn("removing entity field which is incompatible with data plane")
				}
			}
		}

		// Field element array removal
		for _, array := range configTableUpdate.RemoveElementsFromArray {
			configField := array.Field
			fieldArray := gjson.Get(updatedRaw, configField)
			// Gather indexes to remove from array
			var arrayRemovalIndexes []int
			for i, arrayRes := range fieldArray.Array() {
				conditionField := fmt.Sprintf("..#(%s)", array.Condition)
				if gjson.Get(arrayRes.Raw, conditionField).Exists() {
					arrayRemovalIndexes = append(arrayRemovalIndexes, i)
				}
			}

			for i, arrayIndex := range arrayRemovalIndexes {
				fieldArrayWithIndex := fmt.Sprintf("%s.%d", array.Field, arrayIndex-i)
				var err error
				updated = true
				if updatedRaw, err = sjson.Delete(updatedRaw, fieldArrayWithIndex); err != nil {
					vc.logger.With(zap.String("entity", entityType)).
						With(zap.String("name", name)).
						With(zap.String("field", configField)).
						With(zap.String("condition", array.Condition)).
						With(zap.Int("index", arrayIndex)).
						With(zap.String("data-plane", dataPlaneVersionStr)).
						With(zap.Error(err)).
						Error("plugin configuration array item was not removed from configuration")
				} else {
					vc.logger.With(zap.String("entity", entityType)).
						With(zap.String("name", name)).
						With(zap.String("field", configField)).
						With(zap.String("condition", array.Condition)).
						With(zap.Int("index", arrayIndex)).
						With(zap.String("data-plane", dataPlaneVersionStr)).
						Warn("removing plugin configuration array item which is incompatible with data plane")
				}
			}
		}

		// Field update
		for _, update := range configTableUpdate.FieldUpdates {
			fmt.Println("update: ", update)
			configField := update.Field
			if gjson.Get(updatedRaw, configField).Exists() {
				conditionField := fmt.Sprintf("[@this].#(%s)", update.Condition)
				if gjson.Get(updatedRaw, conditionField).Exists() {
					for _, fieldUpdate := range update.Updates {
						conditionUpdate := fieldUpdate.Field
						// Ensure the original field is not empty if specified; do not overwrite
						if fieldUpdate.FieldMustBeEmpty && !ValueIsEmpty(gjson.Get(updatedRaw, conditionUpdate)) {
							continue
						}
						if fieldUpdate.Value == nil && len(fieldUpdate.ValueFromField) == 0 {
							// Handle field removal
							updated = true
							if updatedRaw, err = sjson.Delete(updatedRaw, conditionUpdate); err != nil {
								vc.logger.With(zap.String("entity", entityType)).
									With(zap.String("name", name)).
									With(zap.String("field", conditionUpdate)).
									With(zap.String("data-plane", dataPlaneVersionStr)).
									With(zap.Error(err)).
									Error("entity item was not removed from configuration")
							} else {
								vc.logger.With(zap.String("entity", entityType)).
									With(zap.String("name", name)).
									With(zap.String("field", configField)).
									With(zap.String("condition", conditionUpdate)).
									With(zap.String("data-plane", dataPlaneVersionStr)).
									Warn("removing entity item which is incompatible with data plane")
							}
						} else {
							// Get the field value if "Value" is a field
							var value interface{}
							if fieldUpdate.Value != nil {
								value = fieldUpdate.Value
							} else {
								res := gjson.Get(updatedRaw, fieldUpdate.ValueFromField)
								if !ValueIsEmpty(res) {
									value = res.Value()
								} else {
									vc.logger.With(zap.String("entity", entityType)).
										With(zap.String("name", name)).
										With(zap.String("field", configField)).
										With(zap.String("condition", update.Condition)).
										With(zap.Any("new-value", fieldUpdate.Value)).
										With(zap.String("data-plane", dataPlaneVersionStr)).
										With(zap.Error(err)).
										Error("entity does not contain field value")
									break
								}
							}

							// Handle field update from value of value of field
							updated = true
							if updatedRaw, err = sjson.Set(updatedRaw, conditionUpdate, value); err != nil {
								vc.logger.With(zap.String("entity", entityType)).
									With(zap.String("name", name)).
									With(zap.String("field", configField)).
									With(zap.String("condition", update.Condition)).
									With(zap.Any("new-value", fieldUpdate.Value)).
									With(zap.String("data-plane", dataPlaneVersionStr)).
									With(zap.Error(err)).
									Error("entity field was not updated int configuration")
							} else {
								vc.logger.With(zap.String("entity", entityType)).
									With(zap.String("name", name)).
									With(zap.String("field", configField)).
									With(zap.String("condition", update.Condition)).
									With(zap.Any("new-value", fieldUpdate.Value)).
									With(zap.String("data-plane", dataPlaneVersionStr)).
									Warn("updating entity field which is incompatible with data plane")
							}
						}
					}
				}
			}
		}

		if err = json.Unmarshal([]byte(updatedRaw), &entityJSON); err != nil {
			vc.logger.With(zap.String("entity", entityType)).
				With(zap.String("name", name)).
				With(zap.String("config", updatedRaw)).
				With(zap.String("data-plane", dataPlaneVersionStr)).
				With(zap.Error(err)).
				Error("couldn't unmarshal entity config")
		} else {
			updates = append(updates, entityJSON)
		}

		if updated && shouldTrackChange(configTableUpdate, originalRaw) {
			entityID := gjson.Get(updatedRaw, "id").String()
			err := tracker.TrackForResource(configTableUpdate.ChangeID, ResourceInfo{
				Type: entityType,
				ID:   entityID,
			})
			if err != nil {
				vc.logger.Error("failed to track version compatibility change",
					zap.String("change-id", string(configTableUpdate.ChangeID)),
					zap.String("resource-type", entityType))
			}
		}
	}
	if processedPayload, err = sjson.Set(
		processedPayload, fmt.Sprintf("config_table.%s", configTableKey), updates,
	); err != nil {
		vc.logger.With(zap.String("entity", entityType)).
			With(zap.String("data-plane", dataPlaneVersionStr)).
			With(zap.Error(err)).
			Error("error while updating entities")
	}
	return processedPayload
}

func shouldTrackChange(updates ConfigTableUpdates, entityJSON string) bool {
	if updates.DisableChangeTracking == nil {
		return true
	}
	return !updates.DisableChangeTracking(entityJSON)
}

// ValueIsEmpty returns true to indicate a given gjson Result is considered
// "empty". Empty for a given type is when the type is:
//   - any null JSON value
//   - an object containing no items
//   - an array of zero length
//   - an empty string
func ValueIsEmpty(value gjson.Result) bool {
	if value.Type == gjson.Null {
		return true
	}

	if value.IsObject() {
		return len(value.Map()) == 0
	}

	if value.IsArray() {
		return len(value.Array()) == 0
	}

	if value.String() == "" {
		return true
	}
	return false
}
