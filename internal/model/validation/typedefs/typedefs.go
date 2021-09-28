package typedefs

import (
	"fmt"
	"regexp"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

func IDRules() []validation.Rule {
	return []validation.Rule{validation.Required, UUID()}
}

func UUID() validation.Rule {
	return is.UUID
}

func Name() []validation.Rule {
	return []validation.Rule{
		REMatch(`^[0-9a-zA-z\.\-\_\~]*$`, false),
	}
}

const (
	ProtocolHTTP  = "http"
	ProtocolHTTPS = "https"
	ProtocolGRPC  = "grpc"
	ProtocolGRPCS = "grpcs"
	ProtocolUDP   = "udp"
	ProtocolTCP   = "tcp"
)

var protocols = []string{
	ProtocolHTTP, ProtocolHTTPS,
	ProtocolTCP, ProtocolUDP,
	ProtocolGRPC, ProtocolGRPCS,
}

func Enum(values ...string) validation.Rule {
	err := validation.NewError("invalid_enum",
		fmt.Sprintf("value must be one of %v", values))
	return validation.NewStringRuleWithError(func(s string) bool {
		for _, value := range values {
			if s == value {
				return true
			}
		}
		return false
	}, err)
}

func Protocol() []validation.Rule {
	return []validation.Rule{Enum(protocols...)}
}

func Timestamp() []validation.Rule {
	return []validation.Rule{
		validation.Empty,
	}
}

func Host() []validation.Rule {
	return []validation.Rule{is.Host}
}

const (
	maxPort = (1 << 16) - 1
)

func Port() []validation.Rule {
	return []validation.Rule{
		validation.Min(1),
		validation.Max(maxPort),
	}
}

func Path() []validation.Rule {
	return []validation.Rule{
		StringPrefix("/"),
		REMatch("//", true),
	}
}

func StringPrefix(prefix string) validation.Rule {
	err := validation.NewError("string_prefix", fmt.Sprintf(
		"must be prefixed with '%v'", prefix,
	))
	return validation.NewStringRuleWithError(func(s string) bool {
		return strings.HasPrefix(s, prefix)
	}, err)
}

func REMatch(regex string, invert bool) validation.Rule {
	err := validation.NewError("regex_match",
		fmt.Sprintf("must match regex; %v", regex))
	re := regexp.MustCompile(regex)
	return validation.NewStringRuleWithError(func(s string) bool {
		matched := re.MatchString(s)
		if invert {
			return !matched
		}
		return matched
	}, err)
}

const (
	maxTimeout = (1 << 31) - 2
)

func Timeout() []validation.Rule {
	return []validation.Rule{
		validation.Min(1),
		validation.Max(maxTimeout),
	}
}

func Tags() []validation.Rule {
	return []validation.Rule{
		validation.Each(Name()...),
		validation.Length(0, 8),
	}
}
