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
	maxVersionLength  = 128
	maxHostnameLength = 1024
	hashLength        = 32
	hashPattern       = "[a-z0-9]{32}"

	TypeNode = model.Type("node")

	NodeTypeKongProxy = "kong-proxy"
)

var (
	truthy = true
	falsy  = false
)

func NewNode() Node {
	return Node{
		Node: &v1.Node{},
	}
}

type Node struct {
	Node *v1.Node
}

func (r Node) ID() string {
	if r.Node == nil {
		return ""
	}
	return r.Node.Id
}

func (r Node) Type() model.Type {
	return TypeNode
}

func (r Node) Resource() model.Resource {
	return r.Node
}

// SetResource implements the Object.SetResource interface.
func (r Node) SetResource(pr model.Resource) error { return model.SetResource(r, pr) }

func (r Node) Validate(ctx context.Context) error {
	return validation.Validate(string(TypeNode), r.Node)
}

func (r Node) ProcessDefaults(ctx context.Context) error {
	if r.Node == nil {
		return fmt.Errorf("invalid nil resource")
	}
	defaultID(&r.Node.Id)
	return nil
}

func (r Node) Indexes() []model.Index {
	return nil
}

func init() {
	err := model.RegisterType(TypeNode, &v1.Node{}, func() model.Object {
		return NewNode()
	})
	if err != nil {
		panic(err)
	}

	nodeSchema := &generator.Schema{
		Type: "object",
		Properties: map[string]*generator.Schema{
			"id": typedefs.ID,
			"hostname": {
				Type:      "string",
				MinLength: 1,
				MaxLength: maxHostnameLength,
			},
			"type": {
				Type: "string",
				Enum: []interface{}{
					NodeTypeKongProxy,
				},
			},
			"last_ping": {
				Type:    "integer",
				Minimum: intP(1),
			},
			"config_hash": {
				Type:      "string",
				MinLength: hashLength,
				MaxLength: hashLength,
				Pattern:   hashPattern,
			},
			"version": {
				Type:      "string",
				MinLength: 1,
				MaxLength: maxVersionLength,
			},
			"created_at": typedefs.UnixEpoch,
			"updated_at": typedefs.UnixEpoch,
		},
		AdditionalProperties: &falsy,
		Required: []string{
			"id",
			"hostname",
			"type",
			"last_ping",
			"version",
		},
		XKokoConfig: &extension.Config{
			DisableValidateEndpoint: true,
			ResourceAPIPath:         "nodes",
		},
	}
	err = generator.DefaultRegistry.Register(string(TypeNode), nodeSchema)
	if err != nil {
		panic(err)
	}
}
