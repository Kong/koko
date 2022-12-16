package resource

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	"github.com/kong/koko/internal/model"
	"github.com/kong/koko/internal/model/json/generator"
	"github.com/kong/koko/internal/model/json/validation"
	"github.com/kong/koko/internal/model/json/validation/typedefs"
	"github.com/samber/lo"
)

const (
	TypeConsumerGroup model.Type = "consumer_group"

	ConsumerGroupRuleTitle = "consumer_group_rule"
)

func NewConsumerGroup() ConsumerGroup {
	return ConsumerGroup{
		ConsumerGroup: &v1.ConsumerGroup{},
	}
}

type ConsumerGroup struct {
	ConsumerGroup *v1.ConsumerGroup

	// MemberIDsToAdd defines the consumer IDs to associate to this consumer group.
	MemberIDsToAdd []string

	// MemberIDsToRemove defines the consumer IDs to remove from this consumer group.
	MemberIDsToRemove []string
}

func (c ConsumerGroup) ID() string {
	if c.ConsumerGroup == nil {
		return ""
	}
	return c.ConsumerGroup.Id
}

func (c ConsumerGroup) Type() model.Type {
	return TypeConsumerGroup
}

func (c ConsumerGroup) Resource() model.Resource {
	return c.ConsumerGroup
}

// SetResource implements the Object.SetResource interface.
func (c ConsumerGroup) SetResource(pr model.Resource) error { return model.SetResource(c, pr) }

func (c ConsumerGroup) Validate(ctx context.Context) error {
	var validationErr validation.Error
	for _, memberID := range lo.Union(c.MemberIDsToAdd, c.MemberIDsToRemove) {
		if _, err := uuid.Parse(memberID); err != nil {
			validationErr.Errs = append(validationErr.Errs, &v1.ErrorDetail{
				Type:     v1.ErrorType_ERROR_TYPE_FIELD,
				Field:    "consumer_id",
				Messages: []string{typedefs.ID.Description},
			})
		}
	}
	if len(validationErr.Errs) > 0 {
		return validationErr
	}

	return validation.Validate(string(TypeConsumerGroup), c.ConsumerGroup)
}

func (c ConsumerGroup) ProcessDefaults(ctx context.Context) error {
	if c.ConsumerGroup == nil {
		return fmt.Errorf("invalid nil resource")
	}
	defaultID(&c.ConsumerGroup.Id)
	return nil
}

func (c ConsumerGroup) Indexes() []model.Index {
	// If we're adding/removing members, only return those indexes, as
	// these are managed outside the persistence store integration.
	var memberForeignKeys []model.Index
	for i, memberIDs := range [][]string{c.MemberIDsToAdd, c.MemberIDsToRemove} {
		for _, memberID := range memberIDs {
			memberForeignKeys = append(memberForeignKeys, model.Index{
				Name:        "consumer_id",
				Type:        model.IndexForeign,
				ForeignType: TypeConsumer,
				FieldName:   "consumer.id",
				Value:       memberID,
				Action:      lo.Ternary(i == 0, model.IndexActionAdd, model.IndexActionRemove),
			})
		}
	}
	if len(memberForeignKeys) > 0 {
		return memberForeignKeys
	}

	return []model.Index{{
		Name:      "name",
		Type:      model.IndexUnique,
		Value:     c.ConsumerGroup.Name,
		FieldName: "name",
	}}
}

// Options implements the model.ObjectWithOptions interface.
func (c ConsumerGroup) Options() model.ObjectOptions {
	return model.ObjectOptions{
		// This is a one-to-many entity, and as such, regardless if the last member is
		// deleted from the consumer group, we'll want to keep the consumer group around.
		CascadeOnDelete: false,
	}
}

func init() {
	err := model.RegisterType(TypeConsumerGroup, &v1.ConsumerGroup{}, func() model.Object {
		return NewConsumerGroup()
	})
	if err != nil {
		panic(err)
	}

	consumerGroupSchema := &generator.Schema{
		AdditionalProperties: &falsy,
		Properties: map[string]*generator.Schema{
			"id":         typedefs.ID,
			"name":       typedefs.Name,
			"tags":       typedefs.Tags,
			"created_at": typedefs.UnixEpoch,
			"updated_at": typedefs.UnixEpoch,
		},
		Required: []string{
			"id",
			"name",
		},
		AllOf: []*generator.Schema{
			{
				Title: ConsumerGroupRuleTitle,
				Description: "Consumer Groups are a Kong Enterprise-only feature. " +
					"Please upgrade to Kong Enterprise to use this feature.",
				Not: &generator.Schema{
					Required: []string{"name"},
				},
			},
		},
	}
	err = generator.DefaultRegistry.Register(string(TypeConsumerGroup), consumerGroupSchema)
	if err != nil {
		panic(err)
	}
}
