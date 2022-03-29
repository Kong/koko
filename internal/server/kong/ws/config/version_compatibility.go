package config

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/kong/go-kong/kong"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"go.uber.org/zap"
)

const (
	KongGatewayCompatibilityVersion = "2.8.0"

	buildVersionPattern = `(?P<build_version>^[0-9]+)[a-zA-Z\-]*`
	invalidVersionOctet = 1000
	majorVersionBase    = 1000000000
	minorVersionBase    = majorVersionBase / 1000
	patchVersionBase    = minorVersionBase / 1000
	base                = 10
	bitSize             = 64
)

var buildVersionRegex = regexp.MustCompile(buildVersionPattern)

type VersionCompatibility interface {
	AddConfigTableUpdates(configTableUpdates map[uint64][]ConfigTableUpdates) error
	ProcessConfigTableUpdates(dataPlaneVersionStr string, compressedPayload []byte) ([]byte, error)
}

type VersionCompatibilityOpts struct {
	Logger          *zap.Logger
	KongCPVersion   string
	ExtraProcessing func(uncompressedPayload string,
		dataPlaneVersion uint64, isEnterprise bool) (string, error)
}

type UpdateType uint8

const (
	Plugin UpdateType = 1
)

//nolint: revive
type ConfigTableUpdates struct {
	Name         string
	Type         UpdateType
	RemoveFields []string
}

type WSVersionCompatibility struct {
	logger             *zap.Logger
	kongCPVersion      uint64
	configTableUpdates map[uint64][]ConfigTableUpdates
	extraProcessing    func(uncompressedPayload string,
		dataPlaneVersion uint64, isEnterprise bool) (string, error)
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
		extraProcessing:    opts.ExtraProcessing,
	}, nil
}

func (vc *WSVersionCompatibility) AddConfigTableUpdates(pluginPayloadUpdates map[uint64][]ConfigTableUpdates) error {
	for version, pluginUpdates := range pluginPayloadUpdates {
		vc.configTableUpdates[version] = append(vc.configTableUpdates[version], pluginUpdates...)
	}
	return nil
}

func (vc *WSVersionCompatibility) ProcessConfigTableUpdates(dataPlaneVersionStr string,
	compressedPayload []byte,
) ([]byte, error) {
	isEnterprise := strings.Contains(dataPlaneVersionStr, "enterprise")
	dataPlaneVersion, err := parseSemanticVersion(dataPlaneVersionStr)
	if err != nil {
		return nil, fmt.Errorf("unable to parse data plane version: %v", err)
	}

	// Short circuit if possible (extra processing cannot be skipped)
	if vc.kongCPVersion == dataPlaneVersion && vc.extraProcessing == nil {
		return compressedPayload, nil
	}

	uncompressedPayloadBytes, err := UncompressPayload(compressedPayload)
	if err != nil {
		return nil, fmt.Errorf("unable to uncompress payload: %v", err)
	}
	// TODO(fero) perf use bytes
	uncompressedPayload := string(uncompressedPayloadBytes)

	processedPayload, err := vc.processConfigTableUpdates(uncompressedPayload, dataPlaneVersion)
	if err != nil {
		return nil, err
	}
	processedPayload, err = vc.performExtraProcessing(processedPayload, dataPlaneVersion, isEnterprise)
	if err != nil {
		return nil, err
	}

	return CompressPayload([]byte(processedPayload))
}

func (vc *WSVersionCompatibility) performExtraProcessing(uncompressedPayload string, dataPlaneVersion uint64,
	isEnterprise bool,
) (string, error) {
	if vc.extraProcessing != nil {
		processedPayload, err := vc.extraProcessing(uncompressedPayload, dataPlaneVersion, isEnterprise)
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
		if configTableUpdate.Type == Plugin {
			processedPayload = processPluginUpdates(processedPayload, configTableUpdate.Name,
				configTableUpdate.RemoveFields, vc.kongCPVersion, dataPlaneVersion,
				vc.logger)
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

func processPluginUpdates(payload string, name string, fields []string, controlPlaneVersion uint64,
	dataPlaneVersion uint64, logger *zap.Logger,
) string {
	processedPayload := payload
	results := gjson.Get(processedPayload, fmt.Sprintf("config_table.plugins.#(name=%s)#", name))
	if len(results.Indexes) > 0 {
		indexUpdate := 0
		for _, res := range results.Array() {
			for _, field := range fields {
				configField := fmt.Sprintf("config.%s", field)
				if gjson.Get(res.Raw, configField).Exists() {
					if updatedRaw, err := sjson.Delete(res.Raw, configField); err != nil {
						logger.With(zap.String("plugin", name)).
							With(zap.String("field", configField)).
							With(zap.Uint64("control-plane", controlPlaneVersion)).
							With(zap.Uint64("data-plane", dataPlaneVersion)).
							With(zap.Error(err)).
							Error("plugin configuration field was not removed from configuration")
					} else {
						logger.With(zap.String("plugin", name)).
							With(zap.String("field", configField)).
							With(zap.Uint64("control-plane", controlPlaneVersion)).
							With(zap.Uint64("data-plane", dataPlaneVersion)).
							Warn("removing plugin configuration field which is incompatible with data plane")
						resIndex := res.Index - indexUpdate
						updatedPayload := processedPayload[:resIndex] + updatedRaw +
							processedPayload[resIndex+len(res.Raw):]
						indexUpdate = len(processedPayload) - len(updatedPayload)
						processedPayload = updatedPayload
					}
				}
			}
		}
	}

	return processedPayload
}
