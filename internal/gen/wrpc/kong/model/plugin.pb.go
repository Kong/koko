// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.27.1
// 	protoc        v3.17.3
// source: kong/model/plugin.proto

package model

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	structpb "google.golang.org/protobuf/types/known/structpb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type Plugin struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id        string           `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Name      string           `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	Config    *structpb.Struct `protobuf:"bytes,3,opt,name=config,proto3" json:"config,omitempty"`
	Enabled   bool             `protobuf:"varint,4,opt,name=enabled,proto3" json:"enabled,omitempty"`
	Protocols []string         `protobuf:"bytes,5,rep,name=protocols,proto3" json:"protocols,omitempty"`
	Tags      []string         `protobuf:"bytes,6,rep,name=tags,proto3" json:"tags,omitempty"`
	CreatedAt int64            `protobuf:"varint,7,opt,name=created_at,json=createdAt,proto3" json:"created_at,omitempty"`
	Route     *Route           `protobuf:"bytes,8,opt,name=route,proto3" json:"route,omitempty"`
	Service   *Service         `protobuf:"bytes,9,opt,name=service,proto3" json:"service,omitempty"`
	Consumer  *Consumer        `protobuf:"bytes,10,opt,name=consumer,proto3" json:"consumer,omitempty"`
}

func (x *Plugin) Reset() {
	*x = Plugin{}
	if protoimpl.UnsafeEnabled {
		mi := &file_kong_model_plugin_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Plugin) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Plugin) ProtoMessage() {}

func (x *Plugin) ProtoReflect() protoreflect.Message {
	mi := &file_kong_model_plugin_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Plugin.ProtoReflect.Descriptor instead.
func (*Plugin) Descriptor() ([]byte, []int) {
	return file_kong_model_plugin_proto_rawDescGZIP(), []int{0}
}

func (x *Plugin) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *Plugin) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Plugin) GetConfig() *structpb.Struct {
	if x != nil {
		return x.Config
	}
	return nil
}

func (x *Plugin) GetEnabled() bool {
	if x != nil {
		return x.Enabled
	}
	return false
}

func (x *Plugin) GetProtocols() []string {
	if x != nil {
		return x.Protocols
	}
	return nil
}

func (x *Plugin) GetTags() []string {
	if x != nil {
		return x.Tags
	}
	return nil
}

func (x *Plugin) GetCreatedAt() int64 {
	if x != nil {
		return x.CreatedAt
	}
	return 0
}

func (x *Plugin) GetRoute() *Route {
	if x != nil {
		return x.Route
	}
	return nil
}

func (x *Plugin) GetService() *Service {
	if x != nil {
		return x.Service
	}
	return nil
}

func (x *Plugin) GetConsumer() *Consumer {
	if x != nil {
		return x.Consumer
	}
	return nil
}

var File_kong_model_plugin_proto protoreflect.FileDescriptor

var file_kong_model_plugin_proto_rawDesc = []byte{
	0x0a, 0x17, 0x6b, 0x6f, 0x6e, 0x67, 0x2f, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x2f, 0x70, 0x6c, 0x75,
	0x67, 0x69, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0a, 0x6b, 0x6f, 0x6e, 0x67, 0x2e,
	0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x1a, 0x1c, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x73, 0x74, 0x72, 0x75, 0x63, 0x74, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x1a, 0x19, 0x6b, 0x6f, 0x6e, 0x67, 0x2f, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x2f,
	0x63, 0x6f, 0x6e, 0x73, 0x75, 0x6d, 0x65, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x18,
	0x6b, 0x6f, 0x6e, 0x67, 0x2f, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x2f, 0x73, 0x65, 0x72, 0x76, 0x69,
	0x63, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x16, 0x6b, 0x6f, 0x6e, 0x67, 0x2f, 0x6d,
	0x6f, 0x64, 0x65, 0x6c, 0x2f, 0x72, 0x6f, 0x75, 0x74, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x22, 0xd2, 0x02, 0x0a, 0x06, 0x50, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x12, 0x0e, 0x0a, 0x02, 0x69,
	0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x6e,
	0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12,
	0x2f, 0x0a, 0x06, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x17, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75,
	0x66, 0x2e, 0x53, 0x74, 0x72, 0x75, 0x63, 0x74, 0x52, 0x06, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67,
	0x12, 0x18, 0x0a, 0x07, 0x65, 0x6e, 0x61, 0x62, 0x6c, 0x65, 0x64, 0x18, 0x04, 0x20, 0x01, 0x28,
	0x08, 0x52, 0x07, 0x65, 0x6e, 0x61, 0x62, 0x6c, 0x65, 0x64, 0x12, 0x1c, 0x0a, 0x09, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x73, 0x18, 0x05, 0x20, 0x03, 0x28, 0x09, 0x52, 0x09, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x73, 0x12, 0x12, 0x0a, 0x04, 0x74, 0x61, 0x67, 0x73,
	0x18, 0x06, 0x20, 0x03, 0x28, 0x09, 0x52, 0x04, 0x74, 0x61, 0x67, 0x73, 0x12, 0x1d, 0x0a, 0x0a,
	0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x5f, 0x61, 0x74, 0x18, 0x07, 0x20, 0x01, 0x28, 0x03,
	0x52, 0x09, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x41, 0x74, 0x12, 0x27, 0x0a, 0x05, 0x72,
	0x6f, 0x75, 0x74, 0x65, 0x18, 0x08, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x11, 0x2e, 0x6b, 0x6f, 0x6e,
	0x67, 0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x2e, 0x52, 0x6f, 0x75, 0x74, 0x65, 0x52, 0x05, 0x72,
	0x6f, 0x75, 0x74, 0x65, 0x12, 0x2d, 0x0a, 0x07, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x18,
	0x09, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x13, 0x2e, 0x6b, 0x6f, 0x6e, 0x67, 0x2e, 0x6d, 0x6f, 0x64,
	0x65, 0x6c, 0x2e, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x52, 0x07, 0x73, 0x65, 0x72, 0x76,
	0x69, 0x63, 0x65, 0x12, 0x30, 0x0a, 0x08, 0x63, 0x6f, 0x6e, 0x73, 0x75, 0x6d, 0x65, 0x72, 0x18,
	0x0a, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x14, 0x2e, 0x6b, 0x6f, 0x6e, 0x67, 0x2e, 0x6d, 0x6f, 0x64,
	0x65, 0x6c, 0x2e, 0x43, 0x6f, 0x6e, 0x73, 0x75, 0x6d, 0x65, 0x72, 0x52, 0x08, 0x63, 0x6f, 0x6e,
	0x73, 0x75, 0x6d, 0x65, 0x72, 0x42, 0x39, 0x5a, 0x37, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e,
	0x63, 0x6f, 0x6d, 0x2f, 0x6b, 0x6f, 0x6e, 0x67, 0x2f, 0x6b, 0x6f, 0x6b, 0x6f, 0x2f, 0x69, 0x6e,
	0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x77, 0x72, 0x70, 0x63, 0x2f,
	0x6b, 0x6f, 0x6e, 0x67, 0x2f, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x3b, 0x6d, 0x6f, 0x64, 0x65, 0x6c,
	0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_kong_model_plugin_proto_rawDescOnce sync.Once
	file_kong_model_plugin_proto_rawDescData = file_kong_model_plugin_proto_rawDesc
)

func file_kong_model_plugin_proto_rawDescGZIP() []byte {
	file_kong_model_plugin_proto_rawDescOnce.Do(func() {
		file_kong_model_plugin_proto_rawDescData = protoimpl.X.CompressGZIP(file_kong_model_plugin_proto_rawDescData)
	})
	return file_kong_model_plugin_proto_rawDescData
}

var file_kong_model_plugin_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_kong_model_plugin_proto_goTypes = []interface{}{
	(*Plugin)(nil),          // 0: kong.model.Plugin
	(*structpb.Struct)(nil), // 1: google.protobuf.Struct
	(*Route)(nil),           // 2: kong.model.Route
	(*Service)(nil),         // 3: kong.model.Service
	(*Consumer)(nil),        // 4: kong.model.Consumer
}
var file_kong_model_plugin_proto_depIdxs = []int32{
	1, // 0: kong.model.Plugin.config:type_name -> google.protobuf.Struct
	2, // 1: kong.model.Plugin.route:type_name -> kong.model.Route
	3, // 2: kong.model.Plugin.service:type_name -> kong.model.Service
	4, // 3: kong.model.Plugin.consumer:type_name -> kong.model.Consumer
	4, // [4:4] is the sub-list for method output_type
	4, // [4:4] is the sub-list for method input_type
	4, // [4:4] is the sub-list for extension type_name
	4, // [4:4] is the sub-list for extension extendee
	0, // [0:4] is the sub-list for field type_name
}

func init() { file_kong_model_plugin_proto_init() }
func file_kong_model_plugin_proto_init() {
	if File_kong_model_plugin_proto != nil {
		return
	}
	file_kong_model_consumer_proto_init()
	file_kong_model_service_proto_init()
	file_kong_model_route_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_kong_model_plugin_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Plugin); i {
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
			RawDescriptor: file_kong_model_plugin_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_kong_model_plugin_proto_goTypes,
		DependencyIndexes: file_kong_model_plugin_proto_depIdxs,
		MessageInfos:      file_kong_model_plugin_proto_msgTypes,
	}.Build()
	File_kong_model_plugin_proto = out.File
	file_kong_model_plugin_proto_rawDesc = nil
	file_kong_model_plugin_proto_goTypes = nil
	file_kong_model_plugin_proto_depIdxs = nil
}
