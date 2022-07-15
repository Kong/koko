package store

import (
	"fmt"

	"github.com/kong/koko/internal/json"
	"github.com/kong/koko/internal/model"
)

const (
	valueTypeIndexUnique  = 1
	valueTypeIndexForeign = 2
	valueTypeObject       = 3
)

type valueWrapper struct {
	Type   int             `json:"type,omitempty"`
	Object json.RawMessage `json:"object,omitempty"`
	RefID  string          `json:"ref_id,omitempty"`
}

func wrapUniqueIndex(refID string) ([]byte, error) {
	value, err := json.Marshal(valueWrapper{
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
	err := json.Unmarshal(value, &v)
	if err != nil {
		return "", fmt.Errorf("json unmarshal unique index: %v", err)
	}
	if v.RefID == "" {
		panic("invalid unique index value")
	}
	return v.RefID, nil
}

func wrapForeignIndex() ([]byte, error) {
	value, err := json.Marshal(valueWrapper{
		Type: valueTypeIndexForeign,
	})
	if err != nil {
		return nil, fmt.Errorf("json marshal foreign index: %v", err)
	}
	return value, nil
}

func verifyForeignValue(value []byte) error {
	var indexValue valueWrapper
	err := json.Unmarshal(value, &indexValue)
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
	var jsonObject []byte
	var err error
	if r, ok := object.(model.ObjectWithResourceDTO); ok {
		// The object has implemented its own resource JSON marshaller.
		jsonObject, err = r.MarshalResourceJSON()
	} else {
		// The object is being stored as a JSON-encoded representation of the underlining Protobuf resource.
		jsonObject, err = json.ProtoJSONMarshal(object.Resource())
	}
	if err != nil {
		return nil, err
	}
	value, err := json.Marshal(valueWrapper{
		Type:   valueTypeObject,
		Object: jsonObject,
	})
	if err != nil {
		return nil, fmt.Errorf("json marshal object: %w", err)
	}
	return value, nil
}

func unwrapObject(value []byte, object model.Object) error {
	var wrappedValue valueWrapper

	err := json.Unmarshal(value, &wrappedValue)
	if err != nil {
		return fmt.Errorf("json unmarshal wrapperValue: %w", err)
	}
	if r, ok := object.(model.ObjectWithResourceDTO); ok {
		// The object has implemented its own resource JSON unmarshaller.
		err = r.UnmarshalResourceJSON(wrappedValue.Object)
	} else {
		// The object is being unmarshalled from its JSON-encoded representation of the underlining Protobuf resource.
		err = json.ProtoJSONUnmarshal(wrappedValue.Object, object.Resource())
	}
	if err != nil {
		return fmt.Errorf("json unmarshal object: %w", err)
	}
	return nil
}
