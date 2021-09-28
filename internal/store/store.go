package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/kong/koko/internal/model"
	"github.com/kong/koko/internal/persistence"
	"google.golang.org/protobuf/proto"
)

var (
	errNoObject = fmt.Errorf("no object")
	ErrNotFound = fmt.Errorf("not found")
)

type Store interface {
	Create(context.Context, model.Object, ...CreateOptsFunc) error
	Read(context.Context, model.Object, ...ReadOptsFunc) error
	Delete(context.Context, ...DeleteOptsFunc) error
	List(context.Context, model.ObjectList, ...ListOptsFunc) error
	// TODO(hbagdi): required but something to tackle later on
	// Tx(func(s Store) error) error
}

// ObjectStore stores objects.
// TODO(hbagdi): better name needed between this and the interface.
type ObjectStore struct {
	store persistence.Persister
}

func New(persister persistence.Persister) Store {
	return &ObjectStore{
		store: persister,
	}
}

func (s *ObjectStore) Create(ctx context.Context, object model.Object,
	_ ...CreateOptsFunc) error {
	if object == nil {
		return errNoObject
	}
	if err := preProcess(object); err != nil {
		return err
	}

	id, err := storageKey(object)
	if err != nil {
		return err
	}
	value, err := proto.Marshal(object.Resource())
	if err != nil {
		return err
	}

	err = s.store.Put(ctx, id, value)
	if err != nil {
		return err
	}

	return nil
}

func preProcess(object model.Object) error {
	err := object.ProcessDefaults()
	if err != nil {
		return err
	}
	err = object.Validate()
	if err != nil {
		return err
	}
	return nil
}

func (s *ObjectStore) Read(ctx context.Context, object model.Object,
	opts ...ReadOptsFunc) error {
	opt := NewReadOpts(opts...)
	id, err := genID(object.Type(), opt.id)
	if err != nil {
		return err
	}
	value, err := s.store.Get(ctx, id)
	if err != nil {
		if errors.As(err, &persistence.ErrNotFound{}) {
			return ErrNotFound
		}
		return err
	}
	err = proto.Unmarshal(value, object.Resource())
	if err != nil {
		return err
	}
	return nil
}

func (s *ObjectStore) Delete(ctx context.Context,
	opts ...DeleteOptsFunc) error {
	opt := NewDeleteOpts(opts...)
	id, err := genID(opt.typ, opt.id)
	if err != nil {
		return err
	}
	err = s.store.Delete(ctx, id)
	if err != nil {
		if errors.As(err, &persistence.ErrNotFound{}) {
			return ErrNotFound
		}
		return err
	}
	return nil
}

func (s *ObjectStore) List(ctx context.Context, list model.ObjectList, opts ...ListOptsFunc) error {
	typ := list.Type()
	values, err := s.store.List(ctx, fmt.Sprintf("%s/", typ))
	if err != nil {
		return err
	}
	for _, value := range values {
		object, err := model.NewObject(typ)
		if err != nil {
			return err
		}
		err = proto.Unmarshal(value, object.Resource())
		if err != nil {
			return err
		}
		list.Add(object)
	}
	return nil
}

func storageKey(object model.Object) (string, error) {
	if object == nil {
		return "", fmt.Errorf("no ID specified")
	}
	return genID(object.Type(), object.ID())
}

func genID(typ model.Type, id string) (string, error) {
	if id == "" {
		return "", fmt.Errorf("no ID specified")
	}
	if typ == "" {
		return "", fmt.Errorf("no type specified")
	}
	return fmt.Sprintf("%s/%s", typ, id), nil
}
