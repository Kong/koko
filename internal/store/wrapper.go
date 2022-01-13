package store

import (
	"fmt"

	protoJSON "github.com/kong/koko/internal/json"
	"github.com/kong/koko/internal/model"
)

const (
	valueTypeIndexUnique  = 1
	valueTypeIndexForeign = 2
	valueTypeObject       = 3
)

type valueWrapper struct {
	Type   int         `json:"type,omitempty"`
	Object interface{} `json:"object,omitempty"`
	RefID  string      `json:"ref_id,omitempty"`
}

func wrapUniqueIndex(refID string) ([]byte, error) {
	value, err := protoJSON.Marshal(valueWrapper{
		Type:  valueTypeIndexUnique,
		RefID: refID,
	})
	if err != nil {
		return nil, fmt.Errorf("json marshal unique index: %v", err)
	}
	return value, nil
}

func wrapForeignIndex() ([]byte, error) {
	value, err := protoJSON.Marshal(valueWrapper{
		Type: valueTypeIndexForeign,
	})
	if err != nil {
		return nil, fmt.Errorf("json marshal foreign index: %v", err)
	}
	return value, nil
}

func verifyForeignValue(value []byte) error {
	var indexValue valueWrapper
	err := protoJSON.Unmarshal(value, &indexValue)
	if err != nil {
		return err
	}
	if indexValue.Type != valueTypeIndexForeign {
		return fmt.Errorf("invalid index type for foreign index: '%v'",
			indexValue.Type)
	}
	return nil
}

func wrapObject(object model.Object) ([]byte, error) {
	value, err := protoJSON.Marshal(valueWrapper{
		Type:   valueTypeObject,
		Object: object.Resource(),
	})
	if err != nil {
		return nil, fmt.Errorf("json marshal object: %v", err)
	}
	return value, nil
}

func unwrapObject(value []byte, object model.Object) error {
	var wrappedValue valueWrapper
	wrappedValue.Object = object.Resource()

	err := protoJSON.Unmarshal(value, &wrappedValue)
	if err != nil {
		return fmt.Errorf("json unmarshal object: %v", err)
	}
	return nil
}
