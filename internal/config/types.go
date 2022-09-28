package config

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
	// ClientType metrics client type e.g. prometheus, datadog.
	ClientType string `yaml:"client_type" json:"client_type" env:"CLIENT_TYPE" env-default:"noop"`
}

// TLS defines re-usable TLS configuration used in tls.Config.
type TLS struct {
	Enable                bool   `yaml:"enable" json:"enable" env:"ENABLE"`
	RootCAs               string `yaml:"root_ca_certs" json:"root_ca_certs" env:"ROOT_CA_CERTS"`
	Certificates          string `yaml:"certificates" json:"certificates" env:"CERTIFICATES"`
	InsecureSkipVerify    bool   `yaml:"insecure_skip_verify" json:"insecure_skip_verify" env:"INSECURE_SKIP_VERIFY"`
	VerifyPeerCertificate bool   `yaml:"verify_peer_certificate" json:"verify_peer_certificate" env:"VERIFY_PEER_CERTIFICATE"` //nolint:lll
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
