// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.27.1
// 	protoc        v3.17.3
// source: kong/admin/service/v1/node.proto

package v1

import (
	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
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

type GetNodeRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id      string             `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Cluster *v1.RequestCluster `protobuf:"bytes,2,opt,name=cluster,proto3" json:"cluster,omitempty"`
}

func (x *GetNodeRequest) Reset() {
	*x = GetNodeRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_kong_admin_service_v1_node_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetNodeRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetNodeRequest) ProtoMessage() {}

func (x *GetNodeRequest) ProtoReflect() protoreflect.Message {
	mi := &file_kong_admin_service_v1_node_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetNodeRequest.ProtoReflect.Descriptor instead.
func (*GetNodeRequest) Descriptor() ([]byte, []int) {
	return file_kong_admin_service_v1_node_proto_rawDescGZIP(), []int{0}
}

func (x *GetNodeRequest) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *GetNodeRequest) GetCluster() *v1.RequestCluster {
	if x != nil {
		return x.Cluster
	}
	return nil
}

type GetNodeResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Item *v1.Node `protobuf:"bytes,1,opt,name=item,proto3" json:"item,omitempty"`
}

func (x *GetNodeResponse) Reset() {
	*x = GetNodeResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_kong_admin_service_v1_node_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetNodeResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetNodeResponse) ProtoMessage() {}

func (x *GetNodeResponse) ProtoReflect() protoreflect.Message {
	mi := &file_kong_admin_service_v1_node_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetNodeResponse.ProtoReflect.Descriptor instead.
func (*GetNodeResponse) Descriptor() ([]byte, []int) {
	return file_kong_admin_service_v1_node_proto_rawDescGZIP(), []int{1}
}

func (x *GetNodeResponse) GetItem() *v1.Node {
	if x != nil {
		return x.Item
	}
	return nil
}

type CreateNodeRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Item    *v1.Node           `protobuf:"bytes,1,opt,name=item,proto3" json:"item,omitempty"`
	Cluster *v1.RequestCluster `protobuf:"bytes,2,opt,name=cluster,proto3" json:"cluster,omitempty"`
}

func (x *CreateNodeRequest) Reset() {
	*x = CreateNodeRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_kong_admin_service_v1_node_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CreateNodeRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CreateNodeRequest) ProtoMessage() {}

func (x *CreateNodeRequest) ProtoReflect() protoreflect.Message {
	mi := &file_kong_admin_service_v1_node_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CreateNodeRequest.ProtoReflect.Descriptor instead.
func (*CreateNodeRequest) Descriptor() ([]byte, []int) {
	return file_kong_admin_service_v1_node_proto_rawDescGZIP(), []int{2}
}

func (x *CreateNodeRequest) GetItem() *v1.Node {
	if x != nil {
		return x.Item
	}
	return nil
}

func (x *CreateNodeRequest) GetCluster() *v1.RequestCluster {
	if x != nil {
		return x.Cluster
	}
	return nil
}

type CreateNodeResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Item *v1.Node `protobuf:"bytes,1,opt,name=item,proto3" json:"item,omitempty"`
}

func (x *CreateNodeResponse) Reset() {
	*x = CreateNodeResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_kong_admin_service_v1_node_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CreateNodeResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CreateNodeResponse) ProtoMessage() {}

func (x *CreateNodeResponse) ProtoReflect() protoreflect.Message {
	mi := &file_kong_admin_service_v1_node_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CreateNodeResponse.ProtoReflect.Descriptor instead.
func (*CreateNodeResponse) Descriptor() ([]byte, []int) {
	return file_kong_admin_service_v1_node_proto_rawDescGZIP(), []int{3}
}

func (x *CreateNodeResponse) GetItem() *v1.Node {
	if x != nil {
		return x.Item
	}
	return nil
}

type UpsertNodeRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Item    *v1.Node           `protobuf:"bytes,1,opt,name=item,proto3" json:"item,omitempty"`
	Cluster *v1.RequestCluster `protobuf:"bytes,2,opt,name=cluster,proto3" json:"cluster,omitempty"`
}

func (x *UpsertNodeRequest) Reset() {
	*x = UpsertNodeRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_kong_admin_service_v1_node_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UpsertNodeRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpsertNodeRequest) ProtoMessage() {}

func (x *UpsertNodeRequest) ProtoReflect() protoreflect.Message {
	mi := &file_kong_admin_service_v1_node_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UpsertNodeRequest.ProtoReflect.Descriptor instead.
func (*UpsertNodeRequest) Descriptor() ([]byte, []int) {
	return file_kong_admin_service_v1_node_proto_rawDescGZIP(), []int{4}
}

func (x *UpsertNodeRequest) GetItem() *v1.Node {
	if x != nil {
		return x.Item
	}
	return nil
}

func (x *UpsertNodeRequest) GetCluster() *v1.RequestCluster {
	if x != nil {
		return x.Cluster
	}
	return nil
}

type UpsertNodeResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Item *v1.Node `protobuf:"bytes,1,opt,name=item,proto3" json:"item,omitempty"`
}

func (x *UpsertNodeResponse) Reset() {
	*x = UpsertNodeResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_kong_admin_service_v1_node_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UpsertNodeResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpsertNodeResponse) ProtoMessage() {}

func (x *UpsertNodeResponse) ProtoReflect() protoreflect.Message {
	mi := &file_kong_admin_service_v1_node_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UpsertNodeResponse.ProtoReflect.Descriptor instead.
func (*UpsertNodeResponse) Descriptor() ([]byte, []int) {
	return file_kong_admin_service_v1_node_proto_rawDescGZIP(), []int{5}
}

func (x *UpsertNodeResponse) GetItem() *v1.Node {
	if x != nil {
		return x.Item
	}
	return nil
}

type DeleteNodeRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id      string             `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Cluster *v1.RequestCluster `protobuf:"bytes,2,opt,name=cluster,proto3" json:"cluster,omitempty"`
}

func (x *DeleteNodeRequest) Reset() {
	*x = DeleteNodeRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_kong_admin_service_v1_node_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DeleteNodeRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DeleteNodeRequest) ProtoMessage() {}

func (x *DeleteNodeRequest) ProtoReflect() protoreflect.Message {
	mi := &file_kong_admin_service_v1_node_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DeleteNodeRequest.ProtoReflect.Descriptor instead.
func (*DeleteNodeRequest) Descriptor() ([]byte, []int) {
	return file_kong_admin_service_v1_node_proto_rawDescGZIP(), []int{6}
}

func (x *DeleteNodeRequest) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *DeleteNodeRequest) GetCluster() *v1.RequestCluster {
	if x != nil {
		return x.Cluster
	}
	return nil
}

type DeleteNodeResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *DeleteNodeResponse) Reset() {
	*x = DeleteNodeResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_kong_admin_service_v1_node_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DeleteNodeResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DeleteNodeResponse) ProtoMessage() {}

func (x *DeleteNodeResponse) ProtoReflect() protoreflect.Message {
	mi := &file_kong_admin_service_v1_node_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DeleteNodeResponse.ProtoReflect.Descriptor instead.
func (*DeleteNodeResponse) Descriptor() ([]byte, []int) {
	return file_kong_admin_service_v1_node_proto_rawDescGZIP(), []int{7}
}

type ListNodesRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Cluster *v1.RequestCluster `protobuf:"bytes,1,opt,name=cluster,proto3" json:"cluster,omitempty"`
}

func (x *ListNodesRequest) Reset() {
	*x = ListNodesRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_kong_admin_service_v1_node_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ListNodesRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListNodesRequest) ProtoMessage() {}

func (x *ListNodesRequest) ProtoReflect() protoreflect.Message {
	mi := &file_kong_admin_service_v1_node_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListNodesRequest.ProtoReflect.Descriptor instead.
func (*ListNodesRequest) Descriptor() ([]byte, []int) {
	return file_kong_admin_service_v1_node_proto_rawDescGZIP(), []int{8}
}

func (x *ListNodesRequest) GetCluster() *v1.RequestCluster {
	if x != nil {
		return x.Cluster
	}
	return nil
}

type ListNodesResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Items []*v1.Node `protobuf:"bytes,1,rep,name=items,proto3" json:"items,omitempty"`
}

func (x *ListNodesResponse) Reset() {
	*x = ListNodesResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_kong_admin_service_v1_node_proto_msgTypes[9]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ListNodesResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListNodesResponse) ProtoMessage() {}

func (x *ListNodesResponse) ProtoReflect() protoreflect.Message {
	mi := &file_kong_admin_service_v1_node_proto_msgTypes[9]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListNodesResponse.ProtoReflect.Descriptor instead.
func (*ListNodesResponse) Descriptor() ([]byte, []int) {
	return file_kong_admin_service_v1_node_proto_rawDescGZIP(), []int{9}
}

func (x *ListNodesResponse) GetItems() []*v1.Node {
	if x != nil {
		return x.Items
	}
	return nil
}

var File_kong_admin_service_v1_node_proto protoreflect.FileDescriptor

var file_kong_admin_service_v1_node_proto_rawDesc = []byte{
	0x0a, 0x20, 0x6b, 0x6f, 0x6e, 0x67, 0x2f, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2f, 0x73, 0x65, 0x72,
	0x76, 0x69, 0x63, 0x65, 0x2f, 0x76, 0x31, 0x2f, 0x6e, 0x6f, 0x64, 0x65, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x12, 0x15, 0x6b, 0x6f, 0x6e, 0x67, 0x2e, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2e, 0x73,
	0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x76, 0x31, 0x1a, 0x1c, 0x67, 0x6f, 0x6f, 0x67, 0x6c,
	0x65, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e,
	0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1e, 0x6b, 0x6f, 0x6e, 0x67, 0x2f, 0x61, 0x64,
	0x6d, 0x69, 0x6e, 0x2f, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x2f, 0x76, 0x31, 0x2f, 0x6e, 0x6f, 0x64,
	0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x21, 0x6b, 0x6f, 0x6e, 0x67, 0x2f, 0x61, 0x64,
	0x6d, 0x69, 0x6e, 0x2f, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x2f, 0x76, 0x31, 0x2f, 0x63, 0x6c, 0x75,
	0x73, 0x74, 0x65, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x5f, 0x0a, 0x0e, 0x47, 0x65,
	0x74, 0x4e, 0x6f, 0x64, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x0e, 0x0a, 0x02,
	0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12, 0x3d, 0x0a, 0x07,
	0x63, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x23, 0x2e,
	0x6b, 0x6f, 0x6e, 0x67, 0x2e, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c,
	0x2e, 0x76, 0x31, 0x2e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x43, 0x6c, 0x75, 0x73, 0x74,
	0x65, 0x72, 0x52, 0x07, 0x63, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x22, 0x40, 0x0a, 0x0f, 0x47,
	0x65, 0x74, 0x4e, 0x6f, 0x64, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x2d,
	0x0a, 0x04, 0x69, 0x74, 0x65, 0x6d, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x19, 0x2e, 0x6b,
	0x6f, 0x6e, 0x67, 0x2e, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x2e,
	0x76, 0x31, 0x2e, 0x4e, 0x6f, 0x64, 0x65, 0x52, 0x04, 0x69, 0x74, 0x65, 0x6d, 0x22, 0x81, 0x01,
	0x0a, 0x11, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x4e, 0x6f, 0x64, 0x65, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x12, 0x2d, 0x0a, 0x04, 0x69, 0x74, 0x65, 0x6d, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x19, 0x2e, 0x6b, 0x6f, 0x6e, 0x67, 0x2e, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2e, 0x6d,
	0x6f, 0x64, 0x65, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x4e, 0x6f, 0x64, 0x65, 0x52, 0x04, 0x69, 0x74,
	0x65, 0x6d, 0x12, 0x3d, 0x0a, 0x07, 0x63, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x23, 0x2e, 0x6b, 0x6f, 0x6e, 0x67, 0x2e, 0x61, 0x64, 0x6d, 0x69, 0x6e,
	0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x43, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x52, 0x07, 0x63, 0x6c, 0x75, 0x73, 0x74, 0x65,
	0x72, 0x22, 0x43, 0x0a, 0x12, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x4e, 0x6f, 0x64, 0x65, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x2d, 0x0a, 0x04, 0x69, 0x74, 0x65, 0x6d, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x19, 0x2e, 0x6b, 0x6f, 0x6e, 0x67, 0x2e, 0x61, 0x64, 0x6d,
	0x69, 0x6e, 0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x4e, 0x6f, 0x64, 0x65,
	0x52, 0x04, 0x69, 0x74, 0x65, 0x6d, 0x22, 0x81, 0x01, 0x0a, 0x11, 0x55, 0x70, 0x73, 0x65, 0x72,
	0x74, 0x4e, 0x6f, 0x64, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x2d, 0x0a, 0x04,
	0x69, 0x74, 0x65, 0x6d, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x19, 0x2e, 0x6b, 0x6f, 0x6e,
	0x67, 0x2e, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x2e, 0x76, 0x31,
	0x2e, 0x4e, 0x6f, 0x64, 0x65, 0x52, 0x04, 0x69, 0x74, 0x65, 0x6d, 0x12, 0x3d, 0x0a, 0x07, 0x63,
	0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x23, 0x2e, 0x6b,
	0x6f, 0x6e, 0x67, 0x2e, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x2e,
	0x76, 0x31, 0x2e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x43, 0x6c, 0x75, 0x73, 0x74, 0x65,
	0x72, 0x52, 0x07, 0x63, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x22, 0x43, 0x0a, 0x12, 0x55, 0x70,
	0x73, 0x65, 0x72, 0x74, 0x4e, 0x6f, 0x64, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x12, 0x2d, 0x0a, 0x04, 0x69, 0x74, 0x65, 0x6d, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x19,
	0x2e, 0x6b, 0x6f, 0x6e, 0x67, 0x2e, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2e, 0x6d, 0x6f, 0x64, 0x65,
	0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x4e, 0x6f, 0x64, 0x65, 0x52, 0x04, 0x69, 0x74, 0x65, 0x6d, 0x22,
	0x62, 0x0a, 0x11, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x4e, 0x6f, 0x64, 0x65, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x02, 0x69, 0x64, 0x12, 0x3d, 0x0a, 0x07, 0x63, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x23, 0x2e, 0x6b, 0x6f, 0x6e, 0x67, 0x2e, 0x61, 0x64, 0x6d,
	0x69, 0x6e, 0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x43, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x52, 0x07, 0x63, 0x6c, 0x75, 0x73,
	0x74, 0x65, 0x72, 0x22, 0x14, 0x0a, 0x12, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x4e, 0x6f, 0x64,
	0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x51, 0x0a, 0x10, 0x4c, 0x69, 0x73,
	0x74, 0x4e, 0x6f, 0x64, 0x65, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x3d, 0x0a,
	0x07, 0x63, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x23,
	0x2e, 0x6b, 0x6f, 0x6e, 0x67, 0x2e, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2e, 0x6d, 0x6f, 0x64, 0x65,
	0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x43, 0x6c, 0x75, 0x73,
	0x74, 0x65, 0x72, 0x52, 0x07, 0x63, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x22, 0x44, 0x0a, 0x11,
	0x4c, 0x69, 0x73, 0x74, 0x4e, 0x6f, 0x64, 0x65, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x12, 0x2f, 0x0a, 0x05, 0x69, 0x74, 0x65, 0x6d, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b,
	0x32, 0x19, 0x2e, 0x6b, 0x6f, 0x6e, 0x67, 0x2e, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2e, 0x6d, 0x6f,
	0x64, 0x65, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x4e, 0x6f, 0x64, 0x65, 0x52, 0x05, 0x69, 0x74, 0x65,
	0x6d, 0x73, 0x32, 0xb3, 0x04, 0x0a, 0x0b, 0x4e, 0x6f, 0x64, 0x65, 0x53, 0x65, 0x72, 0x76, 0x69,
	0x63, 0x65, 0x12, 0x70, 0x0a, 0x07, 0x47, 0x65, 0x74, 0x4e, 0x6f, 0x64, 0x65, 0x12, 0x25, 0x2e,
	0x6b, 0x6f, 0x6e, 0x67, 0x2e, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x69,
	0x63, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x47, 0x65, 0x74, 0x4e, 0x6f, 0x64, 0x65, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x1a, 0x26, 0x2e, 0x6b, 0x6f, 0x6e, 0x67, 0x2e, 0x61, 0x64, 0x6d, 0x69,
	0x6e, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x47, 0x65, 0x74,
	0x4e, 0x6f, 0x64, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x16, 0x82, 0xd3,
	0xe4, 0x93, 0x02, 0x10, 0x12, 0x0e, 0x2f, 0x76, 0x31, 0x2f, 0x6e, 0x6f, 0x64, 0x65, 0x73, 0x2f,
	0x7b, 0x69, 0x64, 0x7d, 0x12, 0x61, 0x0a, 0x0a, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x4e, 0x6f,
	0x64, 0x65, 0x12, 0x28, 0x2e, 0x6b, 0x6f, 0x6e, 0x67, 0x2e, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2e,
	0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x43, 0x72, 0x65, 0x61, 0x74,
	0x65, 0x4e, 0x6f, 0x64, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x29, 0x2e, 0x6b,
	0x6f, 0x6e, 0x67, 0x2e, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63,
	0x65, 0x2e, 0x76, 0x31, 0x2e, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x4e, 0x6f, 0x64, 0x65, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x61, 0x0a, 0x0a, 0x55, 0x70, 0x73, 0x65, 0x72,
	0x74, 0x4e, 0x6f, 0x64, 0x65, 0x12, 0x28, 0x2e, 0x6b, 0x6f, 0x6e, 0x67, 0x2e, 0x61, 0x64, 0x6d,
	0x69, 0x6e, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x55, 0x70,
	0x73, 0x65, 0x72, 0x74, 0x4e, 0x6f, 0x64, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a,
	0x29, 0x2e, 0x6b, 0x6f, 0x6e, 0x67, 0x2e, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2e, 0x73, 0x65, 0x72,
	0x76, 0x69, 0x63, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x55, 0x70, 0x73, 0x65, 0x72, 0x74, 0x4e, 0x6f,
	0x64, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x79, 0x0a, 0x0a, 0x44, 0x65,
	0x6c, 0x65, 0x74, 0x65, 0x4e, 0x6f, 0x64, 0x65, 0x12, 0x28, 0x2e, 0x6b, 0x6f, 0x6e, 0x67, 0x2e,
	0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x76, 0x31,
	0x2e, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x4e, 0x6f, 0x64, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x1a, 0x29, 0x2e, 0x6b, 0x6f, 0x6e, 0x67, 0x2e, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2e,
	0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x44, 0x65, 0x6c, 0x65, 0x74,
	0x65, 0x4e, 0x6f, 0x64, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x16, 0x82,
	0xd3, 0xe4, 0x93, 0x02, 0x10, 0x2a, 0x0e, 0x2f, 0x76, 0x31, 0x2f, 0x6e, 0x6f, 0x64, 0x65, 0x73,
	0x2f, 0x7b, 0x69, 0x64, 0x7d, 0x12, 0x71, 0x0a, 0x09, 0x4c, 0x69, 0x73, 0x74, 0x4e, 0x6f, 0x64,
	0x65, 0x73, 0x12, 0x27, 0x2e, 0x6b, 0x6f, 0x6e, 0x67, 0x2e, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2e,
	0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x4c, 0x69, 0x73, 0x74, 0x4e,
	0x6f, 0x64, 0x65, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x28, 0x2e, 0x6b, 0x6f,
	0x6e, 0x67, 0x2e, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65,
	0x2e, 0x76, 0x31, 0x2e, 0x4c, 0x69, 0x73, 0x74, 0x4e, 0x6f, 0x64, 0x65, 0x73, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x11, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x0b, 0x12, 0x09, 0x2f,
	0x76, 0x31, 0x2f, 0x6e, 0x6f, 0x64, 0x65, 0x73, 0x42, 0x3c, 0x5a, 0x3a, 0x67, 0x69, 0x74, 0x68,
	0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6b, 0x6f, 0x6e, 0x67, 0x2f, 0x6b, 0x6f, 0x6b, 0x6f,
	0x2f, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x6b, 0x6f,
	0x6e, 0x67, 0x2f, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2f, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65,
	0x2f, 0x76, 0x31, 0x3b, 0x76, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_kong_admin_service_v1_node_proto_rawDescOnce sync.Once
	file_kong_admin_service_v1_node_proto_rawDescData = file_kong_admin_service_v1_node_proto_rawDesc
)

func file_kong_admin_service_v1_node_proto_rawDescGZIP() []byte {
	file_kong_admin_service_v1_node_proto_rawDescOnce.Do(func() {
		file_kong_admin_service_v1_node_proto_rawDescData = protoimpl.X.CompressGZIP(file_kong_admin_service_v1_node_proto_rawDescData)
	})
	return file_kong_admin_service_v1_node_proto_rawDescData
}

var file_kong_admin_service_v1_node_proto_msgTypes = make([]protoimpl.MessageInfo, 10)
var file_kong_admin_service_v1_node_proto_goTypes = []interface{}{
	(*GetNodeRequest)(nil),     // 0: kong.admin.service.v1.GetNodeRequest
	(*GetNodeResponse)(nil),    // 1: kong.admin.service.v1.GetNodeResponse
	(*CreateNodeRequest)(nil),  // 2: kong.admin.service.v1.CreateNodeRequest
	(*CreateNodeResponse)(nil), // 3: kong.admin.service.v1.CreateNodeResponse
	(*UpsertNodeRequest)(nil),  // 4: kong.admin.service.v1.UpsertNodeRequest
	(*UpsertNodeResponse)(nil), // 5: kong.admin.service.v1.UpsertNodeResponse
	(*DeleteNodeRequest)(nil),  // 6: kong.admin.service.v1.DeleteNodeRequest
	(*DeleteNodeResponse)(nil), // 7: kong.admin.service.v1.DeleteNodeResponse
	(*ListNodesRequest)(nil),   // 8: kong.admin.service.v1.ListNodesRequest
	(*ListNodesResponse)(nil),  // 9: kong.admin.service.v1.ListNodesResponse
	(*v1.RequestCluster)(nil),  // 10: kong.admin.model.v1.RequestCluster
	(*v1.Node)(nil),            // 11: kong.admin.model.v1.Node
}
var file_kong_admin_service_v1_node_proto_depIdxs = []int32{
	10, // 0: kong.admin.service.v1.GetNodeRequest.cluster:type_name -> kong.admin.model.v1.RequestCluster
	11, // 1: kong.admin.service.v1.GetNodeResponse.item:type_name -> kong.admin.model.v1.Node
	11, // 2: kong.admin.service.v1.CreateNodeRequest.item:type_name -> kong.admin.model.v1.Node
	10, // 3: kong.admin.service.v1.CreateNodeRequest.cluster:type_name -> kong.admin.model.v1.RequestCluster
	11, // 4: kong.admin.service.v1.CreateNodeResponse.item:type_name -> kong.admin.model.v1.Node
	11, // 5: kong.admin.service.v1.UpsertNodeRequest.item:type_name -> kong.admin.model.v1.Node
	10, // 6: kong.admin.service.v1.UpsertNodeRequest.cluster:type_name -> kong.admin.model.v1.RequestCluster
	11, // 7: kong.admin.service.v1.UpsertNodeResponse.item:type_name -> kong.admin.model.v1.Node
	10, // 8: kong.admin.service.v1.DeleteNodeRequest.cluster:type_name -> kong.admin.model.v1.RequestCluster
	10, // 9: kong.admin.service.v1.ListNodesRequest.cluster:type_name -> kong.admin.model.v1.RequestCluster
	11, // 10: kong.admin.service.v1.ListNodesResponse.items:type_name -> kong.admin.model.v1.Node
	0,  // 11: kong.admin.service.v1.NodeService.GetNode:input_type -> kong.admin.service.v1.GetNodeRequest
	2,  // 12: kong.admin.service.v1.NodeService.CreateNode:input_type -> kong.admin.service.v1.CreateNodeRequest
	4,  // 13: kong.admin.service.v1.NodeService.UpsertNode:input_type -> kong.admin.service.v1.UpsertNodeRequest
	6,  // 14: kong.admin.service.v1.NodeService.DeleteNode:input_type -> kong.admin.service.v1.DeleteNodeRequest
	8,  // 15: kong.admin.service.v1.NodeService.ListNodes:input_type -> kong.admin.service.v1.ListNodesRequest
	1,  // 16: kong.admin.service.v1.NodeService.GetNode:output_type -> kong.admin.service.v1.GetNodeResponse
	3,  // 17: kong.admin.service.v1.NodeService.CreateNode:output_type -> kong.admin.service.v1.CreateNodeResponse
	5,  // 18: kong.admin.service.v1.NodeService.UpsertNode:output_type -> kong.admin.service.v1.UpsertNodeResponse
	7,  // 19: kong.admin.service.v1.NodeService.DeleteNode:output_type -> kong.admin.service.v1.DeleteNodeResponse
	9,  // 20: kong.admin.service.v1.NodeService.ListNodes:output_type -> kong.admin.service.v1.ListNodesResponse
	16, // [16:21] is the sub-list for method output_type
	11, // [11:16] is the sub-list for method input_type
	11, // [11:11] is the sub-list for extension type_name
	11, // [11:11] is the sub-list for extension extendee
	0,  // [0:11] is the sub-list for field type_name
}

func init() { file_kong_admin_service_v1_node_proto_init() }
func file_kong_admin_service_v1_node_proto_init() {
	if File_kong_admin_service_v1_node_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_kong_admin_service_v1_node_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetNodeRequest); i {
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
		file_kong_admin_service_v1_node_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetNodeResponse); i {
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
		file_kong_admin_service_v1_node_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CreateNodeRequest); i {
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
		file_kong_admin_service_v1_node_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CreateNodeResponse); i {
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
		file_kong_admin_service_v1_node_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UpsertNodeRequest); i {
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
		file_kong_admin_service_v1_node_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UpsertNodeResponse); i {
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
		file_kong_admin_service_v1_node_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DeleteNodeRequest); i {
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
		file_kong_admin_service_v1_node_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DeleteNodeResponse); i {
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
		file_kong_admin_service_v1_node_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ListNodesRequest); i {
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
		file_kong_admin_service_v1_node_proto_msgTypes[9].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ListNodesResponse); i {
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
			RawDescriptor: file_kong_admin_service_v1_node_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   10,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_kong_admin_service_v1_node_proto_goTypes,
		DependencyIndexes: file_kong_admin_service_v1_node_proto_depIdxs,
		MessageInfos:      file_kong_admin_service_v1_node_proto_msgTypes,
	}.Build()
	File_kong_admin_service_v1_node_proto = out.File
	file_kong_admin_service_v1_node_proto_rawDesc = nil
	file_kong_admin_service_v1_node_proto_goTypes = nil
	file_kong_admin_service_v1_node_proto_depIdxs = nil
}
