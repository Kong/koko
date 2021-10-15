package store

import (
	"context"
	"errors"
	"fmt"
	"regexp"

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
}

type objectStoreOpts struct {
	logger *zap.Logger
	store  persistence.Persister
}

// ObjectStore stores objects.
// TODO(hbagdi): better name needed between this and the interface.
type ObjectStore struct {
	cluster string
	objectStoreOpts
}

func New(persister persistence.Persister, logger *zap.Logger) *ObjectStore {
	return &ObjectStore{
		objectStoreOpts: objectStoreOpts{
			logger: logger,
			store:  persister,
		},
	}
}

var clusterRegex = regexp.MustCompile("^[_a-z0-9]{1,64}$")

func (s *ObjectStore) ForCluster(cluster string) *ObjectStore {
	if !clusterRegex.MatchString(cluster) {
		panic(fmt.Sprintf("unexpected cluster identifier: %v", cluster))
	}

	return &ObjectStore{
		objectStoreOpts: s.objectStoreOpts,
		cluster:         cluster,
	}
}

func (s *ObjectStore) withTx(ctx context.Context,
	fn func(tx persistence.Tx) error) error {
	tx, err := s.store.Tx(ctx)
	if err != nil {
		return err
	}
	err = fn(tx)
	if err != nil {
		rollbackerr := tx.Rollback()
		if rollbackerr != nil {
			return rollbackerr
		}
		return err
	}
	return tx.Commit()
}

func (s *ObjectStore) Create(ctx context.Context, object model.Object,
	_ ...CreateOptsFunc) error {
	if object == nil {
		return errNoObject
	}
	if err := preProcess(object); err != nil {
		return err
	}

	id, err := s.genID(object.Type(), object.ID())
	if err != nil {
		return err
	}
	value, err := proto.Marshal(object.Resource())
	if err != nil {
		return err
	}

	return s.withTx(ctx, func(tx persistence.Tx) error {
		if err := s.createIndexes(ctx, tx, object); err != nil {
			return err
		}
		return tx.Put(ctx, id, value)
	})
}

func preProcess(object model.Object) error {
	err := object.ProcessDefaults()
	if err != nil {
		return err
	}
	err = object.ValidateCompat()
	if err != nil {
		return err
	}
	return nil
}

func (s *ObjectStore) Read(ctx context.Context, object model.Object,
	opts ...ReadOptsFunc) error {
	opt := NewReadOpts(opts...)
	id, err := s.genID(object.Type(), opt.id)
	if err != nil {
		return err
	}
	return s.withTx(ctx, func(tx persistence.Tx) error {
		return s.readByTypeID(ctx, tx, id, object)
	})
}

func (s *ObjectStore) readByTypeID(ctx context.Context, tx persistence.Tx,
	typeID string, object model.Object) error {
	value, err := tx.Get(ctx, typeID)
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
	id, err := s.genID(opt.typ, opt.id)
	if err != nil {
		return err
	}
	object, err := model.NewObject(opt.typ)
	if err != nil {
		return err
	}
	return s.withTx(ctx, func(tx persistence.Tx) error {
		err = s.readByTypeID(ctx, tx, id, object)
		if err != nil {
			return err
		}
		err = tx.Delete(ctx, id)
		if err != nil {
			if errors.As(err, &persistence.ErrNotFound{}) {
				return ErrNotFound
			}
			return err
		}
		return s.deleteIndexes(ctx, tx, object)
	})
}

func (s *ObjectStore) List(ctx context.Context, list model.ObjectList, opts ...ListOptsFunc) error {
	typ := list.Type()
	values, err := s.store.List(ctx, s.listKey(typ))
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

func (s *ObjectStore) listKey(typ model.Type) string {
	return s.clusterKey(fmt.Sprintf("%s/", typ))
}

func (s *ObjectStore) genID(typ model.Type, id string) (string, error) {
	if id == "" {
		return "", fmt.Errorf("no ID specified")
	}
	if typ == "" {
		return "", fmt.Errorf("no type specified")
	}
	return s.clusterKey(fmt.Sprintf("%s/%s", typ, id)), nil
}

func (s *ObjectStore) clusterKey(key string) string {
	if s.cluster == "" {
		panic("cluster not set")
	}
	return fmt.Sprintf("c/%s/%s", s.cluster, key)
}
