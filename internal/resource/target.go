package resource

import (
	"context"
	"encoding/hex"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"

	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	"github.com/kong/koko/internal/model"
	"github.com/kong/koko/internal/model/json/extension"
	"github.com/kong/koko/internal/model/json/generator"
	"github.com/kong/koko/internal/model/json/validation"
	"github.com/kong/koko/internal/model/json/validation/typedefs"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

const (
	// TypeTarget denotes the Target type.
	TypeTarget model.Type = "target"

	defaultTargetPort = 8000
	maxPortNumber     = 65535
)

var (
	IPv4LikePattern       = regexp.MustCompile(`^[0-9.]+(/\d+)?$`)
	HostnamePattern       = regexp.MustCompile(typedefs.HostnamePattern)
	IPv6HasPortPattern    = regexp.MustCompile(`\]\:\d+$`)
	IPv6HasBracketPattern = regexp.MustCompile(`\[\S+\]$`)
)

var _ model.Object = Target{}

var (
	maxWeight           = 65535
	defaultWeight int32 = 100
)

type hostnameType int

const (
	typeName = iota
	typeIPv4
	typeIPv6
)

func NewTarget() Target {
	return Target{
		Target: &v1.Target{},
	}
}

type Target struct {
	Target *v1.Target
}

func (t Target) ID() string {
	if t.Target == nil {
		return ""
	}
	return t.Target.Id
}

func (t Target) Type() model.Type {
	return TypeTarget
}

func (t Target) Resource() model.Resource {
	return t.Target
}

// SetResource implements the Object.SetResource interface.
func (t Target) SetResource(r model.Resource) error { return model.SetResource(t, r) }

func (t Target) Indexes() []model.Index {
	res := []model.Index{
		{
			Name:      "target",
			Type:      model.IndexUnique,
			Value:     model.MultiValueIndex(t.Target.Upstream.Id, t.Target.Target),
			FieldName: "target",
		},
		{
			Name:        "upstream_id",
			Type:        model.IndexForeign,
			ForeignType: TypeUpstream,
			FieldName:   "upstream.id",
			Value:       t.Target.Upstream.Id,
		},
	}
	return res
}

func (t Target) Validate(ctx context.Context) error {
	err := validation.Validate(string(TypeTarget), t.Target)
	if err != nil {
		return err
	}
	_, err = validateAndFormatTarget(t.Target.Target)
	if err != nil {
		errWrap := validation.Error{}
		errWrap.Errs = append(errWrap.Errs, &v1.ErrorDetail{
			Type:  v1.ErrorType_ERROR_TYPE_FIELD,
			Field: "target",
			Messages: []string{
				fmt.Sprintf("not a valid hostname or ip address: '%s'", t.Target.Target),
			},
		})
		return errWrap
	}
	return nil
}

func (t Target) ProcessDefaults(ctx context.Context) error {
	if t.Target == nil {
		return fmt.Errorf("invalid nil resource")
	}
	defaultID(&t.Target.Id)
	if t.Target.Weight == nil {
		t.Target.Weight = wrapperspb.Int32(defaultWeight)
	}
	var err error
	target, err := validateAndFormatTarget(t.Target.Target)
	if err != nil {
		errWrap := validation.Error{}
		errWrap.Errs = append(errWrap.Errs, &v1.ErrorDetail{
			Type:  v1.ErrorType_ERROR_TYPE_FIELD,
			Field: "target",
			Messages: []string{
				fmt.Sprintf("not a valid hostname or ip address: '%s'", t.Target.Target),
			},
		})
		return errWrap
	}
	t.Target.Target = target
	return nil
}

func init() {
	err := model.RegisterType(TypeTarget, &v1.Target{}, func() model.Object {
		return NewTarget()
	})
	if err != nil {
		panic(err)
	}

	zero := 0
	targetSchema := &generator.Schema{
		Type: "object",
		Properties: map[string]*generator.Schema{
			"id": typedefs.ID,
			"target": {
				Type:      "string",
				MinLength: 1,
				MaxLength: maxHostnameLength,
			},
			"weight": {
				Type:    "integer",
				Minimum: &zero,
				Maximum: maxWeight,
				Default: defaultWeight,
			},
			"tags":       typedefs.Tags,
			"created_at": typedefs.UnixEpoch,
			"updated_at": typedefs.UnixEpoch,
			"upstream":   typedefs.ReferenceObject,
		},
		AdditionalProperties: &falsy,
		Required: []string{
			"id",
			"target",
			"upstream",
		},
		XKokoConfig: &extension.Config{
			ResourceAPIPath: "targets",
		},
	}
	err = generator.DefaultRegistry.Register(string(TypeTarget), targetSchema)
	if err != nil {
		panic(err)
	}
}

// hostnameType checks what kind of format the target has,
// without doing any further validation.
//
// Does it includes multiple ':'           -> it's an ipv6
// Is it a dot-separated string of numbers -> it's an ipv4
// Otherwise                               -> it's a domain name.
func hostnameCheck(hostname string) hostnameType {
	parts := strings.Split(hostname, ":")
	if len(parts) > 2 { //nolint:gomnd
		return typeIPv6
	}
	if IPv4LikePattern.FindString(parts[0]) != "" {
		return typeIPv4
	}
	return typeName
}

func validatePort(portStr string) (int, error) {
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return port, fmt.Errorf("invalid port: %s", portStr)
	}
	if port > maxPortNumber || port < 0 {
		return port, fmt.Errorf("invalid port number: %d", port)
	}
	return port, nil
}

// normalizeTargetHostname returns a normalized target name
// in the format 'name:port' if input is a valid name.
func normalizeTargetHostname(target string) (string, error) {
	var err error

	port := defaultTargetPort
	parts := strings.Split(target, ":")
	domain := parts[0]
	if len(parts) > 2 { //nolint:gomnd
		// multiple colons are not allowed for a domain name.
		return "", fmt.Errorf("invalid target name: %s", target)
	}
	if len(parts) == 2 { //nolint:gomnd
		if port, err = validatePort(parts[1]); err != nil {
			return "", err
		}
	}
	if HostnamePattern.FindString(domain) == "" {
		return "", fmt.Errorf("invalid hostname: %s", domain)
	}
	return fmt.Sprintf("%s:%d", domain, port), nil
}

// normalizeIPv4 returns a normalized ipv4
// in the format 'address:port' if input is a valid ipv4.
func normalizeIPv4(target string) (string, error) {
	var err error
	port := defaultTargetPort
	ip := target
	if strings.Contains(target, ":") {
		// has a port
		var portStr string
		parts := strings.Split(target, ":")
		ip, portStr = parts[0], parts[1]
		port, err = validatePort(portStr)
		if err != nil {
			return "", err
		}
	}
	if net.ParseIP(ip).To4() == nil {
		return "", fmt.Errorf("invalid ipv4 address %s", ip)
	}
	return fmt.Sprintf("%s:%d", ip, port), nil
}

// expandIPv6 decompress an ipv6 address into its 'long' format.
// for example:
//
// from ::1 to 0000:0000:0000:0000:0000:0000:0000:0001.
func expandIPv6(address string) string {
	ip := net.ParseIP(address).To16()
	dst := make([]byte, hex.EncodedLen(len(ip)))
	hex.Encode(dst, ip)
	var final string
	for i := 0; i < len(dst); i += 4 {
		final += fmt.Sprintf("%s:", dst[i:i+4])
	}
	// remove last colon
	return final[:len(final)-1]
}

func removeBrackets(ip string) string {
	ip = strings.ReplaceAll(ip, "[", "")
	return strings.ReplaceAll(ip, "]", "")
}

// normalizeIPv6 returns a normalized ipv6
// in the format '[address]:port' if input is a valid ipv6.
func normalizeIPv6(target string) (string, error) {
	var err error
	ip := target
	port := defaultTargetPort
	match := IPv6HasPortPattern.FindStringSubmatch(target)
	if len(match) > 0 {
		// has [address]:port pattern
		portString := strings.ReplaceAll(match[0], "]:", "")
		port, err = validatePort(portString)
		if err != nil {
			return "", err
		}
		ip = strings.ReplaceAll(target, match[0], "")
		ip = removeBrackets(ip)
	} else {
		match = IPv6HasBracketPattern.FindStringSubmatch(target)
		if len(match) > 0 {
			ip = removeBrackets(match[0])
		}
		if net.ParseIP(ip).To16() == nil {
			return "", fmt.Errorf("invalid ipv6 address %s", target)
		}
	}
	return fmt.Sprintf("[%s]:%d", expandIPv6(ip), port), nil
}

func validateAndFormatTarget(target string) (string, error) {
	if target == "" {
		return target, nil
	}
	var err error
	newTarget := target
	targetType := hostnameCheck(target)
	switch targetType {
	case typeName:
		newTarget, err = normalizeTargetHostname(target)
	case typeIPv4:
		newTarget, err = normalizeIPv4(target)
	case typeIPv6:
		newTarget, err = normalizeIPv6(target)
	}
	return newTarget, err
}
