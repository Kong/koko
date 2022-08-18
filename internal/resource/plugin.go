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
	"github.com/kong/koko/internal/plugin"
)

const (
	TypePlugin = model.Type("plugin")

	// OrderingRuleTitle denotes the title of the ordering-related rule.
	OrderingRuleTitle = "ordering"
)

var validator plugin.Validator

func SetValidator(v plugin.Validator) {
	validator = v
}

func NewPlugin() Plugin {
	return Plugin{
		Plugin: &v1.Plugin{},
	}
}

type Plugin struct {
	Plugin *v1.Plugin
}

func (r Plugin) ID() string {
	if r.Plugin == nil {
		return ""
	}
	return r.Plugin.Id
}

func (r Plugin) Type() model.Type {
	return TypePlugin
}

func (r Plugin) Resource() model.Resource {
	return r.Plugin
}

// SetResource implements the Object.SetResource interface.
func (r Plugin) SetResource(pr model.Resource) error { return model.SetResource(r, pr) }

func (r Plugin) Validate(ctx context.Context) error {
	err := validation.Validate(string(TypePlugin), r.Plugin)
	if err != nil {
		return err
	}
	return validator.Validate(ctx, r.Plugin)
}

func (r Plugin) ProcessDefaults(ctx context.Context) error {
	err := validator.ProcessDefaults(ctx, r.Plugin)
	return err
}

func (r Plugin) Indexes() []model.Index {
	serviceID, routeID, consumerID := "", "", ""
	if r.Plugin.Service != nil {
		serviceID = r.Plugin.Service.Id
	}
	if r.Plugin.Route != nil {
		routeID = r.Plugin.Route.Id
	}
	if r.Plugin.Consumer != nil {
		consumerID = r.Plugin.Consumer.Id
	}
	uniqueValue := fmt.Sprintf("%s.%s.%s.%s", r.Plugin.Name,
		serviceID, routeID, consumerID)

	res := []model.Index{
		{
			Name: "unique-plugin-per-entity",
			// TODO(hbagdi): needs IndexUniqueMulti for multiple fields?
			Type:  model.IndexUnique,
			Value: uniqueValue,
			// TODO(hbagdi): maybe needs FieldNames?
			FieldName: "",
		},
	}
	if r.Plugin.Route != nil {
		res = append(res, model.Index{
			Name:        "route_id",
			Type:        model.IndexForeign,
			ForeignType: TypeRoute,
			FieldName:   "route.id",
			Value:       r.Plugin.Route.Id,
		})
	}
	if r.Plugin.Service != nil {
		res = append(res, model.Index{
			Name:        "service_id",
			Type:        model.IndexForeign,
			ForeignType: TypeService,
			FieldName:   "service.id",
			Value:       r.Plugin.Service.Id,
		})
	}
	if r.Plugin.Consumer != nil {
		res = append(res, model.Index{
			Name:        "consumer_id",
			Type:        model.IndexForeign,
			ForeignType: TypeConsumer,
			FieldName:   "consumer.id",
			Value:       r.Plugin.Consumer.Id,
		})
	}
	return res
}

func init() {
	err := model.RegisterType(TypePlugin, &v1.Plugin{}, func() model.Object {
		return NewPlugin()
	})
	if err != nil {
		panic(err)
	}

	const maxProtocols = 8
	pluginSchema := &generator.Schema{
		Type: "object",
		Properties: map[string]*generator.Schema{
			"id":         typedefs.ID,
			"name":       typedefs.PluginName,
			"created_at": typedefs.UnixEpoch,
			"updated_at": typedefs.UnixEpoch,
			"enabled": {
				Type: "boolean",
			},
			"tags": typedefs.Tags,
			"protocols": {
				Type:     "array",
				Items:    typedefs.Protocol,
				MaxItems: maxProtocols,
			},
			"config": {
				Type:                 "object",
				AdditionalProperties: &truthy,
			},
			"service":  typedefs.ReferenceObject,
			"route":    typedefs.ReferenceObject,
			"consumer": typedefs.ReferenceObject,
			"ordering": {
				Type:                 "object",
				AdditionalProperties: &truthy,
			},
		},
		AdditionalProperties: &falsy,
		Required: []string{
			"name",
		},
		XKokoConfig: &extension.Config{
			ResourceAPIPath: "plugins",
		},
		AllOf: []*generator.Schema{
			{
				Title: OrderingRuleTitle,
				Description: "'ordering' is a Kong Enterprise-only feature. " +
					"Please upgrade to Kong Enterprise to use this feature.",
				Not: &generator.Schema{
					Required: []string{"ordering"},
				},
			},
		},
	}
	err = generator.Register(string(TypePlugin), pluginSchema)
	if err != nil {
		panic(err)
	}
}
