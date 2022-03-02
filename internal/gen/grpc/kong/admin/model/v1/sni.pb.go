// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.27.1
// 	protoc        (unknown)
// source: kong/admin/model/v1/sni.proto

package v1

import (
	_ "github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2/options"
	_ "google.golang.org/genproto/googleapis/api/annotations"
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

type SNI struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id          string       `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Name        string       `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	Certificate *Certificate `protobuf:"bytes,3,opt,name=certificate,proto3" json:"certificate,omitempty"`
	CreatedAt   int32        `protobuf:"varint,4,opt,name=created_at,json=createdAt,proto3" json:"created_at,omitempty"`
	UpdatedAt   int32        `protobuf:"varint,5,opt,name=updated_at,json=updatedAt,proto3" json:"updated_at,omitempty"`
	Tags        []string     `protobuf:"bytes,6,rep,name=tags,proto3" json:"tags,omitempty"`
}

func (x *SNI) Reset() {
	*x = SNI{}
	if protoimpl.UnsafeEnabled {
		mi := &file_kong_admin_model_v1_sni_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SNI) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SNI) ProtoMessage() {}

func (x *SNI) ProtoReflect() protoreflect.Message {
	mi := &file_kong_admin_model_v1_sni_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SNI.ProtoReflect.Descriptor instead.
func (*SNI) Descriptor() ([]byte, []int) {
	return file_kong_admin_model_v1_sni_proto_rawDescGZIP(), []int{0}
}

func (x *SNI) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *SNI) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *SNI) GetCertificate() *Certificate {
	if x != nil {
		return x.Certificate
	}
	return nil
}

func (x *SNI) GetCreatedAt() int32 {
	if x != nil {
		return x.CreatedAt
	}
	return 0
}

func (x *SNI) GetUpdatedAt() int32 {
	if x != nil {
		return x.UpdatedAt
	}
	return 0
}

func (x *SNI) GetTags() []string {
	if x != nil {
		return x.Tags
	}
	return nil
}

var File_kong_admin_model_v1_sni_proto protoreflect.FileDescriptor

var file_kong_admin_model_v1_sni_proto_rawDesc = []byte{
	0x0a, 0x1d, 0x6b, 0x6f, 0x6e, 0x67, 0x2f, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2f, 0x6d, 0x6f, 0x64,
	0x65, 0x6c, 0x2f, 0x76, 0x31, 0x2f, 0x73, 0x6e, 0x69, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12,
	0x13, 0x6b, 0x6f, 0x6e, 0x67, 0x2e, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2e, 0x6d, 0x6f, 0x64, 0x65,
	0x6c, 0x2e, 0x76, 0x31, 0x1a, 0x25, 0x6b, 0x6f, 0x6e, 0x67, 0x2f, 0x61, 0x64, 0x6d, 0x69, 0x6e,
	0x2f, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x2f, 0x76, 0x31, 0x2f, 0x63, 0x65, 0x72, 0x74, 0x69, 0x66,
	0x69, 0x63, 0x61, 0x74, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1f, 0x67, 0x6f, 0x6f,
	0x67, 0x6c, 0x65, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x66, 0x69, 0x65, 0x6c, 0x64, 0x5f, 0x62, 0x65,
	0x68, 0x61, 0x76, 0x69, 0x6f, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x63, 0x2d, 0x67, 0x65, 0x6e, 0x2d, 0x6f, 0x70, 0x65, 0x6e, 0x61, 0x70, 0x69,
	0x76, 0x32, 0x2f, 0x6f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2f, 0x61, 0x6e, 0x6e, 0x6f, 0x74,
	0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x87, 0x02, 0x0a,
	0x03, 0x53, 0x4e, 0x49, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x02, 0x69, 0x64, 0x12, 0x3f, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x09, 0x42, 0x2b, 0x92, 0x41, 0x24, 0x32, 0x22, 0x68, 0x6f, 0x73, 0x74, 0x6e, 0x61, 0x6d,
	0x65, 0x20, 0x6f, 0x72, 0x20, 0x68, 0x6f, 0x73, 0x74, 0x6e, 0x61, 0x6d, 0x65, 0x20, 0x77, 0x69,
	0x74, 0x68, 0x20, 0x77, 0x69, 0x6c, 0x64, 0x63, 0x61, 0x72, 0x64, 0xe2, 0x41, 0x01, 0x02, 0x52,
	0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x48, 0x0a, 0x0b, 0x63, 0x65, 0x72, 0x74, 0x69, 0x66, 0x69,
	0x63, 0x61, 0x74, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x20, 0x2e, 0x6b, 0x6f, 0x6e,
	0x67, 0x2e, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x2e, 0x76, 0x31,
	0x2e, 0x43, 0x65, 0x72, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x65, 0x42, 0x04, 0xe2, 0x41,
	0x01, 0x02, 0x52, 0x0b, 0x63, 0x65, 0x72, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x65, 0x12,
	0x1d, 0x0a, 0x0a, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x5f, 0x61, 0x74, 0x18, 0x04, 0x20,
	0x01, 0x28, 0x05, 0x52, 0x09, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x41, 0x74, 0x12, 0x1d,
	0x0a, 0x0a, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x64, 0x5f, 0x61, 0x74, 0x18, 0x05, 0x20, 0x01,
	0x28, 0x05, 0x52, 0x09, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x64, 0x41, 0x74, 0x12, 0x12, 0x0a,
	0x04, 0x74, 0x61, 0x67, 0x73, 0x18, 0x06, 0x20, 0x03, 0x28, 0x09, 0x52, 0x04, 0x74, 0x61, 0x67,
	0x73, 0x3a, 0x13, 0x92, 0x41, 0x10, 0x0a, 0x0e, 0xd2, 0x01, 0x0b, 0x63, 0x65, 0x72, 0x74, 0x69,
	0x66, 0x69, 0x63, 0x61, 0x74, 0x65, 0x42, 0x3f, 0x5a, 0x3d, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62,
	0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6b, 0x6f, 0x6e, 0x67, 0x2f, 0x6b, 0x6f, 0x6b, 0x6f, 0x2f, 0x69,
	0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x67, 0x72, 0x70, 0x63,
	0x2f, 0x6b, 0x6f, 0x6e, 0x67, 0x2f, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2f, 0x6d, 0x6f, 0x64, 0x65,
	0x6c, 0x2f, 0x76, 0x31, 0x3b, 0x76, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_kong_admin_model_v1_sni_proto_rawDescOnce sync.Once
	file_kong_admin_model_v1_sni_proto_rawDescData = file_kong_admin_model_v1_sni_proto_rawDesc
)

func file_kong_admin_model_v1_sni_proto_rawDescGZIP() []byte {
	file_kong_admin_model_v1_sni_proto_rawDescOnce.Do(func() {
		file_kong_admin_model_v1_sni_proto_rawDescData = protoimpl.X.CompressGZIP(file_kong_admin_model_v1_sni_proto_rawDescData)
	})
	return file_kong_admin_model_v1_sni_proto_rawDescData
}

var file_kong_admin_model_v1_sni_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_kong_admin_model_v1_sni_proto_goTypes = []interface{}{
	(*SNI)(nil),         // 0: kong.admin.model.v1.SNI
	(*Certificate)(nil), // 1: kong.admin.model.v1.Certificate
}
var file_kong_admin_model_v1_sni_proto_depIdxs = []int32{
	1, // 0: kong.admin.model.v1.SNI.certificate:type_name -> kong.admin.model.v1.Certificate
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_kong_admin_model_v1_sni_proto_init() }
func file_kong_admin_model_v1_sni_proto_init() {
	if File_kong_admin_model_v1_sni_proto != nil {
		return
	}
	file_kong_admin_model_v1_certificate_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_kong_admin_model_v1_sni_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SNI); i {
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
			RawDescriptor: file_kong_admin_model_v1_sni_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_kong_admin_model_v1_sni_proto_goTypes,
		DependencyIndexes: file_kong_admin_model_v1_sni_proto_depIdxs,
		MessageInfos:      file_kong_admin_model_v1_sni_proto_msgTypes,
	}.Build()
	File_kong_admin_model_v1_sni_proto = out.File
	file_kong_admin_model_v1_sni_proto_rawDesc = nil
	file_kong_admin_model_v1_sni_proto_goTypes = nil
	file_kong_admin_model_v1_sni_proto_depIdxs = nil
}
