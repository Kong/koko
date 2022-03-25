package store

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	nonPublic "github.com/kong/koko/internal/gen/grpc/kong/nonpublic/v1"
	"github.com/kong/koko/internal/model"
	"github.com/kong/koko/internal/persistence"
	"github.com/kong/koko/internal/persistence/sqlite"
	"github.com/kong/koko/internal/resource"
	"github.com/kong/koko/internal/store/event"
	"go.uber.org/zap"
)

const DefaultDBQueryTimeout = 5 * time.Second

var (
	errNoObject = fmt.Errorf("no object")
	ErrNotFound = fmt.Errorf("not found")
)

type Store interface {
	Create(context.Context, model.Object, ...CreateOptsFunc) error
	Upsert(context.Context, model.Object, ...CreateOptsFunc) error
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
	if logger == nil {
		panic("logger is required")
	}
	if persister == nil {
		panic("persister is required")
	}
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
	fn func(tx persistence.Tx) error,
) error {
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
	_ ...CreateOptsFunc,
) error {
	ctx, cancel := context.WithTimeout(ctx, DefaultDBQueryTimeout)
	defer cancel()
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
	value, err := wrapObject(object)
	if err != nil {
		return err
	}

	return s.withTx(ctx, func(tx persistence.Tx) error {
		if err := s.createIndexes(ctx, tx, object); err != nil {
			return err
		}
		if err := s.updateEvent(ctx, tx, object); err != nil {
			return err
		}
		return tx.Put(ctx, id, value)
	})
}

func (s *ObjectStore) Upsert(ctx context.Context, object model.Object,
	_ ...CreateOptsFunc,
) error {
	ctx, cancel := context.WithTimeout(ctx, DefaultDBQueryTimeout)
	defer cancel()
	if object == nil {
		return errNoObject
	}
	if err := preProcess(object); err != nil {
		return err
	}

	value, err := wrapObject(object)
	if err != nil {
		return err
	}

	return s.withTx(ctx, func(tx persistence.Tx) error {
		// 1.delete old indexes if they exist
		// no need to delete the object itself since it will be overwritten
		// anyway
		oldObject, err := model.NewObject(object.Type())
		if err != nil {
			return err
		}
		err = s.readByTypeID(ctx, tx, object.Type(), object.ID(), oldObject)
		switch err {
		case nil:
			// TODO(hbagdi): perf: drop and rebuild index only if changed
			// object exists, delete the indexes
			if err := s.deleteIndexes(ctx, tx, oldObject); err != nil {
				return err
			}

		case ErrNotFound:
			// object doesn't exist, move on

		default:
			// some other error
			return err
		}

		// 2. create new indexes
		if err := s.createIndexes(ctx, tx, object); err != nil {
			return err
		}

		// 3. fire off new update event
		if err := s.updateEvent(ctx, tx, object); err != nil {
			return err
		}

		// 4. write the object
		key, err := s.genID(object.Type(), object.ID())
		if err != nil {
			return err
		}
		return tx.Put(ctx, key, value)
	})
}

func (s *ObjectStore) updateEvent(ctx context.Context, tx persistence.Tx,
	object model.Object,
) error {
	if object.Type() == resource.TypeNode {
		return nil
	}
	event := event.Event{
		StoreEvent: &nonPublic.StoreEvent{
			Id:    s.clusterKey(event.ID),
			Value: s.clock(),
		},
	}
	value, err := wrapObject(event)
	if err != nil {
		return fmt.Errorf("proto marshal update event: %v", err)
	}
	id, err := s.genID(event.Type(), event.ID())
	if err != nil {
		return err
	}
	return tx.Put(ctx, id, value)
}

func (s *ObjectStore) clock() string {
	return uuid.NewString()
}

func preProcess(object model.Object) error {
	err := object.ProcessDefaults()
	if err != nil {
		return err
	}
	addTS(object.Resource())

	err = object.Validate()
	if err != nil {
		return err
	}
	return nil
}

func (s *ObjectStore) Read(ctx context.Context, object model.Object,
	opts ...ReadOptsFunc,
) error {
	ctx, cancel := context.WithTimeout(ctx, DefaultDBQueryTimeout)
	defer cancel()
	opt := NewReadOpts(opts...)
	switch {
	case opt.id != "":
		return s.withTx(ctx, func(tx persistence.Tx) error {
			return s.readByTypeID(ctx, tx, object.Type(), opt.id, object)
		})
	case opt.name != "":
		return s.withTx(ctx, func(tx persistence.Tx) error {
			return s.readByName(ctx, tx, opt.name, object)
		})
	case opt.idxName != "" && opt.idxValue != "":
		return s.withTx(ctx, func(tx persistence.Tx) error {
			return s.readByIdx(ctx, tx, opt.idxName, opt.idxValue, object)
		})
	default:
		return fmt.Errorf("invalid opt")
	}
}

func (s *ObjectStore) readByIdx(ctx context.Context, tx persistence.Tx,
	idxName, idxValue string, object model.Object,
) error {
	key := s.uniqueIndexKey(object.Type(), idxName, idxValue)
	value, err := tx.Get(ctx, key)
	if err != nil {
		if errors.As(err, &persistence.ErrNotFound{}) {
			return ErrNotFound
		}
		return err
	}
	refID, err := unwrapUniqueIndex(value)
	if err != nil {
		return err
	}
	err = s.readByTypeID(ctx, tx, object.Type(), refID, object)
	if err != nil {
		return err
	}
	return nil
}

func (s *ObjectStore) readByName(ctx context.Context, tx persistence.Tx,
	name string, object model.Object,
) error {
	return s.readByIdx(ctx, tx, "name", name, object)
}

func (s *ObjectStore) readByTypeID(ctx context.Context, tx persistence.Tx,
	typ model.Type, id string, object model.Object,
) error {
	typeID, err := s.genID(typ, id)
	if err != nil {
		return err
	}
	value, err := tx.Get(ctx, typeID)
	if err != nil {
		if errors.As(err, &persistence.ErrNotFound{}) {
			return ErrNotFound
		}
		return err
	}
	err = unwrapObject(value, object)
	if err != nil {
		return err
	}
	return nil
}

func (s *ObjectStore) Delete(ctx context.Context,
	opts ...DeleteOptsFunc,
) error {
	ctx, cancel := context.WithTimeout(ctx, DefaultDBQueryTimeout)
	defer cancel()
	opt := NewDeleteOpts(opts...)
	return s.withTx(ctx, func(tx persistence.Tx) error {
		return s.delete(ctx, tx, opt.typ, opt.id)
	})
}

func (s *ObjectStore) delete(ctx context.Context, tx persistence.Tx,
	typ model.Type, id string,
) error {
	object, err := model.NewObject(typ)
	if err != nil {
		return err
	}
	err = s.readByTypeID(ctx, tx, typ, id, object)
	if err != nil {
		return err
	}

	err = s.onDeleteCascade(ctx, tx, object)
	if err != nil {
		return err
	}

	key, err := s.genID(typ, id)
	if err != nil {
		return err
	}

	err = tx.Delete(ctx, key)
	if err != nil {
		return err
	}
	if err := s.deleteIndexes(ctx, tx, object); err != nil {
		return err
	}
	return s.updateEvent(ctx, tx, object)
}

func (s *ObjectStore) List(ctx context.Context, list model.ObjectList, opts ...ListOptsFunc) error {
	ctx, cancel := context.WithTimeout(ctx, DefaultDBQueryTimeout)
	defer cancel()
	typ := list.Type()
	opt := NewListOpts(opts...)
	if opt != nil && opt.ReferenceType != "" {
		return s.referencedList(ctx, list, opt)
	}

	listResult, err := s.store.List(ctx, s.listKey(typ), getPersistenceListOptions(opt))
	if err != nil {
		return err
	}
	list.SetTotalCount(listResult.TotalCount)
	lastPage := toLastPage(opt.PageSize, listResult.TotalCount)
	if lastPage > opt.Page {
		list.SetNextPage(opt.Page + 1)
	}
	for _, kv := range listResult.KVList {
		value := kv.Value
		object, err := model.NewObject(typ)
		if err != nil {
			return err
		}

		err = unwrapObject(value, object)
		if err != nil {
			return err
		}
		list.Add(object)
	}
	return nil
}

func (s *ObjectStore) referencedList(ctx context.Context, list model.ObjectList, opt *ListOpts) error {
	typ := list.Type()
	err := s.withTx(ctx, func(tx persistence.Tx) error {
		keyPrefix := s.referencedListKey(typ, opt)
		persistenceOpts := getPersistenceListOptions(opt)
		listResult, err := tx.List(ctx, keyPrefix, persistenceOpts)
		if err != nil {
			return err
		}

		for _, kv := range listResult.KVList {
			key := string(kv.Key)
			value := kv.Value
			if err := verifyForeignValue(value); err != nil {
				panic(err)
			}

			object, err := model.NewObject(typ)
			if err != nil {
				return err
			}
			err = s.readByTypeID(ctx, tx, typ, key[strings.LastIndex(key, "/")+1:], object)
			if err != nil {
				return err
			}
			list.Add(object)
		}
		list.SetTotalCount(listResult.TotalCount)
		lastPage := toLastPage(opt.PageSize, list.GetTotalCount())
		if lastPage > opt.Page {
			list.SetNextPage(opt.Page + 1)
		}
		return nil
	})
	return err
}

func (s *ObjectStore) referencedListKey(typ model.Type, opt *ListOpts) string {
	if opt.ReferenceID != "" {
		return s.clusterKey(fmt.Sprintf("ix/f/%s/%s/%s/",
			opt.ReferenceType, opt.ReferenceID, typ))
	}

	if _, ok := s.store.(*sqlite.SQLite); ok {
		return s.clusterKey(fmt.Sprintf("ix/f/%s/*/%s/", opt.ReferenceType, typ))
	}
	return s.clusterKey(fmt.Sprintf("ix/f/%s/%%/%s/", opt.ReferenceType, typ))
}

func (s *ObjectStore) listKey(typ model.Type) string {
	return s.clusterKey(fmt.Sprintf("o/%s/", typ))
}

func (s *ObjectStore) genID(typ model.Type, id string) (string, error) {
	if id == "" {
		return "", fmt.Errorf("no ID specified")
	}
	if typ == "" {
		return "", fmt.Errorf("no type specified")
	}
	return s.clusterKey(fmt.Sprintf("o/%s/%s", typ, id)), nil
}

func (s *ObjectStore) clusterKey(key string) string {
	if s.cluster == "" {
		panic("cluster not set")
	}
	return fmt.Sprintf("c/%s/%s", s.cluster, key)
}
