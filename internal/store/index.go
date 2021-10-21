package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/kong/koko/internal/model"
	"github.com/kong/koko/internal/persistence"
	"go.uber.org/zap"
)

func (s *ObjectStore) indexKV(index model.Index, object model.Object) (string,
	[]byte) {
	if index.Name == "" || index.Value == "" {
		return "", nil
	}
	switch index.Type {
	case model.IndexUnique:
		key := fmt.Sprintf("ix/u/%s/%s/%s", object.Type(), index.Name,
			index.Value)
		value := object.ID()
		return s.clusterKey(key), []byte(value)
	case model.IndexForeign:
		key := fmt.Sprintf("ix/f/%s/%s/%s/%s",
			index.ForeignType, index.Value,
			object.Type(), object.ID())
		value := []byte{'1'}
		return s.clusterKey(key), value
	default:
		panic("invalid index type")
	}
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
	tx persistence.Tx, object model.Object) error {
	indexes := object.Indexes()
	for _, index := range indexes {
		switch index.Type {
		case model.IndexUnique:
			key, value := s.indexKV(index, object)
			if key == "" {
				continue
			}
			err := s.checkIndex(ctx, tx, index, key)
			if err != nil {
				return err
			}

			err = tx.Put(ctx, key, value)
			if err != nil {
				return fmt.Errorf("add '%s(%s)' index for '%s' type", index.Name,
					index.Type, object.Type())
			}
		case model.IndexForeign:
			key, value := s.indexKV(index, object)
			if key == "" {
				continue
			}
			err := s.checkIndex(ctx, tx, index, key)
			if err != nil {
				return err
			}

			err = tx.Put(ctx, key, value)
			if err != nil {
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
					Index:   index,
					Message: "not found",
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
	index model.Index, key string) error {
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
	object model.Object) error {
	indexes := object.Indexes()
	for _, index := range indexes {
		switch index.Type {
		case model.IndexUnique:
			key, _ := s.indexKV(index, object)
			if key == "" {
				continue
			}
			err := tx.Delete(ctx, key)
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
			key, _ := s.indexKV(index, object)
			if key == "" {
				continue
			}
			err := tx.Delete(ctx, key)
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

func (s *ObjectStore) checkForeignIndexesForDelete(ctx context.Context,
	tx persistence.Tx,
	object model.Object) error {
	key := fmt.Sprintf("ix/f/%s/%s", object.Type(), object.ID())
	values, err := tx.List(ctx, s.clusterKey(key))
	if err != nil {
		return err
	}
	if len(values) > 0 {
		return ErrConstraint{Message: "foreign references exist"}
	}
	return nil
}
