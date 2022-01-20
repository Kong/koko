// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.27.1
// 	protoc        v3.17.3
// source: kong/admin/service/v1/schemas.proto

package v1

import (
	_ "google.golang.org/genproto/googleapis/api/annotations"
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

type GetSchemasRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
}

func (x *GetSchemasRequest) Reset() {
	*x = GetSchemasRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_kong_admin_service_v1_schemas_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetSchemasRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetSchemasRequest) ProtoMessage() {}

func (x *GetSchemasRequest) ProtoReflect() protoreflect.Message {
	mi := &file_kong_admin_service_v1_schemas_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetSchemasRequest.ProtoReflect.Descriptor instead.
func (*GetSchemasRequest) Descriptor() ([]byte, []int) {
	return file_kong_admin_service_v1_schemas_proto_rawDescGZIP(), []int{0}
}

func (x *GetSchemasRequest) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

var File_kong_admin_service_v1_schemas_proto protoreflect.FileDescriptor

var file_kong_admin_service_v1_schemas_proto_rawDesc = []byte{
	0x0a, 0x23, 0x6b, 0x6f, 0x6e, 0x67, 0x2f, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2f, 0x73, 0x65, 0x72,
	0x76, 0x69, 0x63, 0x65, 0x2f, 0x76, 0x31, 0x2f, 0x73, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x73, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x15, 0x6b, 0x6f, 0x6e, 0x67, 0x2e, 0x61, 0x64, 0x6d, 0x69,
	0x6e, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x76, 0x31, 0x1a, 0x1c, 0x67, 0x6f,
	0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1c, 0x67, 0x6f, 0x6f, 0x67,
	0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x73, 0x74, 0x72, 0x75,
	0x63, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x27, 0x0a, 0x11, 0x47, 0x65, 0x74, 0x53,
	0x63, 0x68, 0x65, 0x6d, 0x61, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x12, 0x0a,
	0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d,
	0x65, 0x32, 0x82, 0x01, 0x0a, 0x0e, 0x53, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x73, 0x53, 0x65, 0x72,
	0x76, 0x69, 0x63, 0x65, 0x12, 0x70, 0x0a, 0x0a, 0x47, 0x65, 0x74, 0x53, 0x63, 0x68, 0x65, 0x6d,
	0x61, 0x73, 0x12, 0x28, 0x2e, 0x6b, 0x6f, 0x6e, 0x67, 0x2e, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2e,
	0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x47, 0x65, 0x74, 0x53, 0x63,
	0x68, 0x65, 0x6d, 0x61, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x17, 0x2e, 0x67,
	0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x53,
	0x74, 0x72, 0x75, 0x63, 0x74, 0x22, 0x1f, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x19, 0x12, 0x17, 0x2f,
	0x76, 0x31, 0x2f, 0x73, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x73, 0x2f, 0x6a, 0x73, 0x6f, 0x6e, 0x2f,
	0x7b, 0x6e, 0x61, 0x6d, 0x65, 0x7d, 0x42, 0x3c, 0x5a, 0x3a, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62,
	0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6b, 0x6f, 0x6e, 0x67, 0x2f, 0x6b, 0x6f, 0x6b, 0x6f, 0x2f, 0x69,
	0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x6b, 0x6f, 0x6e, 0x67,
	0x2f, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2f, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2f, 0x76,
	0x31, 0x3b, 0x76, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_kong_admin_service_v1_schemas_proto_rawDescOnce sync.Once
	file_kong_admin_service_v1_schemas_proto_rawDescData = file_kong_admin_service_v1_schemas_proto_rawDesc
)

func file_kong_admin_service_v1_schemas_proto_rawDescGZIP() []byte {
	file_kong_admin_service_v1_schemas_proto_rawDescOnce.Do(func() {
		file_kong_admin_service_v1_schemas_proto_rawDescData = protoimpl.X.CompressGZIP(file_kong_admin_service_v1_schemas_proto_rawDescData)
	})
	return file_kong_admin_service_v1_schemas_proto_rawDescData
}

var file_kong_admin_service_v1_schemas_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_kong_admin_service_v1_schemas_proto_goTypes = []interface{}{
	(*GetSchemasRequest)(nil), // 0: kong.admin.service.v1.GetSchemasRequest
	(*structpb.Struct)(nil),   // 1: google.protobuf.Struct
}
var file_kong_admin_service_v1_schemas_proto_depIdxs = []int32{
	0, // 0: kong.admin.service.v1.SchemasService.GetSchemas:input_type -> kong.admin.service.v1.GetSchemasRequest
	1, // 1: kong.admin.service.v1.SchemasService.GetSchemas:output_type -> google.protobuf.Struct
	1, // [1:2] is the sub-list for method output_type
	0, // [0:1] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_kong_admin_service_v1_schemas_proto_init() }
func file_kong_admin_service_v1_schemas_proto_init() {
	if File_kong_admin_service_v1_schemas_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_kong_admin_service_v1_schemas_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetSchemasRequest); i {
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
			RawDescriptor: file_kong_admin_service_v1_schemas_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_kong_admin_service_v1_schemas_proto_goTypes,
		DependencyIndexes: file_kong_admin_service_v1_schemas_proto_depIdxs,
		MessageInfos:      file_kong_admin_service_v1_schemas_proto_msgTypes,
	}.Build()
	File_kong_admin_service_v1_schemas_proto = out.File
	file_kong_admin_service_v1_schemas_proto_rawDesc = nil
	file_kong_admin_service_v1_schemas_proto_goTypes = nil
	file_kong_admin_service_v1_schemas_proto_depIdxs = nil
}
