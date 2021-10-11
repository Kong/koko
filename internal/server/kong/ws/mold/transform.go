package mold

import (
	"fmt"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/kong/koko/internal/gen/wrpc/kong/model"
	"github.com/kong/koko/internal/server/kong/ws/config"
	"google.golang.org/protobuf/encoding/protojson"
)

func GrpcToWrpc(input GrpcContent) (config.Content, error) {
	res := config.Content{
		FormatVersion: "2.1",
	}
	// service ID to service
	services := map[string]*config.Service{}
	for _, service := range input.Services {
		var res model.Service
		err := convert(service, &res)
		if err != nil {
			return config.Content{}, fmt.Errorf("convert service: %v", err)
		}
		services[res.Id] = &config.Service{
			Service: &res,
		}
	}

	for _, route := range input.Routes {
		var res model.Route
		err := convert(route, &res)
		if err != nil {
			return config.Content{}, fmt.Errorf("convert service: %v", err)
		}

		if res.Service != nil && res.Service.Id != "" {
			service := services[res.Service.Id]
			res.Service = nil
			service.Routes = append(service.Routes, &res)
		}
		// TODO(hbagdi): handle service-less routes
	}

	for _, service := range services {
		res.Services = append(res.Services, service)
	}
	return res, nil
}

var jsonpb = runtime.JSONPb{
	MarshalOptions: protojson.MarshalOptions{UseProtoNames: true},
}

// convert uses JSON marshal/unmarshal to convert between types.
// TODO(hbagdi): explore better alternatives using reflect.
func convert(from, to interface{}) error {
	jsonBytes, err := jsonpb.Marshal(from)
	if err != nil {
		return fmt.Errorf("jsonpb marshal: %v", err)
	}
	err = jsonpb.Unmarshal(jsonBytes, to)
	if err != nil {
		return fmt.Errorf("jsonpb unmarshal: %v", err)
	}
	return nil
}
