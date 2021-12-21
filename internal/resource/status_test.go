package resource

import (
	"testing"

	"github.com/google/uuid"
	model "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	"github.com/kong/koko/internal/model/json/validation"
	"github.com/stretchr/testify/require"
)

func TestNewStatus(t *testing.T) {
	s := NewStatus()
	require.NotNil(t, s)
	require.NotNil(t, s.Status)
}

func TestStatus_Type(t *testing.T) {
	require.Equal(t, TypeStatus, NewStatus().Type())
}

func TestStatus_ProcessDefaults(t *testing.T) {
	s := NewStatus()
	require.Nil(t, s.ProcessDefaults())
	require.NotPanics(t, func() {
		uuid.MustParse(s.ID())
	})
}

func TestStatus_Validate(t *testing.T) {
	tests := []struct {
		name    string
		Status  func() Status
		wantErr bool
		Errs    []*model.ErrorDetail
	}{
		{
			name: "empty status throws an error",
			Status: func() Status {
				return NewStatus()
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type: model.ErrorType_ERROR_TYPE_ENTITY,
					Messages: []string{
						"missing properties: 'id', 'context_reference', " +
							"'conditions'",
					},
				},
			},
		},
		{
			name: "default status throws an error",
			Status: func() Status {
				s := NewStatus()
				_ = s.ProcessDefaults()
				return s
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type: model.ErrorType_ERROR_TYPE_ENTITY,
					Messages: []string{
						"missing properties: 'context_reference', " +
							"'conditions'",
					},
				},
			},
		},
		{
			name: "status without any conditions throws an error",
			Status: func() Status {
				s := NewStatus()
				s.Status = &model.Status{
					Id: uuid.NewString(),
					ContextReference: &model.EntityReference{
						Type: string(TypeService),
						Id:   uuid.NewString(),
					},
					Conditions: []*model.Condition{},
				}
				return s
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type: model.ErrorType_ERROR_TYPE_ENTITY,
					Messages: []string{
						"missing properties: 'conditions'",
					},
				},
			},
		},
		{
			name: "valid status doesn't thrown an error",
			Status: func() Status {
				s := NewStatus()
				s.Status = &model.Status{
					Id: uuid.NewString(),
					ContextReference: &model.EntityReference{
						Type: string(TypeService),
						Id:   uuid.NewString(),
					},
					Conditions: []*model.Condition{
						{
							Code:     "R0023",
							Message:  "foo bar",
							Severity: SeverityError,
						},
					},
				}
				return s
			},
			wantErr: false,
		},
		{
			name: "invalid uuid in context_reference.id throws an error",
			Status: func() Status {
				s := NewStatus()
				s.Status = &model.Status{
					Id: uuid.NewString(),
					ContextReference: &model.EntityReference{
						Type: string(TypeService),
						Id:   "borked",
					},
					Conditions: []*model.Condition{
						{
							Code:     "R0023",
							Message:  "foo bar",
							Severity: SeverityError,
						},
					},
				}
				return s
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type:  model.ErrorType_ERROR_TYPE_FIELD,
					Field: "context_reference.id",
					Messages: []string{
						"must be a valid UUID",
					},
				},
			},
		},
		{
			name: "reference without id throws an error",
			Status: func() Status {
				s := NewStatus()
				s.Status = &model.Status{
					Id: uuid.NewString(),
					ContextReference: &model.EntityReference{
						Type: string(TypeService),
					},
					Conditions: []*model.Condition{
						{
							Code:     "R0023",
							Message:  "foo bar",
							Severity: SeverityError,
						},
					},
				}
				return s
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type:  model.ErrorType_ERROR_TYPE_FIELD,
					Field: "context_reference",
					Messages: []string{
						"missing properties: 'id'",
					},
				},
			},
		},
		{
			name: "reference without type throws an error",
			Status: func() Status {
				s := NewStatus()
				s.Status = &model.Status{
					Id: uuid.NewString(),
					ContextReference: &model.EntityReference{
						Id: uuid.NewString(),
					},
					Conditions: []*model.Condition{
						{
							Code:     "R0023",
							Message:  "foo bar",
							Severity: SeverityError,
						},
					},
				}
				return s
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type:  model.ErrorType_ERROR_TYPE_FIELD,
					Field: "context_reference",
					Messages: []string{
						"missing properties: 'type'",
					},
				},
			},
		},
		{
			name: "condition without severity and code throws an error",
			Status: func() Status {
				s := NewStatus()
				s.Status = &model.Status{
					Id: uuid.NewString(),
					ContextReference: &model.EntityReference{
						Type: string(TypeRoute),
						Id:   uuid.NewString(),
					},
					Conditions: []*model.Condition{
						{
							Message: "foo bar",
						},
					},
				}
				return s
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type:  model.ErrorType_ERROR_TYPE_FIELD,
					Field: "conditions[0]",
					Messages: []string{
						"missing properties: 'code', 'severity'",
					},
				},
			},
		},
		{
			name: "condition with invalid code throws an error",
			Status: func() Status {
				s := NewStatus()
				s.Status = &model.Status{
					Id: uuid.NewString(),
					ContextReference: &model.EntityReference{
						Type: string(TypeRoute),
						Id:   uuid.NewString(),
					},
					Conditions: []*model.Condition{
						{
							Code:     "borked",
							Message:  "foo bar",
							Severity: SeverityError,
						},
					},
				}
				return s
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type:  model.ErrorType_ERROR_TYPE_FIELD,
					Field: "conditions[0].code",
					Messages: []string{
						"length must be <= 5, but got 6",
					},
				},
			},
		},
		{
			name: "condition with invalid severity throws an error",
			Status: func() Status {
				s := NewStatus()
				s.Status = &model.Status{
					Id: uuid.NewString(),
					ContextReference: &model.EntityReference{
						Type: string(TypeRoute),
						Id:   uuid.NewString(),
					},
					Conditions: []*model.Condition{
						{
							Code:     "F4242",
							Message:  "foo bar",
							Severity: "yolo",
						},
					},
				}
				return s
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type:  model.ErrorType_ERROR_TYPE_FIELD,
					Field: "conditions[0].severity",
					Messages: []string{
						`value must be one of "warning", "error"`,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.Status().Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.Errs != nil {
				verr, _ := err.(validation.Error)
				require.ElementsMatch(t, tt.Errs, verr.Errs)
			}
		})
	}
}
