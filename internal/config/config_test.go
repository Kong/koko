package config

import (
	"testing"
	"time"

	"github.com/kong/koko/internal/db"
	"github.com/kong/koko/internal/persistence"
	"github.com/kong/koko/internal/persistence/postgres"
	"github.com/stretchr/testify/require"
)

var defaultConfig = Config{
	Log: Log{
		Level:  "info",
		Format: "json",
	},
	Admin: AdminServer{
		Address: ":3000",
	},
	Control: ControlServer{},
	Database: Database{
		Dialect:      db.DialectSQLite3,
		QueryTimeout: "5s",
		Postgres: Postgres{
			Pool: defaultPostgresPool,
		},
	},
	Metrics: Metrics{
		ClientType: "noop",
		Prometheus: PrometheusMetrics{
			Enable:  false,
			Address: ":9090",
		},
	},
	DisableAnonymousReports: false,
}

var defaultPostgresPool = PostgresPool{
	Name:              postgres.DefaultPool,
	MaxConns:          persistence.DefaultMaxConn,
	MinConns:          persistence.DefaultMinConn,
	MaxConnLifetime:   persistence.DefaultMaxConnLifetime,
	MaxConnIdleTime:   persistence.DefaultMaxConnIdleTime,
	HealthCheckPeriod: persistence.DefaultHealthCheckPeriod,
}

func TestGet(t *testing.T) {
	type args struct {
		filename string
		envVars  map[string]string
	}
	duration20m, _ := time.ParseDuration("20m")
	duration10m, _ := time.ParseDuration("10m")
	duration30s, _ := time.ParseDuration("30s")
	tests := []struct {
		name      string
		args      args
		want      Config
		errString string
	}{
		{
			name: "gets default configuration when no file is specified",
			want: defaultConfig,
		},
		{
			name: "gets default configuration when file is missing",
			args: args{
				filename: "does-not-exist.yaml",
			},
			want: defaultConfig,
		},
		{
			name: "overrides from file",
			args: args{
				filename: "good-0.yaml",
			},
			want: Config{
				Log: Log{
					Level:  "debug",
					Format: "console",
				},
				Admin: AdminServer{
					Address: ":3001",
				},
				Control: ControlServer{
					TLSCertPath: "foo.crt",
					TLSKeyPath:  "bar.key",
				},
				Database: Database{
					Dialect: db.DialectPostgres,
					SQLite: SQLite{
						InMemory: true,
						Filename: "test.db",
					},
					Postgres: Postgres{
						DBName:   "koko",
						Hostname: "localhost",
						ReadReplica: PostgresReadReplica{
							Hostname: "read-localhost",
						},
						Port:     5433,
						User:     "koko",
						Password: "koko",
						TLS: PostgresTLS{
							Enable:       true,
							CABundlePath: "/tmp/foo.crt",
						},
						Pool: PostgresPool{
							Name:              "pgx",
							MaxConns:          30,
							MinConns:          5,
							MaxConnLifetime:   duration20m,
							MaxConnIdleTime:   duration10m,
							HealthCheckPeriod: duration30s,
						},
					},
					QueryTimeout: "2s",
				},
				Metrics: Metrics{
					ClientType: "noop",
					Prometheus: PrometheusMetrics{
						Enable:  false,
						Address: ":9090",
					},
				},
				DisableAnonymousReports: true,
			},
		},
		{
			name: "overrides from json file",
			args: args{
				filename: "good-1.json",
			},
			want: Config{
				Log: Log{
					Level:  "debug",
					Format: "console",
				},
				Admin: AdminServer{
					Address: ":3001",
				},
				Control: ControlServer{
					TLSCertPath: "foo.crt",
					TLSKeyPath:  "bar.key",
				},
				Database: Database{
					Dialect: db.DialectPostgres,
					SQLite: SQLite{
						InMemory: true,
						Filename: "test.db",
					},
					Postgres: Postgres{
						DBName:   "koko",
						Hostname: "localhost",
						ReadReplica: PostgresReadReplica{
							Hostname: "read-localhost",
						},
						Port:     5433,
						User:     "koko",
						Password: "koko",
						TLS: PostgresTLS{
							Enable:       true,
							CABundlePath: "/tmp/foo.crt",
						},
						Pool: defaultPostgresPool,
					},
					QueryTimeout: "2s",
				},
				Metrics: Metrics{
					ClientType: "noop",
					Prometheus: PrometheusMetrics{
						Enable:  false,
						Address: ":9090",
					},
				},
				DisableAnonymousReports: true,
			},
		},
		{
			name: "configuration can be provided via env vars",
			args: args{
				envVars: map[string]string{
					"KOKO_LOG_LEVEL":                                "error",
					"KOKO_LOG_FORMAT":                               "FOOBAR",
					"KOKO_DATABASE_DIALECT":                         db.DialectPostgres,
					"KOKO_DATABASE_POSTGRES_READ_REPLICA_HOSTNAME":  "foobar",
					"KOKO_DATABASE_POSTGRES_TLS_ENABLE":             "true",
					"KOKO_DATABASE_POSTGRES_POOL_MAX_CONN_LIFETIME": "20m",
					"KOKO_METRICS_PROMETHEUS_ENABLE":                "true",
				},
			},
			want: Config{
				Log: Log{
					Level:  "error",
					Format: "FOOBAR",
				},
				Admin: AdminServer{
					Address: ":3000",
				},
				Database: Database{
					Dialect: db.DialectPostgres,
					Postgres: Postgres{
						ReadReplica: PostgresReadReplica{
							Hostname: "foobar",
						},
						TLS: PostgresTLS{
							Enable: true,
						},
						Pool: PostgresPool{
							Name:              postgres.DefaultPool,
							MaxConns:          persistence.DefaultMaxConn,
							MinConns:          persistence.DefaultMinConn,
							MaxConnLifetime:   duration20m,
							MaxConnIdleTime:   persistence.DefaultMaxConnIdleTime,
							HealthCheckPeriod: persistence.DefaultHealthCheckPeriod,
						},
					},
					QueryTimeout: "5s",
				},
				Metrics: Metrics{
					ClientType: "noop",
					Prometheus: PrometheusMetrics{
						Enable:  true,
						Address: ":9090",
					},
				},
				DisableAnonymousReports: false,
			},
		},
		{
			name: "environment variables take the highest priority",
			args: args{
				filename: "good-1.json",
				envVars: map[string]string{
					"KOKO_LOG_LEVEL":  "error",
					"KOKO_LOG_FORMAT": "FOOBAR",
				},
			},
			want: Config{
				Log: Log{
					Level:  "error",
					Format: "FOOBAR",
				},
				Admin: AdminServer{
					Address: ":3001",
				},
				Control: ControlServer{
					TLSCertPath: "foo.crt",
					TLSKeyPath:  "bar.key",
				},
				Database: Database{
					Dialect: db.DialectPostgres,
					SQLite: SQLite{
						InMemory: true,
						Filename: "test.db",
					},
					Postgres: Postgres{
						DBName:   "koko",
						Hostname: "localhost",
						ReadReplica: PostgresReadReplica{
							Hostname: "read-localhost",
						},
						Port:     5433,
						User:     "koko",
						Password: "koko",
						TLS: PostgresTLS{
							Enable:       true,
							CABundlePath: "/tmp/foo.crt",
						},
						Pool: defaultPostgresPool,
					},
					QueryTimeout: "2s",
				},
				Metrics: Metrics{
					ClientType: "noop",
					Prometheus: PrometheusMetrics{
						Enable:  false,
						Address: ":9090",
					},
				},
				DisableAnonymousReports: true,
			},
		},
		{
			name: "invalid YAML errors",
			args: args{
				filename: "bad.yaml",
			},
			errString: "read config: config file parsing error" +
				": yaml: unmarshal errors",
		},
		{
			name: "invalid JSON errors",
			args: args{
				filename: "bad.json",
			},
			errString: "read config: config file parsing error: invalid" +
				" character 'b' looking for beginning of value",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.args.envVars {
				t.Setenv(k, v)
			}
			filename := tt.args.filename
			if filename != "" {
				filename = "testdata/" + filename
			}
			got, err := Get(filename)
			if tt.errString != "" {
				require.Error(t, err)
				require.ErrorContains(t, err, tt.errString)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, got)
			}
		})
	}
}
