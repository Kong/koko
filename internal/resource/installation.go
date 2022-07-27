package resource

import (
	"context"

	v1 "github.com/kong/koko/internal/gen/grpc/kong/nonpublic/v1"
	"github.com/kong/koko/internal/model"
	"github.com/kong/koko/internal/model/json/extension"
	"github.com/kong/koko/internal/model/json/generator"
	"github.com/kong/koko/internal/model/json/validation"
	"github.com/kong/koko/internal/model/json/validation/typedefs"
)

const (
	ID = "installation_id"
)

var TypeInstallation = model.Type("installation")

// NewInstallation returns a new Installation resource with the default Id value.
// The installationID is generated once in the lifetime of a koko cluster and must be unique.
// This unique value is generated at runtime and stored in Installation.Value.
func NewInstallation() Installation {
	return Installation{
		Installation: &v1.Installation{Id: ID},
	}
}

// Installation represents an installation of koko.
type Installation struct {
	Installation *v1.Installation
}

// ID returns the resource ID. This is always "installation_id" for resources of type Installation to avoid
// a race condition encountered when using a dynamic value for the ID.
func (r Installation) ID() string {
	return ID
}

// Type returns the type of this resource.
func (r Installation) Type() model.Type {
	return TypeInstallation
}

// Resource returns the underlying Installation resource.
func (r Installation) Resource() model.Resource {
	return r.Installation
}

// SetResource implements the Object.SetResource interface.
func (r Installation) SetResource(ir model.Resource) error { return model.SetResource(r, ir) }

// Validate wraps validation.Validate, which assesses the validity of the generated schema.
func (r Installation) Validate(ctx context.Context) error {
	return validation.Validate(string(TypeInstallation), r.Installation)
}

// ProcessDefaults sets the default values for the resource. This is a no-op for Installation.
func (r Installation) ProcessDefaults(ctx context.Context) error {
	return nil
}

func (r Installation) Indexes() []model.Index {
	return nil
}

func init() {
	err := model.RegisterType(TypeInstallation, &v1.Installation{}, func() model.Object {
		return NewInstallation()
	})
	if err != nil {
		panic(err)
	}

	installationSchema := &generator.Schema{
		Type: "object",
		Properties: map[string]*generator.Schema{
			"id":         typedefs.Name,
			"value":      typedefs.ID,
			"created_at": typedefs.UnixEpoch,
			"updated_at": typedefs.UnixEpoch,
		},
		AdditionalProperties: &falsy,
		Required: []string{
			"id",
			"value",
		},
		XKokoConfig: &extension.Config{DisableValidateEndpoint: true},
	}
	err = generator.Register(string(TypeInstallation), installationSchema)
	if err != nil {
		panic(err)
	}
}
