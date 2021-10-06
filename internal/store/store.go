package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/kong/koko/internal/model"
	"github.com/kong/koko/internal/persistence"
	"go.uber.org/zap"
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
	logger *zap.Logger
	store  persistence.Persister
}

func New(persister persistence.Persister, logger *zap.Logger) Store {
	return &ObjectStore{
		logger: logger,
		store:  persister,
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

	if err := s.createIndexes(ctx, object); err != nil {
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
	return s.readByTypeID(ctx, id, object)
}

func (s *ObjectStore) readByTypeID(ctx context.Context, typeID string,
	object model.Object) error {
	value, err := s.store.Get(ctx, typeID)
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
	object, err := model.NewObject(opt.typ)
	if err != nil {
		return err
	}
	err = s.readByTypeID(ctx, id, object)
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
	err = s.deleteIndexes(ctx, object)
	if err != nil {
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
