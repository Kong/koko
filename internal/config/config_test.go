package config

import (
	"testing"

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
		Dialect:      "sqlite3",
		QueryTimeout: "5s",
	},
	Metrics: Metrics{
		ClientType: "noop",
	},
	DisableAnonymousReports: false,
}

func TestGet(t *testing.T) {
	type args struct {
		filename string
		envVars  map[string]string
	}
	tests := []struct {
		name    string
		args    args
		want    Config
		wantErr bool
	}{
		{
			name:    "gets default configuration when no file is specified",
			want:    defaultConfig,
			wantErr: false,
		},
		{
			name: "gets default configuration when file is missing",
			args: args{
				filename: "does-not-exist.yaml",
			},
			want:    defaultConfig,
			wantErr: false,
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
					Dialect: "postgres",
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
					},
					QueryTimeout: "2s",
				},
				Metrics: Metrics{
					ClientType: "noop",
				},
				DisableAnonymousReports: true,
			},
			wantErr: false,
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
					Dialect: "postgres",
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
					},
					QueryTimeout: "2s",
				},
				Metrics: Metrics{
					ClientType: "noop",
				},
				DisableAnonymousReports: true,
			},
			wantErr: false,
		},
		{
			name: "configuration can be provided via env vars",
			args: args{
				envVars: map[string]string{
					"KOKO_LOG_LEVEL":                               "error",
					"KOKO_LOG_FORMAT":                              "FOOBAR",
					"KOKO_DATABASE_DIALECT":                        "postgres",
					"KOKO_DATABASE_POSTGRES_READ_REPLICA_HOSTNAME": "foobar",
					"KOKO_DATABASE_POSTGRES_TLS_ENABLE":            "true",
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
					Dialect: "postgres",
					Postgres: Postgres{
						ReadReplica: PostgresReadReplica{
							Hostname: "foobar",
						},
						TLS: PostgresTLS{
							Enable: true,
						},
					},
					QueryTimeout: "5s",
				},
				Metrics: Metrics{
					ClientType: "noop",
				},
				DisableAnonymousReports: false,
			},
			wantErr: false,
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
					Dialect: "postgres",
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
					},
					QueryTimeout: "2s",
				},
				Metrics: Metrics{
					ClientType: "noop",
				},
				DisableAnonymousReports: true,
			},
			wantErr: false,
		},
		{
			name: "invalid YAML errors",
			args: args{
				filename: "bad.yaml",
			},
			wantErr: true,
		},
		{
			name: "invalid JSON errors",
			args: args{
				filename: "bad.json",
			},
			wantErr: true,
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
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			require.Equal(t, tt.want, got)
		})
	}
}
