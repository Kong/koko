// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.27.1
// 	protoc        v3.17.3
// source: kong/model/target.proto

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

type Target struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	CreatedAt float64   `protobuf:"fixed64,1,opt,name=created_at,json=createdAt,proto3" json:"created_at,omitempty"`
	Id        string    `protobuf:"bytes,2,opt,name=id,proto3" json:"id,omitempty"`
	Target    string    `protobuf:"bytes,3,opt,name=target,proto3" json:"target,omitempty"`
	Weight    int64     `protobuf:"varint,4,opt,name=weight,proto3" json:"weight,omitempty"`
	Tags      []string  `protobuf:"bytes,5,rep,name=tags,proto3" json:"tags,omitempty"`
	Upstream  *Upstream `protobuf:"bytes,6,opt,name=upstream,proto3" json:"upstream,omitempty"`
}

func (x *Target) Reset() {
	*x = Target{}
	if protoimpl.UnsafeEnabled {
		mi := &file_kong_model_target_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Target) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Target) ProtoMessage() {}

func (x *Target) ProtoReflect() protoreflect.Message {
	mi := &file_kong_model_target_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Target.ProtoReflect.Descriptor instead.
func (*Target) Descriptor() ([]byte, []int) {
	return file_kong_model_target_proto_rawDescGZIP(), []int{0}
}

func (x *Target) GetCreatedAt() float64 {
	if x != nil {
		return x.CreatedAt
	}
	return 0
}

func (x *Target) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *Target) GetTarget() string {
	if x != nil {
		return x.Target
	}
	return ""
}

func (x *Target) GetWeight() int64 {
	if x != nil {
		return x.Weight
	}
	return 0
}

func (x *Target) GetTags() []string {
	if x != nil {
		return x.Tags
	}
	return nil
}

func (x *Target) GetUpstream() *Upstream {
	if x != nil {
		return x.Upstream
	}
	return nil
}

var File_kong_model_target_proto protoreflect.FileDescriptor

var file_kong_model_target_proto_rawDesc = []byte{
	0x0a, 0x17, 0x6b, 0x6f, 0x6e, 0x67, 0x2f, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x2f, 0x74, 0x61, 0x72,
	0x67, 0x65, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0a, 0x6b, 0x6f, 0x6e, 0x67, 0x2e,
	0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x1a, 0x19, 0x6b, 0x6f, 0x6e, 0x67, 0x2f, 0x6d, 0x6f, 0x64, 0x65,
	0x6c, 0x2f, 0x75, 0x70, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x22, 0xad, 0x01, 0x0a, 0x06, 0x54, 0x61, 0x72, 0x67, 0x65, 0x74, 0x12, 0x1d, 0x0a, 0x0a, 0x63,
	0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x5f, 0x61, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x01, 0x52,
	0x09, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x41, 0x74, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12, 0x16, 0x0a, 0x06, 0x74, 0x61,
	0x72, 0x67, 0x65, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x74, 0x61, 0x72, 0x67,
	0x65, 0x74, 0x12, 0x16, 0x0a, 0x06, 0x77, 0x65, 0x69, 0x67, 0x68, 0x74, 0x18, 0x04, 0x20, 0x01,
	0x28, 0x03, 0x52, 0x06, 0x77, 0x65, 0x69, 0x67, 0x68, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x74, 0x61,
	0x67, 0x73, 0x18, 0x05, 0x20, 0x03, 0x28, 0x09, 0x52, 0x04, 0x74, 0x61, 0x67, 0x73, 0x12, 0x30,
	0x0a, 0x08, 0x75, 0x70, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x18, 0x06, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x14, 0x2e, 0x6b, 0x6f, 0x6e, 0x67, 0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x2e, 0x55, 0x70,
	0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x52, 0x08, 0x75, 0x70, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d,
	0x42, 0x39, 0x5a, 0x37, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6b,
	0x6f, 0x6e, 0x67, 0x2f, 0x6b, 0x6f, 0x6b, 0x6f, 0x2f, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61,
	0x6c, 0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x77, 0x72, 0x70, 0x63, 0x2f, 0x6b, 0x6f, 0x6e, 0x67, 0x2f,
	0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x3b, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x62, 0x06, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x33,
}

var (
	file_kong_model_target_proto_rawDescOnce sync.Once
	file_kong_model_target_proto_rawDescData = file_kong_model_target_proto_rawDesc
)

func file_kong_model_target_proto_rawDescGZIP() []byte {
	file_kong_model_target_proto_rawDescOnce.Do(func() {
		file_kong_model_target_proto_rawDescData = protoimpl.X.CompressGZIP(file_kong_model_target_proto_rawDescData)
	})
	return file_kong_model_target_proto_rawDescData
}

var file_kong_model_target_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_kong_model_target_proto_goTypes = []interface{}{
	(*Target)(nil),   // 0: kong.model.Target
	(*Upstream)(nil), // 1: kong.model.Upstream
}
var file_kong_model_target_proto_depIdxs = []int32{
	1, // 0: kong.model.Target.upstream:type_name -> kong.model.Upstream
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_kong_model_target_proto_init() }
func file_kong_model_target_proto_init() {
	if File_kong_model_target_proto != nil {
		return
	}
	file_kong_model_upstream_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_kong_model_target_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Target); i {
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
			RawDescriptor: file_kong_model_target_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_kong_model_target_proto_goTypes,
		DependencyIndexes: file_kong_model_target_proto_depIdxs,
		MessageInfos:      file_kong_model_target_proto_msgTypes,
	}.Build()
	File_kong_model_target_proto = out.File
	file_kong_model_target_proto_rawDesc = nil
	file_kong_model_target_proto_goTypes = nil
	file_kong_model_target_proto_depIdxs = nil
}
