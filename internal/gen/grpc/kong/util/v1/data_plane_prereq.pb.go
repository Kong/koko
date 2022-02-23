// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.27.1
// 	protoc        (unknown)
// source: kong/util/v1/data_plane_prereq.proto

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

type DataPlanePrerequisite struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to Config:
	//	*DataPlanePrerequisite_RequiredPlugins
	Config isDataPlanePrerequisite_Config `protobuf_oneof:"config"`
}

func (x *DataPlanePrerequisite) Reset() {
	*x = DataPlanePrerequisite{}
	if protoimpl.UnsafeEnabled {
		mi := &file_kong_util_v1_data_plane_prereq_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DataPlanePrerequisite) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DataPlanePrerequisite) ProtoMessage() {}

func (x *DataPlanePrerequisite) ProtoReflect() protoreflect.Message {
	mi := &file_kong_util_v1_data_plane_prereq_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DataPlanePrerequisite.ProtoReflect.Descriptor instead.
func (*DataPlanePrerequisite) Descriptor() ([]byte, []int) {
	return file_kong_util_v1_data_plane_prereq_proto_rawDescGZIP(), []int{0}
}

func (m *DataPlanePrerequisite) GetConfig() isDataPlanePrerequisite_Config {
	if m != nil {
		return m.Config
	}
	return nil
}

func (x *DataPlanePrerequisite) GetRequiredPlugins() *RequiredPluginsFilter {
	if x, ok := x.GetConfig().(*DataPlanePrerequisite_RequiredPlugins); ok {
		return x.RequiredPlugins
	}
	return nil
}

type isDataPlanePrerequisite_Config interface {
	isDataPlanePrerequisite_Config()
}

type DataPlanePrerequisite_RequiredPlugins struct {
	RequiredPlugins *RequiredPluginsFilter `protobuf:"bytes,2,opt,name=required_plugins,json=requiredPlugins,proto3,oneof"`
}

func (*DataPlanePrerequisite_RequiredPlugins) isDataPlanePrerequisite_Config() {}

type RequiredPluginsFilter struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	RequiredPlugins []string `protobuf:"bytes,1,rep,name=required_plugins,json=requiredPlugins,proto3" json:"required_plugins,omitempty"`
}

func (x *RequiredPluginsFilter) Reset() {
	*x = RequiredPluginsFilter{}
	if protoimpl.UnsafeEnabled {
		mi := &file_kong_util_v1_data_plane_prereq_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RequiredPluginsFilter) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RequiredPluginsFilter) ProtoMessage() {}

func (x *RequiredPluginsFilter) ProtoReflect() protoreflect.Message {
	mi := &file_kong_util_v1_data_plane_prereq_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RequiredPluginsFilter.ProtoReflect.Descriptor instead.
func (*RequiredPluginsFilter) Descriptor() ([]byte, []int) {
	return file_kong_util_v1_data_plane_prereq_proto_rawDescGZIP(), []int{1}
}

func (x *RequiredPluginsFilter) GetRequiredPlugins() []string {
	if x != nil {
		return x.RequiredPlugins
	}
	return nil
}

var File_kong_util_v1_data_plane_prereq_proto protoreflect.FileDescriptor

var file_kong_util_v1_data_plane_prereq_proto_rawDesc = []byte{
	0x0a, 0x24, 0x6b, 0x6f, 0x6e, 0x67, 0x2f, 0x75, 0x74, 0x69, 0x6c, 0x2f, 0x76, 0x31, 0x2f, 0x64,
	0x61, 0x74, 0x61, 0x5f, 0x70, 0x6c, 0x61, 0x6e, 0x65, 0x5f, 0x70, 0x72, 0x65, 0x72, 0x65, 0x71,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0c, 0x6b, 0x6f, 0x6e, 0x67, 0x2e, 0x75, 0x74, 0x69,
	0x6c, 0x2e, 0x76, 0x31, 0x22, 0x73, 0x0a, 0x15, 0x44, 0x61, 0x74, 0x61, 0x50, 0x6c, 0x61, 0x6e,
	0x65, 0x50, 0x72, 0x65, 0x72, 0x65, 0x71, 0x75, 0x69, 0x73, 0x69, 0x74, 0x65, 0x12, 0x50, 0x0a,
	0x10, 0x72, 0x65, 0x71, 0x75, 0x69, 0x72, 0x65, 0x64, 0x5f, 0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e,
	0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x23, 0x2e, 0x6b, 0x6f, 0x6e, 0x67, 0x2e, 0x75,
	0x74, 0x69, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x52, 0x65, 0x71, 0x75, 0x69, 0x72, 0x65, 0x64, 0x50,
	0x6c, 0x75, 0x67, 0x69, 0x6e, 0x73, 0x46, 0x69, 0x6c, 0x74, 0x65, 0x72, 0x48, 0x00, 0x52, 0x0f,
	0x72, 0x65, 0x71, 0x75, 0x69, 0x72, 0x65, 0x64, 0x50, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x73, 0x42,
	0x08, 0x0a, 0x06, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x22, 0x42, 0x0a, 0x15, 0x52, 0x65, 0x71,
	0x75, 0x69, 0x72, 0x65, 0x64, 0x50, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x73, 0x46, 0x69, 0x6c, 0x74,
	0x65, 0x72, 0x12, 0x29, 0x0a, 0x10, 0x72, 0x65, 0x71, 0x75, 0x69, 0x72, 0x65, 0x64, 0x5f, 0x70,
	0x6c, 0x75, 0x67, 0x69, 0x6e, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x09, 0x52, 0x0f, 0x72, 0x65,
	0x71, 0x75, 0x69, 0x72, 0x65, 0x64, 0x50, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x73, 0x42, 0x40, 0x5a,
	0x3e, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6b, 0x6f, 0x6e, 0x67,
	0x2f, 0x6b, 0x6f, 0x6b, 0x6f, 0x2f, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2f, 0x70,
	0x72, 0x69, 0x76, 0x61, 0x74, 0x65, 0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x67, 0x72, 0x70, 0x63, 0x2f,
	0x6b, 0x6f, 0x6e, 0x67, 0x2f, 0x75, 0x74, 0x69, 0x6c, 0x2f, 0x76, 0x31, 0x3b, 0x76, 0x31, 0x62,
	0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_kong_util_v1_data_plane_prereq_proto_rawDescOnce sync.Once
	file_kong_util_v1_data_plane_prereq_proto_rawDescData = file_kong_util_v1_data_plane_prereq_proto_rawDesc
)

func file_kong_util_v1_data_plane_prereq_proto_rawDescGZIP() []byte {
	file_kong_util_v1_data_plane_prereq_proto_rawDescOnce.Do(func() {
		file_kong_util_v1_data_plane_prereq_proto_rawDescData = protoimpl.X.CompressGZIP(file_kong_util_v1_data_plane_prereq_proto_rawDescData)
	})
	return file_kong_util_v1_data_plane_prereq_proto_rawDescData
}

var file_kong_util_v1_data_plane_prereq_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_kong_util_v1_data_plane_prereq_proto_goTypes = []interface{}{
	(*DataPlanePrerequisite)(nil), // 0: kong.util.v1.DataPlanePrerequisite
	(*RequiredPluginsFilter)(nil), // 1: kong.util.v1.RequiredPluginsFilter
}
var file_kong_util_v1_data_plane_prereq_proto_depIdxs = []int32{
	1, // 0: kong.util.v1.DataPlanePrerequisite.required_plugins:type_name -> kong.util.v1.RequiredPluginsFilter
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_kong_util_v1_data_plane_prereq_proto_init() }
func file_kong_util_v1_data_plane_prereq_proto_init() {
	if File_kong_util_v1_data_plane_prereq_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_kong_util_v1_data_plane_prereq_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DataPlanePrerequisite); i {
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
		file_kong_util_v1_data_plane_prereq_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RequiredPluginsFilter); i {
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
	file_kong_util_v1_data_plane_prereq_proto_msgTypes[0].OneofWrappers = []interface{}{
		(*DataPlanePrerequisite_RequiredPlugins)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_kong_util_v1_data_plane_prereq_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_kong_util_v1_data_plane_prereq_proto_goTypes,
		DependencyIndexes: file_kong_util_v1_data_plane_prereq_proto_depIdxs,
		MessageInfos:      file_kong_util_v1_data_plane_prereq_proto_msgTypes,
	}.Build()
	File_kong_util_v1_data_plane_prereq_proto = out.File
	file_kong_util_v1_data_plane_prereq_proto_rawDesc = nil
	file_kong_util_v1_data_plane_prereq_proto_goTypes = nil
	file_kong_util_v1_data_plane_prereq_proto_depIdxs = nil
}
