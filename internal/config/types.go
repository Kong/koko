package config

type Log struct {
	Level string `json:"level,omitempty"`
}

type Listener struct {
	Address  string `json:"address,omitempty"`
	Protocol string `json:"protocol,omitempty"`
}

type Admin struct {
	Listeners []Listener `json:"listeners,omitempty"`
}

type ControlServer struct {
	TLSCertPath string `json:"tls_cert_path,omitempty"`
	TLSKeyPath  string `json:"tls_key_path,omitempty"`
}

type SQLite struct {
	InMemory bool   `json:"in_memory,omitempty"`
	Filename string `json:"filename,omitempty"`
}

type Postgres struct {
	DBName         string `json:"db_name,omitempty"`
	Hostname       string `json:"hostname,omitempty"`
	Port           int    `json:"port,omitempty"`
	User           string `json:"user,omitempty"`
	Password       string `json:"password,omitempty"`
	EnableTLS      bool   `json:"enable_tls"`
	CABundleFSPath string `json:"ca_bundle_fs_path"`
}

type Database struct {
	Dialect      string   `json:"dialect,omitempty"`
	SQLite       SQLite   `json:"sqlite,omitempty"`
	Postgres     Postgres `json:"postgres,omitempty"`
	QueryTimeout string   `json:"query_timeout,omitempty"`
}

// Metrics config.
type Metrics struct {
	// ClientType metrics client type e.g. prometheus, datadog.
	ClientType string `json:"client_type,omitempty"`
}

type Config struct {
	Log                     Log           `json:"log,omitempty"`
	Admin                   Admin         `json:"admin,omitempty"`
	Control                 ControlServer `json:"control_server,omitempty"`
	Database                Database      `json:"database,omitempty"`
	Metrics                 Metrics       `json:"metrics,omitempty"`
	DisableAnonymousReports bool          `json:"disable_anonymous_reports"`
}
