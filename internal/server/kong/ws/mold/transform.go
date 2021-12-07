package mold

import (
	"fmt"

	"github.com/kong/koko/internal/json"
	"github.com/kong/koko/internal/server/kong/ws/config"
)

type Map map[string]interface{}

func GrpcToWrpc(input GrpcContent) (config.Content, error) {
	res := config.Content{
		FormatVersion: "2.1",
	}

	for _, service := range input.Services {
		m, err := simplify(service)
		if err != nil {
			return config.Content{}, err
		}
		res.Services = append(res.Services, m)
	}
	for _, route := range input.Routes {
		m, err := simplify(route)
		if err != nil {
			return config.Content{}, err
		}
		res.Routes = append(res.Routes, m)
	}

	return res, nil
}

func flatten(m config.Map) {
	flattenForeign(m, "service")
}

func flattenForeign(m config.Map, entityType string) {
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

func simplify(input interface{}) (config.Map, error) {
	var m config.Map
	if err := convert(input, &m); err != nil {
		return nil, err
	}
	flatten(m)
	return m, nil
}

// convert uses JSON marshal/unmarshal to convert between types.
// TODO(hbagdi): explore better alternatives using reflect.
func convert(from, to interface{}) error {
	jsonBytes, err := json.Marshal(from)
	if err != nil {
		return fmt.Errorf("jsonpb marshal: %v", err)
	}
	err = json.Unmarshal(jsonBytes, to)
	if err != nil {
		return fmt.Errorf("jsonpb unmarshal: %v", err)
	}
	return nil
}
