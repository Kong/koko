package resource

import (
	"fmt"
	"net/http"

	"github.com/imdario/mergo"
	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	"github.com/kong/koko/internal/model"
	"github.com/kong/koko/internal/model/json/generator"
	"github.com/kong/koko/internal/model/json/validation"
	"github.com/kong/koko/internal/model/json/validation/typedefs"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

const (
	// TypeRoute denotes the Route type.
	TypeRoute model.Type = "route"
)

var (
	defaultRoute = &v1.Route{
		Protocols:               []string{typedefs.ProtocolHTTP, typedefs.ProtocolHTTPS},
		RegexPriority:           wrapperspb.Int32(0),
		PreserveHost:            wrapperspb.Bool(false),
		StripPath:               wrapperspb.Bool(true),
		RequestBuffering:        wrapperspb.Bool(true),
		ResponseBuffering:       wrapperspb.Bool(true),
		PathHandling:            "v0",
		HttpsRedirectStatusCode: http.StatusUpgradeRequired,
	}
	_ model.Object = Route{}
)

func NewRoute() Route {
	return Route{
		Route: &v1.Route{},
	}
}

type Route struct {
	Route *v1.Route
}

func (r Route) ID() string {
	if r.Route == nil {
		return ""
	}
	return r.Route.Id
}

func (r Route) Type() model.Type {
	return TypeRoute
}

func (r Route) Resource() model.Resource {
	return r.Route
}

func (r Route) Indexes() []model.Index {
	res := []model.Index{
		{
			Name:      "name",
			Type:      model.IndexUnique,
			Value:     r.Route.Name,
			FieldName: "name",
		},
	}
	if r.Route.Service != nil {
		res = append(res, model.Index{
			Name:        "svc_id",
			Type:        model.IndexForeign,
			ForeignType: TypeService,
			FieldName:   "service.id",
			Value:       r.Route.Service.Id,
		})
	}
	return res
}

func (r Route) Validate() error {
	return validation.Validate(string(TypeRoute), r.Route)
}

func (r Route) ValidateCompat() error {
	return r.Validate()
}

func (r Route) ProcessDefaults() error {
	if r.Route == nil {
		return fmt.Errorf("invalid nil resource")
	}
	err := mergo.Merge(r.Route, defaultRoute,
		mergo.WithTransformers(wrappersPBTransformer{}))
	if err != nil {
		return err
	}
	defaultID(&r.Route.Id)
	addTZ(r.Route)
	return nil
}

func init() {
	err := model.RegisterType(TypeRoute, func() model.Object {
		return NewRoute()
	})
	if err != nil {
		panic(err)
	}

	routeSchema := &generator.Schema{
		Type: "object",
		Properties: map[string]*generator.Schema{
			"id":   typedefs.ID,
			"name": typedefs.Name,
			"tags": typedefs.Tags,
		},
		AdditionalProperties: true,
		Required: []string{
			"id",
		},
	}
	err = generator.Register(string(TypeRoute), routeSchema)
	if err != nil {
		panic(err)
	}
}
