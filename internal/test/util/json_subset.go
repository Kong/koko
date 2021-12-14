package util

import (
	"fmt"
	"reflect"

	protoJSON "github.com/kong/koko/internal/json"
)

func JSONSubset(o1, o2 interface{}) error {
	m1, err := objectToMap(o1)
	if err != nil {
		return err
	}
	m2, err := objectToMap(o2)
	if err != nil {
		return err
	}
	return subset(m1, m2, "")
}

func subset(o1, o2 interface{}, prefix string) error {
	if reflect.TypeOf(o1) != reflect.TypeOf(o2) {
		return fmt.Errorf("value types are different")
	}
	switch typedO1 := o1.(type) {
	case map[string]interface{}:
		typedO2, ok := o2.(map[string]interface{})
		if !ok {
			panic("unexpected type")
		}
		for k1, v1 := range typedO1 {
			v2, ok := typedO2[k1]
			if !ok {
				return fmt.Errorf("key not present: %v.%v", prefix, k1)
			}
			key := k1
			if prefix != "" {
				key = prefix + "." + key
			}
			err := subset(v1, v2, key)
			if err != nil {
				return err
			}
		}
	case []interface{}:
		typedO2, ok := o2.([]interface{})
		if !ok {
			panic("unexpected type")
		}
		if len(typedO1) != len(typedO2) {
			return fmt.Errorf("length not equal at key(%v) expected: %v, "+
				"got :%v", prefix, len(typedO1), len(typedO2))
		}
		for i, o1Element := range typedO1 {
			var key string
			if prefix != "" {
				key = fmt.Sprintf("%s[%d]", prefix, i)
			} else {
				key = fmt.Sprintf("[%d]", i)
			}
			found := false
			for _, o2Element := range typedO2 {
				err := subset(o1Element, o2Element, key)
				if err == nil {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("can't match element at %v", key)
			}
		}
	default:
		if !reflect.DeepEqual(o1, o2) {
			return fmt.Errorf("values are not equal: %v", prefix)
		}
	}
	return nil
}

func objectToMap(o interface{}) (map[string]interface{}, error) {
	var m map[string]interface{}
	marshal := protoJSON.Marshal
	unmarshal := protoJSON.Unmarshal

	jsonBytes, err := marshal(o)
	if err != nil {
		return nil, fmt.Errorf("marshal value into json: %v", err)
	}
	err = unmarshal(jsonBytes, &m)
	if err != nil {
		return nil, fmt.Errorf("unmarshal value into map: %v", err)
	}
	return m, nil
}
