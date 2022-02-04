package resource

import (
	"fmt"

	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	"github.com/kong/koko/internal/model"
	"github.com/kong/koko/internal/model/json/generator"
	"github.com/kong/koko/internal/model/json/validation"
	"github.com/kong/koko/internal/model/json/validation/typedefs"
)

const (
	TypeConsumer = model.Type("consumer")
)

func NewConsumer() Consumer {
	return Consumer{
		Consumer: &v1.Consumer{},
	}
}

type Consumer struct {
	Consumer *v1.Consumer
}

func (c Consumer) ID() string {
	if c.Consumer == nil {
		return ""
	}
	return c.Consumer.Id
}

func (c Consumer) Type() model.Type {
	return TypeConsumer
}

func (c Consumer) Validate() error {
	return validation.Validate(string(TypeConsumer), c.Consumer)
}

func (c Consumer) ProcessDefaults() error {
	if c.Consumer == nil {
		return fmt.Errorf("invalid nil resource")
	}
	defaultID(&c.Consumer.Id)
	return nil
}

func (c Consumer) Resource() model.Resource {
	return c.Consumer
}

func (c Consumer) Indexes() []model.Index {
	return []model.Index{
		{
			Name:      "username",
			Type:      model.IndexUnique,
			Value:     c.Consumer.Username,
			FieldName: "username",
		},
		{
			Name:      "custom_id",
			Type:      model.IndexUnique,
			Value:     c.Consumer.CustomId,
			FieldName: "custom_id",
		},
	}
}

func init() {
	err := model.RegisterType(TypeConsumer, func() model.Object {
		return NewConsumer()
	})
	if err != nil {
		panic(err)
	}

	consumerSchema := &generator.Schema{
		Type: "object",
		Properties: map[string]*generator.Schema{
			"id":         typedefs.ID,
			"username":   typedefs.Name,
			"created_at": typedefs.UnixEpoch,
			"updated_at": typedefs.UnixEpoch,
			"custom_id":  typedefs.Name, // Not a UUID
			"tags":       typedefs.Tags,
		},
		AdditionalProperties: &falsy,
		Required:             []string{"id"},
		AnyOf: []*generator.Schema{
			{
				Required: []string{"custom_id"},
			},
			{
				Required: []string{"username"},
			},
		},
	}
	err = generator.Register(string(TypeConsumer), consumerSchema)
	if err != nil {
		panic(err)
	}
}
