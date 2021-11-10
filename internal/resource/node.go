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
	maxVersionLength  = 128
	maxHostnameLength = 1024

	TypeNode = model.Type("node")

	NodeTypeKongProxy = "kong-proxy"
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

func (r Node) Validate() error {
	return validation.Validate(string(TypeNode), r.Node)
}

func (r Node) ProcessDefaults() error {
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
	err := model.RegisterType(TypeNode, func() model.Object {
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
				Minimum: 1,
			},
			"version": {
				Type:      "string",
				MinLength: 1,
				MaxLength: maxVersionLength,
			},
		},
		AdditionalProperties: false,
		Required: []string{
			"id",
			"hostname",
			"type",
			"last_ping",
			"version",
		},
	}
	err = generator.Register(string(TypeNode), nodeSchema)
	if err != nil {
		panic(err)
	}
}
