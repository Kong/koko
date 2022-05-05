// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.0
// 	protoc        (unknown)
// source: kong/model/certificate.proto

package model

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

type Certificate struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id        string   `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Cert      string   `protobuf:"bytes,2,opt,name=cert,proto3" json:"cert,omitempty"`
	Key       string   `protobuf:"bytes,3,opt,name=key,proto3" json:"key,omitempty"`
	CertAlt   string   `protobuf:"bytes,4,opt,name=cert_alt,json=certAlt,proto3" json:"cert_alt,omitempty"`
	KeyAlt    string   `protobuf:"bytes,5,opt,name=key_alt,json=keyAlt,proto3" json:"key_alt,omitempty"`
	CreatedAt int32    `protobuf:"varint,6,opt,name=created_at,json=createdAt,proto3" json:"created_at,omitempty"`
	Tags      []string `protobuf:"bytes,7,rep,name=tags,proto3" json:"tags,omitempty"`
}

func (x *Certificate) Reset() {
	*x = Certificate{}
	if protoimpl.UnsafeEnabled {
		mi := &file_kong_model_certificate_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Certificate) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Certificate) ProtoMessage() {}

func (x *Certificate) ProtoReflect() protoreflect.Message {
	mi := &file_kong_model_certificate_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Certificate.ProtoReflect.Descriptor instead.
func (*Certificate) Descriptor() ([]byte, []int) {
	return file_kong_model_certificate_proto_rawDescGZIP(), []int{0}
}

func (x *Certificate) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *Certificate) GetCert() string {
	if x != nil {
		return x.Cert
	}
	return ""
}

func (x *Certificate) GetKey() string {
	if x != nil {
		return x.Key
	}
	return ""
}

func (x *Certificate) GetCertAlt() string {
	if x != nil {
		return x.CertAlt
	}
	return ""
}

func (x *Certificate) GetKeyAlt() string {
	if x != nil {
		return x.KeyAlt
	}
	return ""
}

func (x *Certificate) GetCreatedAt() int32 {
	if x != nil {
		return x.CreatedAt
	}
	return 0
}

func (x *Certificate) GetTags() []string {
	if x != nil {
		return x.Tags
	}
	return nil
}

var File_kong_model_certificate_proto protoreflect.FileDescriptor

var file_kong_model_certificate_proto_rawDesc = []byte{
	0x0a, 0x1c, 0x6b, 0x6f, 0x6e, 0x67, 0x2f, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x2f, 0x63, 0x65, 0x72,
	0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0a,
	0x6b, 0x6f, 0x6e, 0x67, 0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x22, 0xaa, 0x01, 0x0a, 0x0b, 0x43,
	0x65, 0x72, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x65, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x63, 0x65,
	0x72, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x63, 0x65, 0x72, 0x74, 0x12, 0x10,
	0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79,
	0x12, 0x19, 0x0a, 0x08, 0x63, 0x65, 0x72, 0x74, 0x5f, 0x61, 0x6c, 0x74, 0x18, 0x04, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x07, 0x63, 0x65, 0x72, 0x74, 0x41, 0x6c, 0x74, 0x12, 0x17, 0x0a, 0x07, 0x6b,
	0x65, 0x79, 0x5f, 0x61, 0x6c, 0x74, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x6b, 0x65,
	0x79, 0x41, 0x6c, 0x74, 0x12, 0x1d, 0x0a, 0x0a, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x5f,
	0x61, 0x74, 0x18, 0x06, 0x20, 0x01, 0x28, 0x05, 0x52, 0x09, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65,
	0x64, 0x41, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x74, 0x61, 0x67, 0x73, 0x18, 0x07, 0x20, 0x03, 0x28,
	0x09, 0x52, 0x04, 0x74, 0x61, 0x67, 0x73, 0x42, 0x39, 0x5a, 0x37, 0x67, 0x69, 0x74, 0x68, 0x75,
	0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6b, 0x6f, 0x6e, 0x67, 0x2f, 0x6b, 0x6f, 0x6b, 0x6f, 0x2f,
	0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x77, 0x72, 0x70,
	0x63, 0x2f, 0x6b, 0x6f, 0x6e, 0x67, 0x2f, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x3b, 0x6d, 0x6f, 0x64,
	0x65, 0x6c, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_kong_model_certificate_proto_rawDescOnce sync.Once
	file_kong_model_certificate_proto_rawDescData = file_kong_model_certificate_proto_rawDesc
)

func file_kong_model_certificate_proto_rawDescGZIP() []byte {
	file_kong_model_certificate_proto_rawDescOnce.Do(func() {
		file_kong_model_certificate_proto_rawDescData = protoimpl.X.CompressGZIP(file_kong_model_certificate_proto_rawDescData)
	})
	return file_kong_model_certificate_proto_rawDescData
}

var file_kong_model_certificate_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_kong_model_certificate_proto_goTypes = []interface{}{
	(*Certificate)(nil), // 0: kong.model.Certificate
}
var file_kong_model_certificate_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_kong_model_certificate_proto_init() }
func file_kong_model_certificate_proto_init() {
	if File_kong_model_certificate_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_kong_model_certificate_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Certificate); i {
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
			RawDescriptor: file_kong_model_certificate_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_kong_model_certificate_proto_goTypes,
		DependencyIndexes: file_kong_model_certificate_proto_depIdxs,
		MessageInfos:      file_kong_model_certificate_proto_msgTypes,
	}.Build()
	File_kong_model_certificate_proto = out.File
	file_kong_model_certificate_proto_rawDesc = nil
	file_kong_model_certificate_proto_goTypes = nil
	file_kong_model_certificate_proto_depIdxs = nil
}
