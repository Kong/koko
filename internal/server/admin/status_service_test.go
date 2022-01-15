package admin

import (
	"testing"

	"github.com/google/uuid"
)

func Test_validateRefs(t *testing.T) {
	type args struct {
		refType string
		refID   string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "valid id and type returns no error",
			args: args{
				refType: "route",
				refID:   uuid.NewString(),
			},
			wantErr: false,
		},
		{
			name: "invalid id throws an error",
			args: args{
				refType: "route",
				refID:   "emoji-id-🔥",
			},
			wantErr: true,
		},
		{
			name: "invalid type throws an error",
			args: args{
				refType: "route42",
				refID:   uuid.NewString(),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateRefs(tt.args.refType, tt.args.refID); (err != nil) != tt.wantErr {
				t.Errorf("validateRefs() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
