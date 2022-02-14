package resource

import (
	"fmt"

	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	"github.com/kong/koko/internal/model"
	"github.com/kong/koko/internal/model/json/generator"
	"github.com/kong/koko/internal/model/json/validation"
	"github.com/kong/koko/internal/model/json/validation/typedefs"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

const (
	// TypeTarget denotes the Target type.
	TypeTarget model.Type = "target"
)

var _ model.Object = Target{}

var (
	maxWeight           = 65535
	defaultWeight int32 = 100
)

func NewTarget() Target {
	return Target{
		Target: &v1.Target{},
	}
}

type Target struct {
	Target *v1.Target
}

func (t Target) ID() string {
	if t.Target == nil {
		return ""
	}
	return t.Target.Id
}

func (t Target) Type() model.Type {
	return TypeTarget
}

func (t Target) Resource() model.Resource {
	return t.Target
}

func (t Target) Indexes() []model.Index {
	res := []model.Index{
		{
			Name:      "target",
			Type:      model.IndexUnique,
			Value:     model.MultiValueIndex(t.Target.Upstream.Id, t.Target.Target),
			FieldName: "target",
		},
		{
			Name:        "upstream_id",
			Type:        model.IndexForeign,
			ForeignType: TypeUpstream,
			FieldName:   "upstream.id",
			Value:       t.Target.Upstream.Id,
		},
	}
	return res
}

func (t Target) Validate() error {
	err := validation.Validate(string(TypeTarget), t.Target)
	if err != nil {
		return err
	}
	err = validateTarget(t.Target.Target)
	if err != nil {
		// TODO(hbagdi): convert the error into *jsonschema.ValidationError
		// representation
		return err
	}
	return nil
}

func (t Target) ProcessDefaults() error {
	if t.Target == nil {
		return fmt.Errorf("invalid nil resource")
	}
	defaultID(&t.Target.Id)
	if t.Target.Weight == nil {
		t.Target.Weight = wrapperspb.Int32(defaultWeight)
	}
	var err error
	t.Target.Target, err = formatTarget(t.Target.Target)
	if err != nil {
		return fmt.Errorf("format target: %v", err)
	}
	return nil
}

func init() {
	err := model.RegisterType(TypeTarget, func() model.Object {
		return NewTarget()
	})
	if err != nil {
		panic(err)
	}

	zero := 0
	targetSchema := &generator.Schema{
		Type: "object",
		Properties: map[string]*generator.Schema{
			"id": typedefs.ID,
			"target": {
				Type:      "string",
				MinLength: 1,
				MaxLength: maxHostnameLength,
			},
			"weight": {
				Type:    "integer",
				Minimum: &zero,
				Maximum: maxWeight,
			},
			"tags":       typedefs.Tags,
			"created_at": typedefs.UnixEpoch,
			"updated_at": typedefs.UnixEpoch,
			"upstream":   typedefs.ReferenceObject,
		},
		AdditionalProperties: &falsy,
		Required: []string{
			"id",
			"target",
			"upstream",
		},
	}
	err = generator.Register(string(TypeTarget), targetSchema)
	if err != nil {
		panic(err)
	}
}

// TODO(hbagdi): implement validation for target.
func validateTarget(target string) error {
	return nil
}

// TODO(hbagdi): implement expansion for target.
func formatTarget(target string) (string, error) {
	return target, nil
}
