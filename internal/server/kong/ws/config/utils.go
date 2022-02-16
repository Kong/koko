package config

import (
	"fmt"
	"time"

	protoJSON "github.com/kong/koko/internal/json"
)

var (
	defaultRequestTimeout       = 30 * time.Second
	pageSize              int32 = 1000
)

func flattenForeign(m Map, entityType string) {
	if _, ok := m[entityType]; !ok {
		return
	}
	entity, ok := m[entityType].(map[string]interface{})
	if !ok {
		panic(fmt.Sprintf("'%s' key is not a JSON object ("+
			"map[string]interface{}", entityType))
	}
	if _, ok := entity["id"]; !ok {
		panic(fmt.Sprintf("'%s.id' not found within a foreign relation", entityType))
	}
	m[entityType] = entity["id"]
}

// convert uses JSON marshal/unmarshal to convert between types.
// TODO(hbagdi): explore better alternatives using reflect.
func convert(from interface{}) (Map, error) {
	var m Map
	jsonBytes, err := protoJSON.Marshal(from)
	if err != nil {
		return nil, fmt.Errorf("jsonpb marshal: %v", err)
	}
	err = protoJSON.Unmarshal(jsonBytes, &m)
	if err != nil {
		return nil, fmt.Errorf("jsonpb unmarshal: %v", err)
	}
	return m, nil
}
