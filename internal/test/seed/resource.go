package seed

import (
	"errors"
	"fmt"

	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	"github.com/kong/koko/internal/model"
	"github.com/kong/koko/internal/model/json/extension"
	"github.com/kong/koko/internal/resource"
	"github.com/kong/koko/internal/test/util"
	"github.com/santhosh-tekuri/jsonschema/v5"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// DefaultNewResourceFuncs contains the NewResourceFunc functions for resources that require
// other fields to be set. Read the documentation on the NewResourceFunc for more info.
//
// This map must not be updated after application initialization (as it is not safe
// to update while a seeder is running). It is exported to allow callers the ability
// to define new functions that piggyback off the default behavior.
//
// We know the resources are of the proper type, so we're disabling the type assertion linter.
// nolint:forcetypeassert
var DefaultNewResourceFuncs = map[model.Type]NewResourceFunc{
	resource.TypeCACertificate: func(_ Seeder, m proto.Message, _ int) error {
		r := m.(*v1.CACertificate)
		var err error
		r.Cert, _, err = util.GenerateCertificate(defaultCertificateBits)
		return err
	},
	resource.TypeCertificate: func(_ Seeder, m proto.Message, _ int) error {
		r := m.(*v1.Certificate)
		var err error
		r.Cert, r.Key, err = util.GenerateCertificate(defaultCertificateBits)
		return err
	},
	resource.TypeConsumer: func(_ Seeder, m proto.Message, i int) error {
		r := m.(*v1.Consumer)
		r.Username = fmt.Sprintf("username-%d", i+1)
		return nil
	},
	resource.TypePlugin: func(s Seeder, m proto.Message, i int) error {
		r := m.(*v1.Plugin)
		r.Name = "key-auth"
		services := s.Results().ByType(resource.TypeService).All()
		if l := len(services); l == 0 {
			return errors.New("must create services before seeding plugins, in order to ensure unique resources")
		} else if l < i+1 {
			return errors.New("must create an equal number of services & plugins")
		}
		r.Service = &v1.Service{Id: services[i].ID}
		return nil
	},
	resource.TypePluginSchema: func(s Seeder, m proto.Message, i int) error {
		r := m.(*v1.PluginSchema)
		r.LuaSchema = fmt.Sprintf(`return {
			name = "%s",
			fields = {
				{ config = {
						type = "record",
						fields = {
							{ field = { type = "string" } }
						}
					}
				}
			}
		}`, fmt.Sprintf("plugin-schema-%d", i+1))
		return nil
	},
	resource.TypeRoute: func(_ Seeder, m proto.Message, i int) error {
		r := m.(*v1.Route)
		r.Name, r.Hosts = fmt.Sprintf("route-%d", i+1), []string{"example.com"}
		return nil
	},
	resource.TypeService: func(_ Seeder, m proto.Message, i int) error {
		r := m.(*v1.Service)
		r.Name, r.Host = fmt.Sprintf("service-%d", i+1), "example.com"
		return nil
	},
	resource.TypeSNI: func(s Seeder, m proto.Message, i int) error {
		r := m.(*v1.SNI)
		r.Name = fmt.Sprintf("example-%d.com", i)
		certificates := s.Results().ByType(resource.TypeCertificate).All()
		if len(certificates) == 0 {
			return errors.New("must create at least one certificate before seeding SNIs")
		}
		r.Certificate = &v1.Certificate{Id: certificates[0].ID}
		return nil
	},
	resource.TypeTarget: func(s Seeder, m proto.Message, i int) error {
		r := m.(*v1.Target)
		r.Target = "127.0.0.1:8080"
		upstreams := s.Results().ByType(resource.TypeUpstream).All()
		if l := len(upstreams); l == 0 {
			return errors.New("must create upstreams before seeding targets, in order to ensure unique resources")
		} else if l < i+1 {
			return errors.New("must create an equal number of upstreams & targets")
		}
		r.Upstream = &v1.Upstream{Id: upstreams[i].ID}
		return nil
	},
	resource.TypeUpstream: func(_ Seeder, m proto.Message, i int) error {
		r := m.(*v1.Upstream)
		r.Name = fmt.Sprintf("upstream-%d", i+1)
		return nil
	},
}

// resourcesToCreateLast defines the resources that require other resources to be created first.
// e.g.: In order to create targets, upstreams must be created first to ensure uniqueness.
//
// While this could technically be computed via the JSON schema, for the sake of simplicity,
// this is being hard-coded.
var resourcesToCreateLast = []model.Type{
	resource.TypeSNI,
	resource.TypePlugin,
	resource.TypeTarget,
}

// Set as low as possible to reduce any performance overhead of certificate generation.
const defaultCertificateBits = 512

// NewResourceFunc allows the seeder to properly set the desired fields on a resource that
// require more fields than just an ID to be created.
//
// The passed in seeder can be used to get IDs of dependent resources.
//
// The passed in integer is the current resource's index in the underling storage. This is
// helpful when the resource requires a dependency, and uniqueness needs to be ensured.
//
// The function must be safe for concurrent use.
type NewResourceFunc func(Seeder, proto.Message, int) error

// ResourceInfo stores various information related to a specific resource.
type ResourceInfo struct {
	// The resource being described.
	Name model.Type

	// Our custom JSON schema extension defining internal config for a resource.
	JSONSchemaConfig *extension.Config

	// The relevant JSON schema for this resource.
	Schema *jsonschema.Schema

	// The resource's underlining model object, with its Protobuf message
	// info. Internally used for cloning & optimizing protoreflect calls.
	object           model.Object
	fieldDescriptors protoreflect.FieldDescriptors
}

// HasField is a helper function to determine if a field exists on the JSON schema.
func (ri *ResourceInfo) HasField(fieldName string) bool {
	_, ok := ri.Schema.Properties[fieldName]
	return ok
}

// createEndpoint returns the POST endpoint used to create the resource.
func (ri *ResourceInfo) createEndpoint() string {
	return fmt.Sprintf("/v1/" + ri.JSONSchemaConfig.ResourceAPIPath)
}
