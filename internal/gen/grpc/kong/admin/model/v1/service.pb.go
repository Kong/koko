// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1-devel
// 	protoc        (unknown)
// source: kong/admin/model/v1/service.proto

package v1

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	wrapperspb "google.golang.org/protobuf/types/known/wrapperspb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type Service struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id                string                `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Name              string                `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	ConnectTimeout    int32                 `protobuf:"varint,3,opt,name=connect_timeout,json=connectTimeout,proto3" json:"connect_timeout,omitempty"`
	CreatedAt         int32                 `protobuf:"varint,4,opt,name=created_at,json=createdAt,proto3" json:"created_at,omitempty"`
	Host              string                `protobuf:"bytes,5,opt,name=host,proto3" json:"host,omitempty"`
	Path              string                `protobuf:"bytes,6,opt,name=path,proto3" json:"path,omitempty"`
	Port              int32                 `protobuf:"varint,7,opt,name=port,proto3" json:"port,omitempty"`
	Protocol          string                `protobuf:"bytes,8,opt,name=protocol,proto3" json:"protocol,omitempty"`
	ReadTimeout       int32                 `protobuf:"varint,9,opt,name=read_timeout,json=readTimeout,proto3" json:"read_timeout,omitempty"`
	Retries           int32                 `protobuf:"varint,10,opt,name=retries,proto3" json:"retries,omitempty"`
	UpdatedAt         int32                 `protobuf:"varint,11,opt,name=updated_at,json=updatedAt,proto3" json:"updated_at,omitempty"`
	Url               string                `protobuf:"bytes,12,opt,name=url,proto3" json:"url,omitempty"`
	WriteTimeout      int32                 `protobuf:"varint,13,opt,name=write_timeout,json=writeTimeout,proto3" json:"write_timeout,omitempty"`
	Tags              []string              `protobuf:"bytes,14,rep,name=tags,proto3" json:"tags,omitempty"`
	TlsVerify         *bool                 `protobuf:"varint,15,opt,name=tls_verify,json=tlsVerify,proto3,oneof" json:"tls_verify,omitempty"`
	TlsVerifyDepth    int32                 `protobuf:"varint,16,opt,name=tls_verify_depth,json=tlsVerifyDepth,proto3" json:"tls_verify_depth,omitempty"`
	ClientCertificate *Certificate          `protobuf:"bytes,17,opt,name=client_certificate,json=clientCertificate,proto3" json:"client_certificate,omitempty"`
	CaCertificates    []string              `protobuf:"bytes,18,rep,name=ca_certificates,json=caCertificates,proto3" json:"ca_certificates,omitempty"`
	Enabled           *wrapperspb.BoolValue `protobuf:"bytes,19,opt,name=enabled,proto3" json:"enabled,omitempty"`
}

func (x *Service) Reset() {
	*x = Service{}
	if protoimpl.UnsafeEnabled {
		mi := &file_kong_admin_model_v1_service_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Service) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Service) ProtoMessage() {}

func (x *Service) ProtoReflect() protoreflect.Message {
	mi := &file_kong_admin_model_v1_service_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Service.ProtoReflect.Descriptor instead.
func (*Service) Descriptor() ([]byte, []int) {
	return file_kong_admin_model_v1_service_proto_rawDescGZIP(), []int{0}
}

func (x *Service) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *Service) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Service) GetConnectTimeout() int32 {
	if x != nil {
		return x.ConnectTimeout
	}
	return 0
}

func (x *Service) GetCreatedAt() int32 {
	if x != nil {
		return x.CreatedAt
	}
	return 0
}

func (x *Service) GetHost() string {
	if x != nil {
		return x.Host
	}
	return ""
}

func (x *Service) GetPath() string {
	if x != nil {
		return x.Path
	}
	return ""
}

func (x *Service) GetPort() int32 {
	if x != nil {
		return x.Port
	}
	return 0
}

func (x *Service) GetProtocol() string {
	if x != nil {
		return x.Protocol
	}
	return ""
}

func (x *Service) GetReadTimeout() int32 {
	if x != nil {
		return x.ReadTimeout
	}
	return 0
}

func (x *Service) GetRetries() int32 {
	if x != nil {
		return x.Retries
	}
	return 0
}

func (x *Service) GetUpdatedAt() int32 {
	if x != nil {
		return x.UpdatedAt
	}
	return 0
}

func (x *Service) GetUrl() string {
	if x != nil {
		return x.Url
	}
	return ""
}

func (x *Service) GetWriteTimeout() int32 {
	if x != nil {
		return x.WriteTimeout
	}
	return 0
}

func (x *Service) GetTags() []string {
	if x != nil {
		return x.Tags
	}
	return nil
}

func (x *Service) GetTlsVerify() bool {
	if x != nil && x.TlsVerify != nil {
		return *x.TlsVerify
	}
	return false
}

func (x *Service) GetTlsVerifyDepth() int32 {
	if x != nil {
		return x.TlsVerifyDepth
	}
	return 0
}

func (x *Service) GetClientCertificate() *Certificate {
	if x != nil {
		return x.ClientCertificate
	}
	return nil
}

func (x *Service) GetCaCertificates() []string {
	if x != nil {
		return x.CaCertificates
	}
	return nil
}

func (x *Service) GetEnabled() *wrapperspb.BoolValue {
	if x != nil {
		return x.Enabled
	}
	return nil
}

var File_kong_admin_model_v1_service_proto protoreflect.FileDescriptor

var file_kong_admin_model_v1_service_proto_rawDesc = []byte{
	0x0a, 0x21, 0x6b, 0x6f, 0x6e, 0x67, 0x2f, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2f, 0x6d, 0x6f, 0x64,
	0x65, 0x6c, 0x2f, 0x76, 0x31, 0x2f, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x12, 0x13, 0x6b, 0x6f, 0x6e, 0x67, 0x2e, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2e,
	0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x2e, 0x76, 0x31, 0x1a, 0x1e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65,
	0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x77, 0x72, 0x61, 0x70, 0x70, 0x65,
	0x72, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x25, 0x6b, 0x6f, 0x6e, 0x67, 0x2f, 0x61,
	0x64, 0x6d, 0x69, 0x6e, 0x2f, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x2f, 0x76, 0x31, 0x2f, 0x63, 0x65,
	0x72, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22,
	0x81, 0x05, 0x0a, 0x07, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x0e, 0x0a, 0x02, 0x69,
	0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x6e,
	0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12,
	0x27, 0x0a, 0x0f, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x5f, 0x74, 0x69, 0x6d, 0x65, 0x6f,
	0x75, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0e, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63,
	0x74, 0x54, 0x69, 0x6d, 0x65, 0x6f, 0x75, 0x74, 0x12, 0x1d, 0x0a, 0x0a, 0x63, 0x72, 0x65, 0x61,
	0x74, 0x65, 0x64, 0x5f, 0x61, 0x74, 0x18, 0x04, 0x20, 0x01, 0x28, 0x05, 0x52, 0x09, 0x63, 0x72,
	0x65, 0x61, 0x74, 0x65, 0x64, 0x41, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x68, 0x6f, 0x73, 0x74, 0x18,
	0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x68, 0x6f, 0x73, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x70,
	0x61, 0x74, 0x68, 0x18, 0x06, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x70, 0x61, 0x74, 0x68, 0x12,
	0x12, 0x0a, 0x04, 0x70, 0x6f, 0x72, 0x74, 0x18, 0x07, 0x20, 0x01, 0x28, 0x05, 0x52, 0x04, 0x70,
	0x6f, 0x72, 0x74, 0x12, 0x1a, 0x0a, 0x08, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x18,
	0x08, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x12,
	0x21, 0x0a, 0x0c, 0x72, 0x65, 0x61, 0x64, 0x5f, 0x74, 0x69, 0x6d, 0x65, 0x6f, 0x75, 0x74, 0x18,
	0x09, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0b, 0x72, 0x65, 0x61, 0x64, 0x54, 0x69, 0x6d, 0x65, 0x6f,
	0x75, 0x74, 0x12, 0x18, 0x0a, 0x07, 0x72, 0x65, 0x74, 0x72, 0x69, 0x65, 0x73, 0x18, 0x0a, 0x20,
	0x01, 0x28, 0x05, 0x52, 0x07, 0x72, 0x65, 0x74, 0x72, 0x69, 0x65, 0x73, 0x12, 0x1d, 0x0a, 0x0a,
	0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x64, 0x5f, 0x61, 0x74, 0x18, 0x0b, 0x20, 0x01, 0x28, 0x05,
	0x52, 0x09, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x64, 0x41, 0x74, 0x12, 0x10, 0x0a, 0x03, 0x75,
	0x72, 0x6c, 0x18, 0x0c, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x75, 0x72, 0x6c, 0x12, 0x23, 0x0a,
	0x0d, 0x77, 0x72, 0x69, 0x74, 0x65, 0x5f, 0x74, 0x69, 0x6d, 0x65, 0x6f, 0x75, 0x74, 0x18, 0x0d,
	0x20, 0x01, 0x28, 0x05, 0x52, 0x0c, 0x77, 0x72, 0x69, 0x74, 0x65, 0x54, 0x69, 0x6d, 0x65, 0x6f,
	0x75, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x74, 0x61, 0x67, 0x73, 0x18, 0x0e, 0x20, 0x03, 0x28, 0x09,
	0x52, 0x04, 0x74, 0x61, 0x67, 0x73, 0x12, 0x22, 0x0a, 0x0a, 0x74, 0x6c, 0x73, 0x5f, 0x76, 0x65,
	0x72, 0x69, 0x66, 0x79, 0x18, 0x0f, 0x20, 0x01, 0x28, 0x08, 0x48, 0x00, 0x52, 0x09, 0x74, 0x6c,
	0x73, 0x56, 0x65, 0x72, 0x69, 0x66, 0x79, 0x88, 0x01, 0x01, 0x12, 0x28, 0x0a, 0x10, 0x74, 0x6c,
	0x73, 0x5f, 0x76, 0x65, 0x72, 0x69, 0x66, 0x79, 0x5f, 0x64, 0x65, 0x70, 0x74, 0x68, 0x18, 0x10,
	0x20, 0x01, 0x28, 0x05, 0x52, 0x0e, 0x74, 0x6c, 0x73, 0x56, 0x65, 0x72, 0x69, 0x66, 0x79, 0x44,
	0x65, 0x70, 0x74, 0x68, 0x12, 0x4f, 0x0a, 0x12, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x5f, 0x63,
	0x65, 0x72, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x65, 0x18, 0x11, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x20, 0x2e, 0x6b, 0x6f, 0x6e, 0x67, 0x2e, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2e, 0x6d, 0x6f,
	0x64, 0x65, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x43, 0x65, 0x72, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61,
	0x74, 0x65, 0x52, 0x11, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x43, 0x65, 0x72, 0x74, 0x69, 0x66,
	0x69, 0x63, 0x61, 0x74, 0x65, 0x12, 0x27, 0x0a, 0x0f, 0x63, 0x61, 0x5f, 0x63, 0x65, 0x72, 0x74,
	0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x65, 0x73, 0x18, 0x12, 0x20, 0x03, 0x28, 0x09, 0x52, 0x0e,
	0x63, 0x61, 0x43, 0x65, 0x72, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x65, 0x73, 0x12, 0x34,
	0x0a, 0x07, 0x65, 0x6e, 0x61, 0x62, 0x6c, 0x65, 0x64, 0x18, 0x13, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75,
	0x66, 0x2e, 0x42, 0x6f, 0x6f, 0x6c, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x52, 0x07, 0x65, 0x6e, 0x61,
	0x62, 0x6c, 0x65, 0x64, 0x42, 0x0d, 0x0a, 0x0b, 0x5f, 0x74, 0x6c, 0x73, 0x5f, 0x76, 0x65, 0x72,
	0x69, 0x66, 0x79, 0x42, 0x3f, 0x5a, 0x3d, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f,
	0x6d, 0x2f, 0x6b, 0x6f, 0x6e, 0x67, 0x2f, 0x6b, 0x6f, 0x6b, 0x6f, 0x2f, 0x69, 0x6e, 0x74, 0x65,
	0x72, 0x6e, 0x61, 0x6c, 0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x67, 0x72, 0x70, 0x63, 0x2f, 0x6b, 0x6f,
	0x6e, 0x67, 0x2f, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2f, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x2f, 0x76,
	0x31, 0x3b, 0x76, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_kong_admin_model_v1_service_proto_rawDescOnce sync.Once
	file_kong_admin_model_v1_service_proto_rawDescData = file_kong_admin_model_v1_service_proto_rawDesc
)

func file_kong_admin_model_v1_service_proto_rawDescGZIP() []byte {
	file_kong_admin_model_v1_service_proto_rawDescOnce.Do(func() {
		file_kong_admin_model_v1_service_proto_rawDescData = protoimpl.X.CompressGZIP(file_kong_admin_model_v1_service_proto_rawDescData)
	})
	return file_kong_admin_model_v1_service_proto_rawDescData
}

var file_kong_admin_model_v1_service_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_kong_admin_model_v1_service_proto_goTypes = []interface{}{
	(*Service)(nil),              // 0: kong.admin.model.v1.Service
	(*Certificate)(nil),          // 1: kong.admin.model.v1.Certificate
	(*wrapperspb.BoolValue)(nil), // 2: google.protobuf.BoolValue
}
var file_kong_admin_model_v1_service_proto_depIdxs = []int32{
	1, // 0: kong.admin.model.v1.Service.client_certificate:type_name -> kong.admin.model.v1.Certificate
	2, // 1: kong.admin.model.v1.Service.enabled:type_name -> google.protobuf.BoolValue
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_kong_admin_model_v1_service_proto_init() }
func file_kong_admin_model_v1_service_proto_init() {
	if File_kong_admin_model_v1_service_proto != nil {
		return
	}
	file_kong_admin_model_v1_certificate_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_kong_admin_model_v1_service_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Service); i {
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
	file_kong_admin_model_v1_service_proto_msgTypes[0].OneofWrappers = []interface{}{}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_kong_admin_model_v1_service_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_kong_admin_model_v1_service_proto_goTypes,
		DependencyIndexes: file_kong_admin_model_v1_service_proto_depIdxs,
		MessageInfos:      file_kong_admin_model_v1_service_proto_msgTypes,
	}.Build()
	File_kong_admin_model_v1_service_proto = out.File
	file_kong_admin_model_v1_service_proto_rawDesc = nil
	file_kong_admin_model_v1_service_proto_goTypes = nil
	file_kong_admin_model_v1_service_proto_depIdxs = nil
}
