// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.0
// 	protoc        (unknown)
// source: kong/admin/model/v1/plugin_schema.proto

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

type PluginSchema struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name      string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	LuaSchema string `protobuf:"bytes,2,opt,name=lua_schema,json=luaSchema,proto3" json:"lua_schema,omitempty"`
	CreatedAt int32  `protobuf:"varint,3,opt,name=created_at,json=createdAt,proto3" json:"created_at,omitempty"`
	UpdatedAt int32  `protobuf:"varint,4,opt,name=updated_at,json=updatedAt,proto3" json:"updated_at,omitempty"`
}

func (x *PluginSchema) Reset() {
	*x = PluginSchema{}
	if protoimpl.UnsafeEnabled {
		mi := &file_kong_admin_model_v1_plugin_schema_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PluginSchema) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PluginSchema) ProtoMessage() {}

func (x *PluginSchema) ProtoReflect() protoreflect.Message {
	mi := &file_kong_admin_model_v1_plugin_schema_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PluginSchema.ProtoReflect.Descriptor instead.
func (*PluginSchema) Descriptor() ([]byte, []int) {
	return file_kong_admin_model_v1_plugin_schema_proto_rawDescGZIP(), []int{0}
}

func (x *PluginSchema) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *PluginSchema) GetLuaSchema() string {
	if x != nil {
		return x.LuaSchema
	}
	return ""
}

func (x *PluginSchema) GetCreatedAt() int32 {
	if x != nil {
		return x.CreatedAt
	}
	return 0
}

func (x *PluginSchema) GetUpdatedAt() int32 {
	if x != nil {
		return x.UpdatedAt
	}
	return 0
}

var File_kong_admin_model_v1_plugin_schema_proto protoreflect.FileDescriptor

var file_kong_admin_model_v1_plugin_schema_proto_rawDesc = []byte{
	0x0a, 0x27, 0x6b, 0x6f, 0x6e, 0x67, 0x2f, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2f, 0x6d, 0x6f, 0x64,
	0x65, 0x6c, 0x2f, 0x76, 0x31, 0x2f, 0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x5f, 0x73, 0x63, 0x68,
	0x65, 0x6d, 0x61, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x13, 0x6b, 0x6f, 0x6e, 0x67, 0x2e,
	0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x2e, 0x76, 0x31, 0x22, 0x7f,
	0x0a, 0x0c, 0x50, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x53, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x12, 0x12,
	0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61,
	0x6d, 0x65, 0x12, 0x1d, 0x0a, 0x0a, 0x6c, 0x75, 0x61, 0x5f, 0x73, 0x63, 0x68, 0x65, 0x6d, 0x61,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x6c, 0x75, 0x61, 0x53, 0x63, 0x68, 0x65, 0x6d,
	0x61, 0x12, 0x1d, 0x0a, 0x0a, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x5f, 0x61, 0x74, 0x18,
	0x03, 0x20, 0x01, 0x28, 0x05, 0x52, 0x09, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x41, 0x74,
	0x12, 0x1d, 0x0a, 0x0a, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x64, 0x5f, 0x61, 0x74, 0x18, 0x04,
	0x20, 0x01, 0x28, 0x05, 0x52, 0x09, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x64, 0x41, 0x74, 0x42,
	0x3f, 0x5a, 0x3d, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6b, 0x6f,
	0x6e, 0x67, 0x2f, 0x6b, 0x6f, 0x6b, 0x6f, 0x2f, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c,
	0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x67, 0x72, 0x70, 0x63, 0x2f, 0x6b, 0x6f, 0x6e, 0x67, 0x2f, 0x61,
	0x64, 0x6d, 0x69, 0x6e, 0x2f, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x2f, 0x76, 0x31, 0x3b, 0x76, 0x31,
	0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_kong_admin_model_v1_plugin_schema_proto_rawDescOnce sync.Once
	file_kong_admin_model_v1_plugin_schema_proto_rawDescData = file_kong_admin_model_v1_plugin_schema_proto_rawDesc
)

func file_kong_admin_model_v1_plugin_schema_proto_rawDescGZIP() []byte {
	file_kong_admin_model_v1_plugin_schema_proto_rawDescOnce.Do(func() {
		file_kong_admin_model_v1_plugin_schema_proto_rawDescData = protoimpl.X.CompressGZIP(file_kong_admin_model_v1_plugin_schema_proto_rawDescData)
	})
	return file_kong_admin_model_v1_plugin_schema_proto_rawDescData
}

var file_kong_admin_model_v1_plugin_schema_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_kong_admin_model_v1_plugin_schema_proto_goTypes = []interface{}{
	(*PluginSchema)(nil), // 0: kong.admin.model.v1.PluginSchema
}
var file_kong_admin_model_v1_plugin_schema_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_kong_admin_model_v1_plugin_schema_proto_init() }
func file_kong_admin_model_v1_plugin_schema_proto_init() {
	if File_kong_admin_model_v1_plugin_schema_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_kong_admin_model_v1_plugin_schema_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PluginSchema); i {
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
			RawDescriptor: file_kong_admin_model_v1_plugin_schema_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_kong_admin_model_v1_plugin_schema_proto_goTypes,
		DependencyIndexes: file_kong_admin_model_v1_plugin_schema_proto_depIdxs,
		MessageInfos:      file_kong_admin_model_v1_plugin_schema_proto_msgTypes,
	}.Build()
	File_kong_admin_model_v1_plugin_schema_proto = out.File
	file_kong_admin_model_v1_plugin_schema_proto_rawDesc = nil
	file_kong_admin_model_v1_plugin_schema_proto_goTypes = nil
	file_kong_admin_model_v1_plugin_schema_proto_depIdxs = nil
}
