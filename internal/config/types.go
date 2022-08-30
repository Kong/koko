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

type SQLite struct {
	InMemory bool   `yaml:"in_memory" json:"in_memory" env:"IN_MEMORY"`
	Filename string `yaml:"filename" json:"filename" env:"FILENAME"`
}

type Postgres struct {
	DBName      string              `yaml:"db_name" json:"db_name" env:"DB_NAME"`
	Hostname    string              `yaml:"hostname" json:"hostname" env:"HOSTNAME"`
	ReadReplica PostgresReadReplica `yaml:"read_replica" json:"read_replica" env-prefix:"READ_REPLICA_"`
	Port        int                 `yaml:"port" json:"port" env:"PORT"`
	User        string              `yaml:"user" json:"user" env:"USER"`
	Password    string              `yaml:"password" json:"password" env:"PASSWORD"`
	TLS         PostgresTLS         `yaml:"tls" json:"tls" env-prefix:"TLS_"`
}

type PostgresTLS struct {
	Enable       bool   `yaml:"enable" json:"enable" env:"ENABLE"`
	CABundlePath string `yaml:"ca_bundle_path" json:"ca_bundle_path" env:"CA_BUNDLE_PATH"`
}

type PostgresReadReplica struct {
	Hostname string `yaml:"hostname" json:"hostname" env:"HOSTNAME"`
}

type Database struct {
	Dialect      string   `yaml:"dialect" json:"dialect" env:"DIALECT" env-default:"sqlite3"`
	SQLite       SQLite   `yaml:"sqlite" json:"sqlite" env-prefix:"SQLITE_"`
	Postgres     Postgres `yaml:"postgres" json:"postgres" env-prefix:"POSTGRES_"`
	QueryTimeout string   `yaml:"query_timeout" json:"query_timeout" env:"QUERY_TIMEOUT" env-default:"5s"`
}

// Metrics config.
type Metrics struct {
	// ClientType metrics client type e.g. prometheus, datadog.
	ClientType string `yaml:"client_type" json:"client_type" env:"CLIENT_TYPE" env-default:"noop"`
}

type Config struct {
	Log                     Log           `yaml:"log" json:"log" env-prefix:"KOKO_LOG_"`
	Admin                   AdminServer   `yaml:"admin_server" json:"admin_server" env-prefix:"KOKO_ADMIN_SERVER_"`
	Control                 ControlServer `yaml:"control_server" json:"control_server" env-prefix:"KOKO_CONTROL_SERVER_"`
	Database                Database      `yaml:"database" json:"database" env-prefix:"KOKO_DATABASE_"`
	Metrics                 Metrics       `yaml:"metrics" json:"metrics" env-prefix:"KOKO_METRICS_"`
	DisableAnonymousReports bool          `yaml:"disable_anonymous_reports" json:"disable_anonymous_reports" env-prefix:"KOKO_DISABLE_ANONYMOUS_REPORTS"` //nolint:lll
}
