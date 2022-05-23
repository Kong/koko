package resource

import (
	"context"
	"fmt"

	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	"github.com/kong/koko/internal/model"
	"github.com/kong/koko/internal/model/json/extension"
	"github.com/kong/koko/internal/model/json/generator"
	"github.com/kong/koko/internal/model/json/validation"
	"github.com/kong/koko/internal/model/json/validation/typedefs"
)

const (
	TypeStatus = model.Type("status")

	SeverityWarning = "warning"
	SeverityError   = "error"

	maxConditions          = 64
	codeLength             = 5
	maxStatusMessageLength = 1024
)

func NewStatus() Status {
	return Status{
		Status: &v1.Status{},
	}
}

type Status struct {
	Status *v1.Status
}

func (r Status) ID() string {
	if r.Status == nil {
		return ""
	}
	return r.Status.Id
}

func (r Status) Type() model.Type {
	return TypeStatus
}

func (r Status) Resource() model.Resource {
	return r.Status
}

// SetResource implements the Object.SetResource interface.
func (r Status) SetResource(pr model.Resource) error { return model.SetResource(r, pr) }

func (r Status) Validate(ctx context.Context) error {
	return validation.Validate(string(TypeStatus), r.Status)
}

func (r Status) ProcessDefaults(ctx context.Context) error {
	if r.Status == nil {
		return fmt.Errorf("invalid nil resource")
	}
	defaultID(&r.Status.Id)
	return nil
}

func (r Status) Indexes() []model.Index {
	return []model.Index{
		{
			Name: "ctx_ref",
			Type: model.IndexUnique,
			Value: model.MultiValueIndex(r.Status.ContextReference.Type,
				r.Status.ContextReference.Id),
		},
	}
}

func init() {
	err := model.RegisterType(TypeStatus, &v1.Status{}, func() model.Object {
		return NewStatus()
	})
	if err != nil {
		panic(err)
	}

	statusSchema := &generator.Schema{
		Type: "object",
		Properties: map[string]*generator.Schema{
			"id": typedefs.ID,
			"context_reference": {
				Type: "object",
				Properties: map[string]*generator.Schema{
					"type": {
						Type: "string",
					},
					"id": typedefs.ID,
				},
				Required: []string{
					"type",
					"id",
				},
			},
			"conditions": {
				Type:      "array",
				MinLength: 1,
				MaxLength: maxConditions,
				Items: &generator.Schema{
					Type: "object",
					Properties: map[string]*generator.Schema{
						"code": {
							Type:      "string",
							MaxLength: codeLength,
							MinLength: codeLength,
						},
						"message": {
							Type:      "string",
							MaxLength: maxStatusMessageLength,
						},
						"severity": {
							Type: "string",
							Enum: []interface{}{
								SeverityWarning,
								SeverityError,
							},
						},
					},
					Required: []string{
						"code",
						"severity",
						"message",
					},
				},
			},
			"created_at": typedefs.UnixEpoch,
			"updated_at": typedefs.UnixEpoch,
		},
		AdditionalProperties: &falsy,
		Required: []string{
			"id",
			"context_reference",
			"conditions",
		},
		XKokoConfig: &extension.Config{DisableValidateEndpoint: true},
	}
	err = generator.Register(string(TypeStatus), statusSchema)
	if err != nil {
		panic(err)
	}
}
