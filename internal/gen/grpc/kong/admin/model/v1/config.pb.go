// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.27.1
// 	protoc        (unknown)
// source: kong/admin/model/v1/config.proto

package v1

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type TestingConfig struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Services       []*Service       `protobuf:"bytes,1,rep,name=services,proto3" json:"services,omitempty"`
	Routes         []*Route         `protobuf:"bytes,2,rep,name=routes,proto3" json:"routes,omitempty"`
	Plugins        []*Plugin        `protobuf:"bytes,3,rep,name=plugins,proto3" json:"plugins,omitempty"`
	Upstreams      []*Upstream      `protobuf:"bytes,4,rep,name=upstreams,proto3" json:"upstreams,omitempty"`
	Targets        []*Target        `protobuf:"bytes,5,rep,name=targets,proto3" json:"targets,omitempty"`
	Consumers      []*Consumer      `protobuf:"bytes,6,rep,name=consumers,proto3" json:"consumers,omitempty"`
	Certificates   []*Certificate   `protobuf:"bytes,7,rep,name=certificates,proto3" json:"certificates,omitempty"`
	CaCertificates []*CACertificate `protobuf:"bytes,8,rep,name=ca_certificates,json=caCertificates,proto3" json:"ca_certificates,omitempty"`
}

func (x *TestingConfig) Reset() {
	*x = TestingConfig{}
	if protoimpl.UnsafeEnabled {
		mi := &file_kong_admin_model_v1_config_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TestingConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TestingConfig) ProtoMessage() {}

func (x *TestingConfig) ProtoReflect() protoreflect.Message {
	mi := &file_kong_admin_model_v1_config_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TestingConfig.ProtoReflect.Descriptor instead.
func (*TestingConfig) Descriptor() ([]byte, []int) {
	return file_kong_admin_model_v1_config_proto_rawDescGZIP(), []int{0}
}

func (x *TestingConfig) GetServices() []*Service {
	if x != nil {
		return x.Services
	}
	return nil
}

func (x *TestingConfig) GetRoutes() []*Route {
	if x != nil {
		return x.Routes
	}
	return nil
}

func (x *TestingConfig) GetPlugins() []*Plugin {
	if x != nil {
		return x.Plugins
	}
	return nil
}

func (x *TestingConfig) GetUpstreams() []*Upstream {
	if x != nil {
		return x.Upstreams
	}
	return nil
}

func (x *TestingConfig) GetTargets() []*Target {
	if x != nil {
		return x.Targets
	}
	return nil
}

func (x *TestingConfig) GetConsumers() []*Consumer {
	if x != nil {
		return x.Consumers
	}
	return nil
}

func (x *TestingConfig) GetCertificates() []*Certificate {
	if x != nil {
		return x.Certificates
	}
	return nil
}

func (x *TestingConfig) GetCaCertificates() []*CACertificate {
	if x != nil {
		return x.CaCertificates
	}
	return nil
}

var File_kong_admin_model_v1_config_proto protoreflect.FileDescriptor

var file_kong_admin_model_v1_config_proto_rawDesc = []byte{
	0x0a, 0x20, 0x6b, 0x6f, 0x6e, 0x67, 0x2f, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2f, 0x6d, 0x6f, 0x64,
	0x65, 0x6c, 0x2f, 0x76, 0x31, 0x2f, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x12, 0x13, 0x6b, 0x6f, 0x6e, 0x67, 0x2e, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2e, 0x6d,
	0x6f, 0x64, 0x65, 0x6c, 0x2e, 0x76, 0x31, 0x1a, 0x21, 0x6b, 0x6f, 0x6e, 0x67, 0x2f, 0x61, 0x64,
	0x6d, 0x69, 0x6e, 0x2f, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x2f, 0x76, 0x31, 0x2f, 0x73, 0x65, 0x72,
	0x76, 0x69, 0x63, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1f, 0x6b, 0x6f, 0x6e, 0x67,
	0x2f, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2f, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x2f, 0x76, 0x31, 0x2f,
	0x72, 0x6f, 0x75, 0x74, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x20, 0x6b, 0x6f, 0x6e,
	0x67, 0x2f, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2f, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x2f, 0x76, 0x31,
	0x2f, 0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x22, 0x6b,
	0x6f, 0x6e, 0x67, 0x2f, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2f, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x2f,
	0x76, 0x31, 0x2f, 0x75, 0x70, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x1a, 0x20, 0x6b, 0x6f, 0x6e, 0x67, 0x2f, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2f, 0x6d, 0x6f,
	0x64, 0x65, 0x6c, 0x2f, 0x76, 0x31, 0x2f, 0x74, 0x61, 0x72, 0x67, 0x65, 0x74, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x1a, 0x22, 0x6b, 0x6f, 0x6e, 0x67, 0x2f, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2f,
	0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x2f, 0x76, 0x31, 0x2f, 0x63, 0x6f, 0x6e, 0x73, 0x75, 0x6d, 0x65,
	0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x25, 0x6b, 0x6f, 0x6e, 0x67, 0x2f, 0x61, 0x64,
	0x6d, 0x69, 0x6e, 0x2f, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x2f, 0x76, 0x31, 0x2f, 0x63, 0x65, 0x72,
	0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x28,
	0x6b, 0x6f, 0x6e, 0x67, 0x2f, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2f, 0x6d, 0x6f, 0x64, 0x65, 0x6c,
	0x2f, 0x76, 0x31, 0x2f, 0x63, 0x61, 0x5f, 0x63, 0x65, 0x72, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61,
	0x74, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xf8, 0x03, 0x0a, 0x0d, 0x54, 0x65, 0x73,
	0x74, 0x69, 0x6e, 0x67, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x12, 0x38, 0x0a, 0x08, 0x73, 0x65,
	0x72, 0x76, 0x69, 0x63, 0x65, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1c, 0x2e, 0x6b,
	0x6f, 0x6e, 0x67, 0x2e, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x2e,
	0x76, 0x31, 0x2e, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x52, 0x08, 0x73, 0x65, 0x72, 0x76,
	0x69, 0x63, 0x65, 0x73, 0x12, 0x32, 0x0a, 0x06, 0x72, 0x6f, 0x75, 0x74, 0x65, 0x73, 0x18, 0x02,
	0x20, 0x03, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x6b, 0x6f, 0x6e, 0x67, 0x2e, 0x61, 0x64, 0x6d, 0x69,
	0x6e, 0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x52, 0x6f, 0x75, 0x74, 0x65,
	0x52, 0x06, 0x72, 0x6f, 0x75, 0x74, 0x65, 0x73, 0x12, 0x35, 0x0a, 0x07, 0x70, 0x6c, 0x75, 0x67,
	0x69, 0x6e, 0x73, 0x18, 0x03, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1b, 0x2e, 0x6b, 0x6f, 0x6e, 0x67,
	0x2e, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x2e, 0x76, 0x31, 0x2e,
	0x50, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x52, 0x07, 0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x73, 0x12,
	0x3b, 0x0a, 0x09, 0x75, 0x70, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x73, 0x18, 0x04, 0x20, 0x03,
	0x28, 0x0b, 0x32, 0x1d, 0x2e, 0x6b, 0x6f, 0x6e, 0x67, 0x2e, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2e,
	0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x55, 0x70, 0x73, 0x74, 0x72, 0x65, 0x61,
	0x6d, 0x52, 0x09, 0x75, 0x70, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x73, 0x12, 0x35, 0x0a, 0x07,
	0x74, 0x61, 0x72, 0x67, 0x65, 0x74, 0x73, 0x18, 0x05, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1b, 0x2e,
	0x6b, 0x6f, 0x6e, 0x67, 0x2e, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c,
	0x2e, 0x76, 0x31, 0x2e, 0x54, 0x61, 0x72, 0x67, 0x65, 0x74, 0x52, 0x07, 0x74, 0x61, 0x72, 0x67,
	0x65, 0x74, 0x73, 0x12, 0x3b, 0x0a, 0x09, 0x63, 0x6f, 0x6e, 0x73, 0x75, 0x6d, 0x65, 0x72, 0x73,
	0x18, 0x06, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1d, 0x2e, 0x6b, 0x6f, 0x6e, 0x67, 0x2e, 0x61, 0x64,
	0x6d, 0x69, 0x6e, 0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x43, 0x6f, 0x6e,
	0x73, 0x75, 0x6d, 0x65, 0x72, 0x52, 0x09, 0x63, 0x6f, 0x6e, 0x73, 0x75, 0x6d, 0x65, 0x72, 0x73,
	0x12, 0x44, 0x0a, 0x0c, 0x63, 0x65, 0x72, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x65, 0x73,
	0x18, 0x07, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x20, 0x2e, 0x6b, 0x6f, 0x6e, 0x67, 0x2e, 0x61, 0x64,
	0x6d, 0x69, 0x6e, 0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x43, 0x65, 0x72,
	0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x65, 0x52, 0x0c, 0x63, 0x65, 0x72, 0x74, 0x69, 0x66,
	0x69, 0x63, 0x61, 0x74, 0x65, 0x73, 0x12, 0x4b, 0x0a, 0x0f, 0x63, 0x61, 0x5f, 0x63, 0x65, 0x72,
	0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x65, 0x73, 0x18, 0x08, 0x20, 0x03, 0x28, 0x0b, 0x32,
	0x22, 0x2e, 0x6b, 0x6f, 0x6e, 0x67, 0x2e, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2e, 0x6d, 0x6f, 0x64,
	0x65, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x43, 0x41, 0x43, 0x65, 0x72, 0x74, 0x69, 0x66, 0x69, 0x63,
	0x61, 0x74, 0x65, 0x52, 0x0e, 0x63, 0x61, 0x43, 0x65, 0x72, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61,
	0x74, 0x65, 0x73, 0x42, 0x3f, 0x5a, 0x3d, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f,
	0x6d, 0x2f, 0x6b, 0x6f, 0x6e, 0x67, 0x2f, 0x6b, 0x6f, 0x6b, 0x6f, 0x2f, 0x69, 0x6e, 0x74, 0x65,
	0x72, 0x6e, 0x61, 0x6c, 0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x67, 0x72, 0x70, 0x63, 0x2f, 0x6b, 0x6f,
	0x6e, 0x67, 0x2f, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2f, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x2f, 0x76,
	0x31, 0x3b, 0x76, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_kong_admin_model_v1_config_proto_rawDescOnce sync.Once
	file_kong_admin_model_v1_config_proto_rawDescData = file_kong_admin_model_v1_config_proto_rawDesc
)

func file_kong_admin_model_v1_config_proto_rawDescGZIP() []byte {
	file_kong_admin_model_v1_config_proto_rawDescOnce.Do(func() {
		file_kong_admin_model_v1_config_proto_rawDescData = protoimpl.X.CompressGZIP(file_kong_admin_model_v1_config_proto_rawDescData)
	})
	return file_kong_admin_model_v1_config_proto_rawDescData
}

var file_kong_admin_model_v1_config_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_kong_admin_model_v1_config_proto_goTypes = []interface{}{
	(*TestingConfig)(nil), // 0: kong.admin.model.v1.TestingConfig
	(*Service)(nil),       // 1: kong.admin.model.v1.Service
	(*Route)(nil),         // 2: kong.admin.model.v1.Route
	(*Plugin)(nil),        // 3: kong.admin.model.v1.Plugin
	(*Upstream)(nil),      // 4: kong.admin.model.v1.Upstream
	(*Target)(nil),        // 5: kong.admin.model.v1.Target
	(*Consumer)(nil),      // 6: kong.admin.model.v1.Consumer
	(*Certificate)(nil),   // 7: kong.admin.model.v1.Certificate
	(*CACertificate)(nil), // 8: kong.admin.model.v1.CACertificate
}
var file_kong_admin_model_v1_config_proto_depIdxs = []int32{
	1, // 0: kong.admin.model.v1.TestingConfig.services:type_name -> kong.admin.model.v1.Service
	2, // 1: kong.admin.model.v1.TestingConfig.routes:type_name -> kong.admin.model.v1.Route
	3, // 2: kong.admin.model.v1.TestingConfig.plugins:type_name -> kong.admin.model.v1.Plugin
	4, // 3: kong.admin.model.v1.TestingConfig.upstreams:type_name -> kong.admin.model.v1.Upstream
	5, // 4: kong.admin.model.v1.TestingConfig.targets:type_name -> kong.admin.model.v1.Target
	6, // 5: kong.admin.model.v1.TestingConfig.consumers:type_name -> kong.admin.model.v1.Consumer
	7, // 6: kong.admin.model.v1.TestingConfig.certificates:type_name -> kong.admin.model.v1.Certificate
	8, // 7: kong.admin.model.v1.TestingConfig.ca_certificates:type_name -> kong.admin.model.v1.CACertificate
	8, // [8:8] is the sub-list for method output_type
	8, // [8:8] is the sub-list for method input_type
	8, // [8:8] is the sub-list for extension type_name
	8, // [8:8] is the sub-list for extension extendee
	0, // [0:8] is the sub-list for field type_name
}

func init() { file_kong_admin_model_v1_config_proto_init() }
func file_kong_admin_model_v1_config_proto_init() {
	if File_kong_admin_model_v1_config_proto != nil {
		return
	}
	file_kong_admin_model_v1_service_proto_init()
	file_kong_admin_model_v1_route_proto_init()
	file_kong_admin_model_v1_plugin_proto_init()
	file_kong_admin_model_v1_upstream_proto_init()
	file_kong_admin_model_v1_target_proto_init()
	file_kong_admin_model_v1_consumer_proto_init()
	file_kong_admin_model_v1_certificate_proto_init()
	file_kong_admin_model_v1_ca_certificate_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_kong_admin_model_v1_config_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TestingConfig); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_kong_admin_model_v1_config_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_kong_admin_model_v1_config_proto_goTypes,
		DependencyIndexes: file_kong_admin_model_v1_config_proto_depIdxs,
		MessageInfos:      file_kong_admin_model_v1_config_proto_msgTypes,
	}.Build()
	File_kong_admin_model_v1_config_proto = out.File
	file_kong_admin_model_v1_config_proto_rawDesc = nil
	file_kong_admin_model_v1_config_proto_goTypes = nil
	file_kong_admin_model_v1_config_proto_depIdxs = nil
}
