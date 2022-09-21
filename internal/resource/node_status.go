package resource

import (
	"context"
	"fmt"

	nonPublic "github.com/kong/koko/internal/gen/grpc/kong/nonpublic/v1"
	"github.com/kong/koko/internal/model"
	"github.com/kong/koko/internal/model/json/extension"
	"github.com/kong/koko/internal/model/json/generator"
	"github.com/kong/koko/internal/model/json/validation"
	"github.com/kong/koko/internal/model/json/validation/typedefs"
)

const (
	TypeNodeStatus = model.Type("node-status")
)

const (
	maxCompatIssues       = 128
	maxResourcesPerIssue  = 128
	maxLengthResourceType = 32
	compatIssueCodeLength = 4
)

func NewNodeStatus() NodeStatus {
	return NodeStatus{
		NodeStatus: &nonPublic.NodeStatus{},
	}
}

type NodeStatus struct {
	NodeStatus *nonPublic.NodeStatus
}

func (r NodeStatus) ID() string {
	if r.NodeStatus == nil {
		return ""
	}
	return r.NodeStatus.Id
}

func (r NodeStatus) Type() model.Type {
	return TypeNodeStatus
}

func (r NodeStatus) Resource() model.Resource {
	return r.NodeStatus
}

// SetResource implements the Object.SetResource interface.
func (r NodeStatus) SetResource(pr model.Resource) error { return model.SetResource(r, pr) }

func (r NodeStatus) Validate(_ context.Context) error {
	return validation.Validate(string(TypeNodeStatus), r.NodeStatus)
}

func (r NodeStatus) ProcessDefaults(_ context.Context) error {
	if r.NodeStatus == nil {
		return fmt.Errorf("invalid nil resource")
	}
	defaultID(&r.NodeStatus.Id)
	return nil
}

func (r NodeStatus) Indexes() []model.Index {
	return nil
}

func init() {
	err := model.RegisterType(TypeNodeStatus, &nonPublic.NodeStatus{}, func() model.Object {
		return NewNodeStatus()
	})
	if err != nil {
		panic(err)
	}

	nodeSchema := &generator.Schema{
		Type: "object",
		Properties: map[string]*generator.Schema{
			"id":         typedefs.ID,
			"created_at": typedefs.UnixEpoch,
			"updated_at": typedefs.UnixEpoch,
			"issues": {
				Type: "array",
				Items: &generator.Schema{
					Type: "object",
					Properties: map[string]*generator.Schema{
						"code": {
							Type:      "string",
							Pattern:   `^[A-Z][A-Z\d]{3}$`,
							MinLength: compatIssueCodeLength,
							MaxLength: compatIssueCodeLength,
						},
						"affected_resources": {
							Type: "array",
							Items: &generator.Schema{
								Type: "object",
								Properties: map[string]*generator.Schema{
									"id": typedefs.ID,
									"type": {
										Type:      "string",
										MaxLength: maxLengthResourceType,
									},
								},
								Required: []string{
									"id",
									"type",
								},
								AdditionalProperties: &falsy,
							},
							MaxItems: maxResourcesPerIssue,
						},
					},
					Required: []string{
						"code",
					},
					AdditionalProperties: &falsy,
				},
				MaxItems: maxCompatIssues,
			},
		},
		AdditionalProperties: &falsy,
		Required: []string{
			"id",
		},
		XKokoConfig: &extension.Config{
			DisableValidateEndpoint: true,
			ResourceAPIPath:         "",
		},
	}
	err = generator.Registry.Register(string(TypeNodeStatus), nodeSchema)
	if err != nil {
		panic(err)
	}
}
