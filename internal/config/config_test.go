package config

import (
	"os"
	"reflect"
	"testing"
)

var expectedDefaultConfig = Config{
	Log: Log{
		Level: "info",
	},
	Admin: Admin{
		Listeners: []Listener{
			{
				Address:  ":3000",
				Protocol: "http",
			},
		},
	},
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name       string
		config     func() Config
		wantErrLen int
	}{
		{
			name: "default",
			config: func() Config {
				c, _ := Get("")
				return c
			},
			wantErrLen: 0,
		},
		{
			name: "invalid log.level",
			config: func() Config {
				c, _ := Get("")
				c.Log.Level = "foo"
				return c
			},
			wantErrLen: 1,
		},
	}
	t.Parallel()
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := Validate(tt.config()); !reflect.DeepEqual(len(got),
				tt.wantErrLen) {
				t.Errorf("Validate() = %v, want %v", len(got), tt.wantErrLen)
			}
		})
	}
}

func TestGet(t *testing.T) {
	type args struct {
		filename string
		envVars  map[string]string
	}
	tests := []struct {
		name    string
		args    args
		want    func() Config
		wantErr bool
	}{
		{
			name: "invalid file path",
			args: args{
				filename: "./testdata/does-not-exist",
			},
			want:    func() Config { return Config{} },
			wantErr: true,
		},
		{
			name: "default configuration",
			args: args{
				filename: "",
			},
			want: func() Config {
				return expectedDefaultConfig
			},
			wantErr: false,
		},
		{
			name: "valid yaml file",
			args: args{
				filename: "./testdata/good.yaml",
			},
			want: func() Config {
				c := expectedDefaultConfig
				c.Log.Level = "debug"
				c.Admin.Listeners = []Listener{
					{
						Address:  ":4000",
						Protocol: "grpc",
					},
				}
				return c
			},
			wantErr: false,
		},
		{
			name: "bad yaml file",
			args: args{
				filename: "./testdata/bad.yaml",
			},
			want:    func() Config { return Config{} },
			wantErr: true,
		},
		{
			name: "bad json file",
			args: args{
				filename: "./testdata/bad.json",
			},
			want:    func() Config { return Config{} },
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.args.envVars {
				os.Setenv(k, v)
				defer func(k string) {
					os.Unsetenv(k)
				}(k)
			}
			got, err := Get(tt.args.filename)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			want := tt.want()
			if !reflect.DeepEqual(got, want) {
				t.Errorf("Get() = %v, want %v", got, want)
			}
		})
	}
}
