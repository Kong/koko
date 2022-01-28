package ws

import (
	"reflect"
	"testing"

	model "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	grpcKongUtil "github.com/kong/koko/internal/gen/grpc/kong/util/v1"
	"github.com/kong/koko/internal/status"
)

func Test_checkMissingPlugins(t *testing.T) {
	type args struct {
		requiredPlugins []string
		nodePlugins     []string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "empty returns empty",
		},
		{
			name: "required plugins are reported as missing",
			args: args{
				requiredPlugins: []string{"foo", "bar"},
				nodePlugins:     []string{"baz"},
			},
			want: []string{"foo", "bar"},
		},
		{
			name: "present plugins are not included in missing",
			args: args{
				requiredPlugins: []string{"foo", "bar", "baz"},
				nodePlugins:     []string{"baz"},
			},
			want: []string{"foo", "bar"},
		},
		{
			name: "no plugins are reported as missing when all required" +
				" plugins are present",
			args: args{
				requiredPlugins: []string{"foo", "bar", "baz"},
				nodePlugins:     []string{"foo", "bar", "baz"},
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := checkMissingPlugins(tt.args.requiredPlugins, tt.args.nodePlugins); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("checkMissingPlugins() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_conditionForMissingPlugins(t *testing.T) {
	type args struct {
		plugins []string
	}
	tests := []struct {
		name string
		args args
		want *model.Condition
	}{
		{
			name: "condition for one missing plugin",
			args: args{
				plugins: []string{"key-auth"},
			},
			want: &model.Condition{
				Code: status.DPMissingPlugin,
				Message: status.MessageForCode(status.DPMissingPlugin,
					"key-auth"),
				Severity: "error",
			},
		},
		{
			name: "condition for multiple missing plugins",
			args: args{
				plugins: []string{"key-auth", "basic-auth"},
			},
			want: &model.Condition{
				Code: status.DPMissingPlugin,
				Message: status.MessageForCode(status.DPMissingPlugin,
					"key-auth, basic-auth"),
				Severity: "error",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := conditionForMissingPlugins(tt.args.plugins); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("conditionForMissingPlugins() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_checkPreReqs(t *testing.T) {
	type args struct {
		attr   nodeAttributes
		checks []*grpcKongUtil.DataPlanePrerequisite
	}
	tests := []struct {
		name string
		args args
		want []*model.Condition
	}{
		{
			name: "missing plugins result in a corresponding conditions",
			args: args{
				attr: nodeAttributes{
					Plugins: []string{"foo", "bar"},
				},
				checks: []*grpcKongUtil.DataPlanePrerequisite{
					{
						Config: &grpcKongUtil.DataPlanePrerequisite_RequiredPlugins{
							RequiredPlugins: &grpcKongUtil.RequiredPluginsFilter{
								RequiredPlugins: []string{
									"rate-limiting",
									"http-log",
								},
							},
						},
					},
				},
			},
			want: []*model.Condition{
				{
					Code: status.DPMissingPlugin,
					Message: status.MessageForCode(status.DPMissingPlugin,
						"rate-limiting, http-log"),
					Severity: "error",
				},
			},
		},
		{
			name: "no missing plugins result in a no conditions",
			args: args{
				attr: nodeAttributes{
					Plugins: []string{"rate-limiting", "http-log"},
				},
				checks: []*grpcKongUtil.DataPlanePrerequisite{
					{
						Config: &grpcKongUtil.DataPlanePrerequisite_RequiredPlugins{
							RequiredPlugins: &grpcKongUtil.RequiredPluginsFilter{
								RequiredPlugins: []string{
									"rate-limiting",
									"http-log",
								},
							},
						},
					},
				},
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := checkPreReqs(tt.args.attr, tt.args.checks); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("checkPreReqs() = %v, want %v", got, tt.want)
			}
		})
	}
}
