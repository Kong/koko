package config

import (
	"fmt"
	"os"
)

type Log struct {
	Level  string `yaml:"level" json:"level" env:"LEVEL" env-default:"info"`
	Format string `yaml:"format" json:"format" env:"FORMAT" env-default:"json"`
}

type AdminServer struct {
	Address string `yaml:"address" json:"address" env:"ADDRESS" env-default:":3000"`
}

type ControlServer struct {
	TLSCertPath string `yaml:"tls_cert_path" json:"tls_cert_path" env:"TLS_CERT_PATH"`
	TLSKeyPath  string `yaml:"tls_key_path" json:"tls_key_path" env:"TLS_KEY_PATH"`
}

// Metrics config.
type Metrics struct {
	Enabled bool `json:"enabled,omitempty"`
}

// TLS defines re-usable TLS configuration used in tls.Config.
//
// If `TLS.Enable` is true and all other fields are empty, peer certificate
// validation will be skipped and only hostname verification will be done.
//
// All `*File` fields will be resolved automatically and thrown into its
// relevant string field upon config initialization.
type TLS struct {
	Enable                   bool   `yaml:"enable" json:"enable" env:"ENABLE"`
	Certificate              string `yaml:"certificate" json:"certificate" env:"CERTIFICATE"`
	CertificateFile          string `yaml:"certificate_file" json:"certificate_file" env:"CERTIFICATE_FILE"`
	Key                      string `yaml:"key" json:"key" env:"KEY"`
	KeyFile                  string `yaml:"key_file" json:"key_file" env:"KEY_FILE"`
	RootCA                   string `yaml:"root_ca" json:"root_ca" env:"ROOT_CA"`
	RootCAFile               string `yaml:"root_ca_file" json:"root_ca_file" env:"ROOT_CA_FILE"`
	SkipHostnameVerification bool   `yaml:"skip_hostname_verification" json:"skip_hostname_verification" env:"SKIP_HOSTNAME_VERIFICATION"` //nolint:lll
}

// Config represents configuration of Koko.
// Configuration is populated via a JSON or YAML file containing or via
// environment variables.
// The precedence order from lowest to highest priority is:
// - defaults
// - values in the configuration file
// - values in environment variables
// Array/Slice types are not supported within this data-structure.
type Config struct {
	Log                     Log           `yaml:"log" json:"log" env-prefix:"KOKO_LOG_"`
	Admin                   AdminServer   `yaml:"admin_server" json:"admin_server" env-prefix:"KOKO_ADMIN_SERVER_"`
	Control                 ControlServer `yaml:"control_server" json:"control_server" env-prefix:"KOKO_CONTROL_SERVER_"`
	Database                Database      `yaml:"database" json:"database" env-prefix:"KOKO_DATABASE_"`
	Metrics                 Metrics       `yaml:"metrics" json:"metrics" env-prefix:"KOKO_METRICS_"`
	DisableAnonymousReports bool          `yaml:"disable_anonymous_reports" json:"disable_anonymous_reports" env-prefix:"KOKO_DISABLE_ANONYMOUS_REPORTS"` //nolint:lll
}

// fetchFileContents gathers file contents from the provided fields and sets it on the applicable string field.
func (t *TLS) fetchFileContents() error {
	type fieldData struct {
		name     string
		filename string
		ptr      *string
	}

	for _, d := range []fieldData{
		{"root CA", t.RootCAFile, &t.RootCA},
		{"certificate", t.CertificateFile, &t.Certificate},
		{"key", t.KeyFile, &t.Key},
	} {
		// No filename set for the field, so we can ignore.
		if d.filename == "" {
			continue
		}

		if *d.ptr != "" {
			return fmt.Errorf("cannot set both %s file & raw string", d.name)
		}

		b, err := os.ReadFile(d.filename)
		if err != nil {
			return fmt.Errorf("unable to read %s file contents: %w", d.name, err)
		}

		*d.ptr = string(b)
	}

	return nil
}
