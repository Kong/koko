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
	"github.com/kong/koko/internal/store/event"
	"github.com/samber/lo"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

const DefaultOperationTimeout = 15 * time.Second

// DefaultCluster defines the cluster ID when one has not been passed in when instantiating a Store object.
const DefaultCluster = "default"

var (
	errNoObject = fmt.Errorf("no object")
	ErrNotFound = fmt.Errorf("not found")
)

type Store interface {
	// Cluster returns the store's underlining cluster ID. When no specific cluster
	// is associated to the Store, the returned value will be DefaultCluster.
	Cluster() string

	// UpdateForeignKeys allows bulk updates of one-to-many relationships for the provided object.
	//
	// The indexes (defined on the passed in model.Object) that have the type of model.IndexForeign will be
	// created/removed based on the provided model.IndexAction. All other indexes will be ignored.
	//
	// The passed in object must have its ID set. Upon a successful update of the object's foreign key(s), the object
	// will have its `updated_at` timestamp altered and the fully resolved object will be set on the passed in object.
	//
	// ErrNotFound will be returned to the caller in the event the ID set on the object cannot be found or in any of
	// the foreign relations (the `value` field on model.Index).
	UpdateForeignKeys(context.Context, model.Object) error

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

var clusterRegex = regexp.MustCompile(`^[\_\-a-z0-9]{1,64}$`)

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

// Cluster implements the Store interface.
func (s *ObjectStore) Cluster() string {
	if s.cluster != "" {
		return s.cluster
	}
	return DefaultCluster
}

func (s *ObjectStore) Create(ctx context.Context, object model.Object,
	_ ...CreateOptsFunc,
) error {
	ctx, cancel := context.WithTimeout(ctx, DefaultOperationTimeout)
	defer cancel()
	if object == nil {
		return errNoObject
	}
	if err := preProcess(ctx, object); err != nil {
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
		if err := s.checkID(ctx, tx, object); err != nil {
			return err
		}
		if err := s.createIndexes(ctx, tx, object); err != nil {
			return err
		}
		if err := s.updateEvent(ctx, tx, object); err != nil {
			return err
		}
		return tx.Insert(ctx, id, value)
	})
}

func (s *ObjectStore) checkID(ctx context.Context,
	tx persistence.Tx, object model.Object,
) error {
	objectInDB, err := model.NewObject(object.Type())
	if err != nil {
		return err
	}

	err = s.readByTypeID(ctx, tx, object.Type(), object.ID(), objectInDB)
	if err != nil {
		if err == ErrNotFound {
			return nil
		}
		return err
	}
	return ErrConstraint{
		Index: model.Index{
			Name:      "id",
			FieldName: "id",
			Type:      model.IndexUnique,
			Value:     object.ID(),
		},
	}
}

func (s *ObjectStore) Upsert(ctx context.Context, object model.Object,
	_ ...CreateOptsFunc,
) error {
	ctx, cancel := context.WithTimeout(ctx, DefaultOperationTimeout)
	defer cancel()
	if object == nil {
		return errNoObject
	}
	if err := preProcess(ctx, object); err != nil {
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
			oldCreatedAtTS := getCreationTimestamp(oldObject.Resource())
			if oldCreatedAtTS != 0 {
				setCreationTimestamp(object.Resource(), oldCreatedAtTS)
				value, err = wrapObject(object)
				if err != nil {
					return err
				}
			}

			// Handle deleting explicit indexes when asked, otherwise, default to the managed behavior.
			objectForIndexDeletions := oldObject
			if _, ok := model.Indexes(object.Indexes()).Actions()[model.IndexActionRemove]; ok {
				objectForIndexDeletions = object
			}

			if err := s.deleteIndexes(ctx, tx, objectForIndexDeletions, false); err != nil {
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
	// TODO(fero): create function on interface to determine if updateEvent should be ignored.
	// this is a stop gap since no other object currently is required.
	if object.Type() == "node" {
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

func preProcess(ctx context.Context, object model.Object) error {
	if err := object.ProcessDefaults(ctx); err != nil {
		return err
	}
	addTS(object.Resource())

	if err := object.Validate(ctx); err != nil {
		return err
	}

	return model.Indexes(object.Indexes()).Validate()
}

func (s *ObjectStore) Read(ctx context.Context, object model.Object,
	opts ...ReadOptsFunc,
) error {
	ctx, cancel := context.WithTimeout(ctx, DefaultOperationTimeout)
	defer cancel()
	opt := NewReadOpts(opts...)
	switch {
	case opt.id != "":
		return s.readByTypeID(ctx, s.store, object.Type(), opt.id, object)
	case opt.name != "":
		return s.readByName(ctx, s.store, opt.name, object)
	case opt.idxName != "" && opt.idxValue != "":
		return s.readByIdx(ctx, s.store, opt.idxName, opt.idxValue, object)
	default:
		return fmt.Errorf("invalid opt")
	}
}

func (s *ObjectStore) readByIdx(ctx context.Context, tx persistence.CRUD,
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

func (s *ObjectStore) readByName(ctx context.Context, tx persistence.CRUD,
	name string, object model.Object,
) error {
	return s.readByIdx(ctx, tx, "name", name, object)
}

func (s *ObjectStore) readByTypeID(ctx context.Context, tx persistence.CRUD,
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

func (s *ObjectStore) UpdateForeignKeys(
	ctx context.Context,
	obj model.Object,
) error {
	if obj == nil {
		return errNoObject
	}

	if obj.ID() == "" {
		return errors.New("required object ID is not set")
	}

	// We only care about the foreign key indexes with the applicable index actions.
	// Otherwise, we'll ignore everything else as we're not concerned about it.
	fkIndexes := model.Indexes(lo.Filter(obj.Indexes(), func(idx model.Index, _ int) bool {
		if idx.Type != model.IndexForeign {
			return false
		}
		return lo.Contains([]model.IndexAction{model.IndexActionAdd, model.IndexActionRemove}, idx.Action)
	}))
	if len(fkIndexes) == 0 {
		return errors.New("no valid foreign key indexes defined")
	}
	if err := fkIndexes.Validate(); err != nil {
		return err
	}

	if err := s.withTx(ctx, func(tx persistence.Tx) error {
		return s.updateForeignKeysTx(ctx, obj, fkIndexes, tx)
	}); err != nil {
		if errors.As(err, &persistence.ErrNotFound{}) {
			return ErrNotFound
		}
		return err
	}

	return nil
}

func (s *ObjectStore) Delete(ctx context.Context,
	opts ...DeleteOptsFunc,
) error {
	ctx, cancel := context.WithTimeout(ctx, DefaultOperationTimeout)
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
	if err := s.deleteIndexes(ctx, tx, object, true); err != nil {
		return err
	}
	return s.updateEvent(ctx, tx, object)
}

func (s *ObjectStore) List(ctx context.Context, list model.ObjectList, opts ...ListOptsFunc) error {
	ctx, cancel := context.WithTimeout(ctx, DefaultOperationTimeout)
	defer cancel()
	typ := list.Type()
	opt, err := NewListOpts(opts...)
	if err != nil {
		return err
	}
	if opt != nil && opt.ReferenceType != "" && opt.ReferenceID != "" {
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
	keyPrefix := s.referencedListKey(typ, opt)
	persistenceOpts := getPersistenceListOptions(opt)
	listResult, err := s.store.List(ctx, keyPrefix, persistenceOpts)
	if err != nil {
		return err
	}

	keyPrefixLen := len(keyPrefix)
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

		// When `opt.ReferenceReverseLookup` is not true, we're extracting <id> from the below key:
		// `c/.../ix/f/<typ>/<opt.ReferenceID>/<opt.ReferenceType>/<id>`
		//
		// When it is true, we're extracting the <id> from the below key:
		// `c/.../ix/f/<typ>/<id>/<opt.ReferenceType>/<opt.ReferenceID>`
		id := key[keyPrefixLen:]
		if opt.ReferenceReverseLookup {
			startIdx := strings.Index(keyPrefix, persistence.WildcardOperator)
			if startIdx < 0 {
				return errors.New("unable to find wildcard operator in key prefix")
			}
			id, _, _ = strings.Cut(key[startIdx:], "/")
		}

		if err := s.readByTypeID(ctx, s.store, typ, id, object); err != nil {
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
}

func (s *ObjectStore) updateForeignKeysTx(
	ctx context.Context,
	obj model.Object,
	indexes model.Indexes,
	tx persistence.Tx,
) error {
	// Ensure the "parent" entity exists. This is also used to resolve the full object for the caller's use.
	if err := s.readByTypeID(ctx, tx, obj.Type(), obj.ID(), obj); err != nil {
		return err
	}

	// Create/remove foreign keys as requested. This is handled asynchronously
	// as bulk management is supported & this should be a performant call.
	g, gCtx := errgroup.WithContext(ctx)
	for i := range indexes {
		index := indexes[i]
		g.Go(func() error {
			key, value, err := s.indexKV(index, obj)
			if err != nil {
				return fmt.Errorf("unable to render indexes: %w", err)
			}

			err = func() error {
				if index.Action == model.IndexActionAdd {
					// Ensure the foreign entity exists.
					fk, err := s.genID(index.ForeignType, index.Value)
					if err != nil {
						return err
					}
					if _, err = tx.Get(gCtx, fk); err != nil {
						return err
					}

					return tx.Insert(gCtx, key, value)
				}

				return tx.Delete(gCtx, key)
			}()

			if err == persistence.ErrUniqueViolation {
				return ErrConstraint{Index: index}
			}

			return err
		})
	}

	if err := g.Wait(); err != nil {
		return err
	}

	if err := s.updateEvent(ctx, tx, obj); err != nil {
		return err
	}

	// Update "parent" entity only so we can update the `updated_at` timestamp.
	setTSField(obj.Resource(), fieldUpdatedAt, true)
	id, err := s.genID(obj.Type(), obj.ID())
	if err != nil {
		return err
	}
	value, err := wrapObject(obj)
	if err != nil {
		return err
	}
	return tx.Put(ctx, id, value)
}

func (s *ObjectStore) referencedListKey(typ model.Type, opt *ListOpts) string {
	parts := []string{string(opt.ReferenceType), opt.ReferenceID, string(typ)}
	suffix := "/"
	if opt.ReferenceReverseLookup {
		parts = []string{string(typ), persistence.WildcardOperator, string(opt.ReferenceType), opt.ReferenceID}
		suffix = ""
	}
	return s.clusterKey(strings.Join(append([]string{"ix", "f"}, parts...), "/")) + suffix
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
