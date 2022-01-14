package store

import (
	encodingJSON "encoding/json"
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
	Type   int                     `json:"type,omitempty"`
	Object encodingJSON.RawMessage `json:"object,omitempty"`
	RefID  string                  `json:"ref_id,omitempty"`
}

func wrapUniqueIndex(refID string) ([]byte, error) {
	value, err := encodingJSON.Marshal(valueWrapper{
		Type:  valueTypeIndexUnique,
		RefID: refID,
	})
	if err != nil {
		return nil, fmt.Errorf("json marshal unique index: %v", err)
	}
	return value, nil
}

func unwrapUniqueIndex(value []byte) (string, error) {
	var v valueWrapper
	err := encodingJSON.Unmarshal(value, &v)
	if err != nil {
		return "", fmt.Errorf("json unmarshal unique index: %v", err)
	}
	if v.RefID == "" {
		panic("invalid unique index value")
	}
	return v.RefID, nil
}

func wrapForeignIndex() ([]byte, error) {
	value, err := encodingJSON.Marshal(valueWrapper{
		Type: valueTypeIndexForeign,
	})
	if err != nil {
		return nil, fmt.Errorf("json marshal foreign index: %v", err)
	}
	return value, nil
}

func verifyForeignValue(value []byte) error {
	var indexValue valueWrapper
	err := encodingJSON.Unmarshal(value, &indexValue)
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
	jsonObject, err := protoJSON.Marshal(object.Resource())
	if err != nil {
		return nil, err
	}
	value, err := encodingJSON.Marshal(valueWrapper{
		Type:   valueTypeObject,
		Object: jsonObject,
	})
	if err != nil {
		return nil, fmt.Errorf("json marshal object: %v", err)
	}
	return value, nil
}

func unwrapObject(value []byte, object model.Object) error {
	var wrappedValue valueWrapper

	err := encodingJSON.Unmarshal(value, &wrappedValue)
	if err != nil {
		return fmt.Errorf("json unmarshal wrapperValue: %v", err)
	}
	err = protoJSON.Unmarshal(wrappedValue.Object, object.Resource())
	if err != nil {
		return fmt.Errorf("json unmarshal object: %v", err)
	}
	return nil
}
