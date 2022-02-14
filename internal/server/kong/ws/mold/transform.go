package mold

import (
	"fmt"

	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
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
		translateRouteHeaders(route, m)
		if err != nil {
			return config.Content{}, err
		}
		res.Routes = append(res.Routes, m)
	}
	for _, plugin := range input.Plugins {
		m, err := simplify(plugin)
		delete(m, "updated_at")
		if err != nil {
			return config.Content{}, err
		}
		res.Plugins = append(res.Plugins, m)
	}
	for _, upstream := range input.Upstreams {
		m, err := simplify(upstream)
		delete(m, "updated_at")
		if err != nil {
			return config.Content{}, err
		}
		res.Upstreams = append(res.Upstreams, m)
	}
	for _, target := range input.Targets {
		m, err := simplify(target)
		delete(m, "updated_at")
		if err != nil {
			return config.Content{}, err
		}
		res.Targets = append(res.Targets, m)
	}

	return res, nil
}

func flatten(m config.Map) {
	flattenForeign(m, "service")
	flattenForeign(m, "route")
	flattenForeign(m, "upstream")
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

func translateRouteHeaders(route *v1.Route, m config.Map) {
	if route.Headers == nil {
		return
	}
	res := map[string][]string{}
	for k, v := range route.Headers {
		res[k] = v.Values
	}
	m["headers"] = res
}
