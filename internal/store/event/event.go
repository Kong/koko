package event

import (
	v1 "github.com/kong/koko/internal/gen/grpc/kong/nonpublic/v1"
	"github.com/kong/koko/internal/model"
)

const (
	ID = "last_update"
)

var Type = model.Type("store_event")

func New() *Event {
	return &Event{
		StoreEvent: &v1.StoreEvent{Id: ID},
	}
}

type Event struct {
	StoreEvent *v1.StoreEvent
}

func (e Event) ID() string {
	return ID
}

func (e Event) Type() model.Type {
	return Type
}

func (e Event) Resource() model.Resource {
	return e.StoreEvent
}

// SetResource implements the Object.SetResource interface.
func (e Event) SetResource(r model.Resource) error { return model.SetResource(e, r) }

func (e Event) Validate() error {
	return nil
}

func (e Event) Indexes() []model.Index {
	return nil
}

func (e Event) ProcessDefaults() error {
	return nil
}
