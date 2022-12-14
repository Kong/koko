package store

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/kong/koko/internal/model"
	"github.com/kong/koko/internal/persistence"
	"go.uber.org/zap"
)

func (s *ObjectStore) uniqueIndexKey(typ model.Type, indexName, indexValue string) string {
	if typ == "" || indexName == "" || indexValue == "" {
		panic("unique index must have a typ, name and a value")
	}
	key := fmt.Sprintf("ix/u/%s/%s/%s", typ, indexName,
		indexValue)
	return s.clusterKey(key)
}

func (s *ObjectStore) foreignIndexKey(foreignType model.Type,
	foreignValue string, objectType model.Type, objectID string,
) string {
	if foreignType == "" || foreignValue == "" ||
		objectType == "" || objectID == "" {
		panic("foreign index with invalid values")
	}
	key := fmt.Sprintf("ix/f/%s/%s/%s/%s",
		foreignType, foreignValue,
		objectType, objectID)
	return s.clusterKey(key)
}

func (s *ObjectStore) indexKV(index model.Index, object model.Object) (string,
	[]byte, error,
) {
	switch index.Type {
	case model.IndexUnique:
		key := s.uniqueIndexKey(object.Type(), index.Name, index.Value)
		value, err := wrapUniqueIndex(object.ID())
		if err != nil {
			return "", nil, err
		}

		return key, value, nil
	case model.IndexForeign:
		key := s.foreignIndexKey(index.ForeignType, index.Value,
			object.Type(), object.ID())
		value, err := wrapForeignIndex()
		if err != nil {
			return "", nil, err
		}
		return key, value, nil
	}

	return "", nil, errors.New("invalid index type")
}

type ErrConstraint struct {
	Index   model.Index
	Message string
}

func (e ErrConstraint) Error() string {
	return fmt.Sprintf("%s (type: %s) constraint failed for value '%s': %s",
		e.Index.Name, e.Index.Type, e.Index.Value, e.Message)
}

func (s *ObjectStore) createIndexes(ctx context.Context,
	tx persistence.Tx, object model.Object,
) error {
	indexes := object.Indexes()
	for _, index := range indexes {
		// Skip indexes that are meant for explicit removal only (these are
		// only created with the model.IndexActionAdd index action).
		if index.Action == model.IndexActionRemove {
			continue
		}

		switch index.Type {
		case model.IndexUnique:
			key, value, err := s.indexKV(index, object)
			if err != nil {
				return fmt.Errorf("unable to render indexes: %w", err)
			}

			err = s.checkIndex(ctx, tx, index, key)
			if err != nil {
				return err
			}

			err = tx.Insert(ctx, key, value)
			if err != nil {
				if err == persistence.ErrUniqueViolation {
					return ErrConstraint{
						Index: index,
					}
				}
				return fmt.Errorf("add '%s(%s)' index for '%s' type", index.Name,
					index.Type, object.Type())
			}
		case model.IndexForeign:
			key, value, err := s.indexKV(index, object)
			if err != nil {
				return fmt.Errorf("unable to render indexes: %w", err)
			}

			err = s.checkIndex(ctx, tx, index, key)
			if err != nil {
				return err
			}

			err = tx.Insert(ctx, key, value)
			if err != nil {
				if err == persistence.ErrUniqueViolation {
					return ErrConstraint{
						Index: index,
					}
				}
				return err
			}

			// check if the foreign entity exists or not
			fk, err := s.genID(index.ForeignType, index.Value)
			if err != nil {
				return err
			}
			_, err = tx.Get(ctx, fk)
			switch {
			case err == nil:
				// happy path
			case errors.As(err, &persistence.ErrNotFound{}):
				return ErrConstraint{
					Index: index,
				}
			default:
				// some other problem
				return err
			}
		default:
			panic("invalid index type")
		}
	}
	return nil
}

func (s *ObjectStore) checkIndex(ctx context.Context, tx persistence.Tx,
	index model.Index, key string,
) error {
	_, err := tx.Get(ctx, key)
	switch {
	case err == nil:
		// found the key, unique constraint violation
		return ErrConstraint{
			Index: index,
		}
	case errors.As(err, &persistence.ErrNotFound{}):
		// happy path, continue ahead
		return nil
	default:
		// some other problem
		return err
	}
}

func (s *ObjectStore) deleteIndexes(ctx context.Context, tx persistence.Tx,
	object model.Object,
) error {
	indexes := object.Indexes()
	for _, index := range indexes {
		// Skip indexes that are meant for explicit addition only (these are
		// only removed with the model.IndexActionRemove index action).
		if index.Action == model.IndexActionAdd {
			continue
		}

		switch index.Type {
		case model.IndexUnique:
			key, _, err := s.indexKV(index, object)
			if err != nil {
				return fmt.Errorf("unable to render indexes: %w", err)
			}
			err = tx.Delete(ctx, key)
			if err != nil {
				s.logger.With(
					zap.Error(err),
					zap.String("index_name", index.Name),
					zap.String("index_type", string(index.Type)),
					zap.String("object_type", string(object.Type())),
					zap.String("object_id", object.ID()),
				).Error("delete index failed, possible data integrity issue")
			}
		case model.IndexForeign:
			key, _, err := s.indexKV(index, object)
			if err != nil {
				return fmt.Errorf("unable to render indexes: %w", err)
			}
			err = tx.Delete(ctx, key)
			if err != nil {
				s.logger.With(
					zap.Error(err),
					zap.String("index_name", index.Name),
					zap.String("index_type", string(index.Type)),
					zap.String("object_type", string(object.Type())),
					zap.String("object_id", object.ID()),
				).Error("delete index failed, possible data integrity issue")
			}
		default:
			panic("invalid index type")
		}
	}
	return nil
}

func (s *ObjectStore) onDeleteCascade(ctx context.Context,
	tx persistence.Tx, object model.Object,
) error {
	key := s.clusterKey(fmt.Sprintf("ix/f/%s/%s", object.Type(), object.ID()))
	listResult, err := getFullList(ctx, tx, key)
	if err != nil {
		return err
	}
	if len(listResult.KVList) > 0 {
		for _, ref := range listResult.KVList {
			refTypeID := strings.TrimPrefix(string(ref.Key), key+"/")
			typeAndID := strings.Split(refTypeID, "/")
			typ := model.Type(typeAndID[0])
			id := typeAndID[1]

			// Handle deletion of the foreign key relation, but skip cascading
			// the entire object whenever applicable.
			//
			// This is useful for one-to-many relationships, like what is used
			// for consumer groups. As when a consumer is deleted, we'll want
			// to delete the association, but not the consumer group itself.
			if !model.OptionsForType(typ).CascadeOnDelete {
				if err := tx.Delete(ctx, string(ref.Key)); err != nil {
					return err
				}
				continue
			}

			// Handle deletion of the foreign key relation along with deleting
			// the entire object.
			//
			// For example, when a consumer is deleted that has route(s)
			// associated to it, those routes will be entirely deleted.
			if err := s.delete(ctx, tx, typ, id); err != nil {
				return err
			}
		}
	}
	return nil
}
