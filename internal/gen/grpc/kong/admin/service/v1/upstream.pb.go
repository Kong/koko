// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.0
// 	protoc        (unknown)
// source: kong/admin/service/v1/upstream.proto

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

type GetUpstreamRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id      string             `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Cluster *v1.RequestCluster `protobuf:"bytes,2,opt,name=cluster,proto3" json:"cluster,omitempty"`
}

func (x *GetUpstreamRequest) Reset() {
	*x = GetUpstreamRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_kong_admin_service_v1_upstream_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetUpstreamRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetUpstreamRequest) ProtoMessage() {}

func (x *GetUpstreamRequest) ProtoReflect() protoreflect.Message {
	mi := &file_kong_admin_service_v1_upstream_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetUpstreamRequest.ProtoReflect.Descriptor instead.
func (*GetUpstreamRequest) Descriptor() ([]byte, []int) {
	return file_kong_admin_service_v1_upstream_proto_rawDescGZIP(), []int{0}
}

func (x *GetUpstreamRequest) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *GetUpstreamRequest) GetCluster() *v1.RequestCluster {
	if x != nil {
		return x.Cluster
	}
	return nil
}

type GetUpstreamResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Item *v1.Upstream `protobuf:"bytes,1,opt,name=item,proto3" json:"item,omitempty"`
}

func (x *GetUpstreamResponse) Reset() {
	*x = GetUpstreamResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_kong_admin_service_v1_upstream_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetUpstreamResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetUpstreamResponse) ProtoMessage() {}

func (x *GetUpstreamResponse) ProtoReflect() protoreflect.Message {
	mi := &file_kong_admin_service_v1_upstream_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetUpstreamResponse.ProtoReflect.Descriptor instead.
func (*GetUpstreamResponse) Descriptor() ([]byte, []int) {
	return file_kong_admin_service_v1_upstream_proto_rawDescGZIP(), []int{1}
}

func (x *GetUpstreamResponse) GetItem() *v1.Upstream {
	if x != nil {
		return x.Item
	}
	return nil
}

type CreateUpstreamRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Item    *v1.Upstream       `protobuf:"bytes,1,opt,name=item,proto3" json:"item,omitempty"`
	Cluster *v1.RequestCluster `protobuf:"bytes,2,opt,name=cluster,proto3" json:"cluster,omitempty"`
}

func (x *CreateUpstreamRequest) Reset() {
	*x = CreateUpstreamRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_kong_admin_service_v1_upstream_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CreateUpstreamRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CreateUpstreamRequest) ProtoMessage() {}

func (x *CreateUpstreamRequest) ProtoReflect() protoreflect.Message {
	mi := &file_kong_admin_service_v1_upstream_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CreateUpstreamRequest.ProtoReflect.Descriptor instead.
func (*CreateUpstreamRequest) Descriptor() ([]byte, []int) {
	return file_kong_admin_service_v1_upstream_proto_rawDescGZIP(), []int{2}
}

func (x *CreateUpstreamRequest) GetItem() *v1.Upstream {
	if x != nil {
		return x.Item
	}
	return nil
}

func (x *CreateUpstreamRequest) GetCluster() *v1.RequestCluster {
	if x != nil {
		return x.Cluster
	}
	return nil
}

type CreateUpstreamResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Item *v1.Upstream `protobuf:"bytes,1,opt,name=item,proto3" json:"item,omitempty"`
}

func (x *CreateUpstreamResponse) Reset() {
	*x = CreateUpstreamResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_kong_admin_service_v1_upstream_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CreateUpstreamResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CreateUpstreamResponse) ProtoMessage() {}

func (x *CreateUpstreamResponse) ProtoReflect() protoreflect.Message {
	mi := &file_kong_admin_service_v1_upstream_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CreateUpstreamResponse.ProtoReflect.Descriptor instead.
func (*CreateUpstreamResponse) Descriptor() ([]byte, []int) {
	return file_kong_admin_service_v1_upstream_proto_rawDescGZIP(), []int{3}
}

func (x *CreateUpstreamResponse) GetItem() *v1.Upstream {
	if x != nil {
		return x.Item
	}
	return nil
}

type UpsertUpstreamRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Item    *v1.Upstream       `protobuf:"bytes,1,opt,name=item,proto3" json:"item,omitempty"`
	Cluster *v1.RequestCluster `protobuf:"bytes,2,opt,name=cluster,proto3" json:"cluster,omitempty"`
}

func (x *UpsertUpstreamRequest) Reset() {
	*x = UpsertUpstreamRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_kong_admin_service_v1_upstream_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UpsertUpstreamRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpsertUpstreamRequest) ProtoMessage() {}

func (x *UpsertUpstreamRequest) ProtoReflect() protoreflect.Message {
	mi := &file_kong_admin_service_v1_upstream_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UpsertUpstreamRequest.ProtoReflect.Descriptor instead.
func (*UpsertUpstreamRequest) Descriptor() ([]byte, []int) {
	return file_kong_admin_service_v1_upstream_proto_rawDescGZIP(), []int{4}
}

func (x *UpsertUpstreamRequest) GetItem() *v1.Upstream {
	if x != nil {
		return x.Item
	}
	return nil
}

func (x *UpsertUpstreamRequest) GetCluster() *v1.RequestCluster {
	if x != nil {
		return x.Cluster
	}
	return nil
}

type UpsertUpstreamResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Item *v1.Upstream `protobuf:"bytes,1,opt,name=item,proto3" json:"item,omitempty"`
}

func (x *UpsertUpstreamResponse) Reset() {
	*x = UpsertUpstreamResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_kong_admin_service_v1_upstream_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UpsertUpstreamResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpsertUpstreamResponse) ProtoMessage() {}

func (x *UpsertUpstreamResponse) ProtoReflect() protoreflect.Message {
	mi := &file_kong_admin_service_v1_upstream_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UpsertUpstreamResponse.ProtoReflect.Descriptor instead.
func (*UpsertUpstreamResponse) Descriptor() ([]byte, []int) {
	return file_kong_admin_service_v1_upstream_proto_rawDescGZIP(), []int{5}
}

func (x *UpsertUpstreamResponse) GetItem() *v1.Upstream {
	if x != nil {
		return x.Item
	}
	return nil
}

type DeleteUpstreamRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id      string             `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Cluster *v1.RequestCluster `protobuf:"bytes,2,opt,name=cluster,proto3" json:"cluster,omitempty"`
}

func (x *DeleteUpstreamRequest) Reset() {
	*x = DeleteUpstreamRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_kong_admin_service_v1_upstream_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DeleteUpstreamRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DeleteUpstreamRequest) ProtoMessage() {}

func (x *DeleteUpstreamRequest) ProtoReflect() protoreflect.Message {
	mi := &file_kong_admin_service_v1_upstream_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DeleteUpstreamRequest.ProtoReflect.Descriptor instead.
func (*DeleteUpstreamRequest) Descriptor() ([]byte, []int) {
	return file_kong_admin_service_v1_upstream_proto_rawDescGZIP(), []int{6}
}

func (x *DeleteUpstreamRequest) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *DeleteUpstreamRequest) GetCluster() *v1.RequestCluster {
	if x != nil {
		return x.Cluster
	}
	return nil
}

type DeleteUpstreamResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *DeleteUpstreamResponse) Reset() {
	*x = DeleteUpstreamResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_kong_admin_service_v1_upstream_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DeleteUpstreamResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DeleteUpstreamResponse) ProtoMessage() {}

func (x *DeleteUpstreamResponse) ProtoReflect() protoreflect.Message {
	mi := &file_kong_admin_service_v1_upstream_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DeleteUpstreamResponse.ProtoReflect.Descriptor instead.
func (*DeleteUpstreamResponse) Descriptor() ([]byte, []int) {
	return file_kong_admin_service_v1_upstream_proto_rawDescGZIP(), []int{7}
}

type ListUpstreamsRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Cluster *v1.RequestCluster    `protobuf:"bytes,1,opt,name=cluster,proto3" json:"cluster,omitempty"`
	Page    *v1.PaginationRequest `protobuf:"bytes,2,opt,name=page,proto3" json:"page,omitempty"`
}

func (x *ListUpstreamsRequest) Reset() {
	*x = ListUpstreamsRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_kong_admin_service_v1_upstream_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ListUpstreamsRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListUpstreamsRequest) ProtoMessage() {}

func (x *ListUpstreamsRequest) ProtoReflect() protoreflect.Message {
	mi := &file_kong_admin_service_v1_upstream_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListUpstreamsRequest.ProtoReflect.Descriptor instead.
func (*ListUpstreamsRequest) Descriptor() ([]byte, []int) {
	return file_kong_admin_service_v1_upstream_proto_rawDescGZIP(), []int{8}
}

func (x *ListUpstreamsRequest) GetCluster() *v1.RequestCluster {
	if x != nil {
		return x.Cluster
	}
	return nil
}

func (x *ListUpstreamsRequest) GetPage() *v1.PaginationRequest {
	if x != nil {
		return x.Page
	}
	return nil
}

type ListUpstreamsResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Items []*v1.Upstream         `protobuf:"bytes,1,rep,name=items,proto3" json:"items,omitempty"`
	Page  *v1.PaginationResponse `protobuf:"bytes,2,opt,name=page,proto3" json:"page,omitempty"`
}

func (x *ListUpstreamsResponse) Reset() {
	*x = ListUpstreamsResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_kong_admin_service_v1_upstream_proto_msgTypes[9]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ListUpstreamsResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListUpstreamsResponse) ProtoMessage() {}

func (x *ListUpstreamsResponse) ProtoReflect() protoreflect.Message {
	mi := &file_kong_admin_service_v1_upstream_proto_msgTypes[9]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListUpstreamsResponse.ProtoReflect.Descriptor instead.
func (*ListUpstreamsResponse) Descriptor() ([]byte, []int) {
	return file_kong_admin_service_v1_upstream_proto_rawDescGZIP(), []int{9}
}

func (x *ListUpstreamsResponse) GetItems() []*v1.Upstream {
	if x != nil {
		return x.Items
	}
	return nil
}

func (x *ListUpstreamsResponse) GetPage() *v1.PaginationResponse {
	if x != nil {
		return x.Page
	}
	return nil
}

var File_kong_admin_service_v1_upstream_proto protoreflect.FileDescriptor

var file_kong_admin_service_v1_upstream_proto_rawDesc = []byte{
	0x0a, 0x24, 0x6b, 0x6f, 0x6e, 0x67, 0x2f, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2f, 0x73, 0x65, 0x72,
	0x76, 0x69, 0x63, 0x65, 0x2f, 0x76, 0x31, 0x2f, 0x75, 0x70, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x15, 0x6b, 0x6f, 0x6e, 0x67, 0x2e, 0x61, 0x64, 0x6d,
	0x69, 0x6e, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x76, 0x31, 0x1a, 0x1c, 0x67,
	0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61,
	0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x21, 0x6b, 0x6f, 0x6e,
	0x67, 0x2f, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2f, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x2f, 0x76, 0x31,
	0x2f, 0x63, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x24,
	0x6b, 0x6f, 0x6e, 0x67, 0x2f, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2f, 0x6d, 0x6f, 0x64, 0x65, 0x6c,
	0x2f, 0x76, 0x31, 0x2f, 0x70, 0x61, 0x67, 0x69, 0x6e, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x22, 0x6b, 0x6f, 0x6e, 0x67, 0x2f, 0x61, 0x64, 0x6d, 0x69, 0x6e,
	0x2f, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x2f, 0x76, 0x31, 0x2f, 0x75, 0x70, 0x73, 0x74, 0x72, 0x65,
	0x61, 0x6d, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x63, 0x0a, 0x12, 0x47, 0x65, 0x74, 0x55,
	0x70, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x0e,
	0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12, 0x3d,
	0x0a, 0x07, 0x63, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x23, 0x2e, 0x6b, 0x6f, 0x6e, 0x67, 0x2e, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2e, 0x6d, 0x6f, 0x64,
	0x65, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x43, 0x6c, 0x75,
	0x73, 0x74, 0x65, 0x72, 0x52, 0x07, 0x63, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x22, 0x48, 0x0a,
	0x13, 0x47, 0x65, 0x74, 0x55, 0x70, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x12, 0x31, 0x0a, 0x04, 0x69, 0x74, 0x65, 0x6d, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x1d, 0x2e, 0x6b, 0x6f, 0x6e, 0x67, 0x2e, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2e,
	0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x55, 0x70, 0x73, 0x74, 0x72, 0x65, 0x61,
	0x6d, 0x52, 0x04, 0x69, 0x74, 0x65, 0x6d, 0x22, 0x89, 0x01, 0x0a, 0x15, 0x43, 0x72, 0x65, 0x61,
	0x74, 0x65, 0x55, 0x70, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x12, 0x31, 0x0a, 0x04, 0x69, 0x74, 0x65, 0x6d, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x1d, 0x2e, 0x6b, 0x6f, 0x6e, 0x67, 0x2e, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2e, 0x6d, 0x6f, 0x64,
	0x65, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x55, 0x70, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x52, 0x04,
	0x69, 0x74, 0x65, 0x6d, 0x12, 0x3d, 0x0a, 0x07, 0x63, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x23, 0x2e, 0x6b, 0x6f, 0x6e, 0x67, 0x2e, 0x61, 0x64, 0x6d,
	0x69, 0x6e, 0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x43, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x52, 0x07, 0x63, 0x6c, 0x75, 0x73,
	0x74, 0x65, 0x72, 0x22, 0x4b, 0x0a, 0x16, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x55, 0x70, 0x73,
	0x74, 0x72, 0x65, 0x61, 0x6d, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x31, 0x0a,
	0x04, 0x69, 0x74, 0x65, 0x6d, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1d, 0x2e, 0x6b, 0x6f,
	0x6e, 0x67, 0x2e, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x2e, 0x76,
	0x31, 0x2e, 0x55, 0x70, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x52, 0x04, 0x69, 0x74, 0x65, 0x6d,
	0x22, 0x89, 0x01, 0x0a, 0x15, 0x55, 0x70, 0x73, 0x65, 0x72, 0x74, 0x55, 0x70, 0x73, 0x74, 0x72,
	0x65, 0x61, 0x6d, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x31, 0x0a, 0x04, 0x69, 0x74,
	0x65, 0x6d, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1d, 0x2e, 0x6b, 0x6f, 0x6e, 0x67, 0x2e,
	0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x55,
	0x70, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x52, 0x04, 0x69, 0x74, 0x65, 0x6d, 0x12, 0x3d, 0x0a,
	0x07, 0x63, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x23,
	0x2e, 0x6b, 0x6f, 0x6e, 0x67, 0x2e, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2e, 0x6d, 0x6f, 0x64, 0x65,
	0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x43, 0x6c, 0x75, 0x73,
	0x74, 0x65, 0x72, 0x52, 0x07, 0x63, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x22, 0x4b, 0x0a, 0x16,
	0x55, 0x70, 0x73, 0x65, 0x72, 0x74, 0x55, 0x70, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x31, 0x0a, 0x04, 0x69, 0x74, 0x65, 0x6d, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x1d, 0x2e, 0x6b, 0x6f, 0x6e, 0x67, 0x2e, 0x61, 0x64, 0x6d, 0x69,
	0x6e, 0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x55, 0x70, 0x73, 0x74, 0x72,
	0x65, 0x61, 0x6d, 0x52, 0x04, 0x69, 0x74, 0x65, 0x6d, 0x22, 0x66, 0x0a, 0x15, 0x44, 0x65, 0x6c,
	0x65, 0x74, 0x65, 0x55, 0x70, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02,
	0x69, 0x64, 0x12, 0x3d, 0x0a, 0x07, 0x63, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x23, 0x2e, 0x6b, 0x6f, 0x6e, 0x67, 0x2e, 0x61, 0x64, 0x6d, 0x69, 0x6e,
	0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x43, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x52, 0x07, 0x63, 0x6c, 0x75, 0x73, 0x74, 0x65,
	0x72, 0x22, 0x18, 0x0a, 0x16, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x55, 0x70, 0x73, 0x74, 0x72,
	0x65, 0x61, 0x6d, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x91, 0x01, 0x0a, 0x14,
	0x4c, 0x69, 0x73, 0x74, 0x55, 0x70, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x73, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x12, 0x3d, 0x0a, 0x07, 0x63, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x23, 0x2e, 0x6b, 0x6f, 0x6e, 0x67, 0x2e, 0x61, 0x64, 0x6d,
	0x69, 0x6e, 0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x43, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x52, 0x07, 0x63, 0x6c, 0x75, 0x73,
	0x74, 0x65, 0x72, 0x12, 0x3a, 0x0a, 0x04, 0x70, 0x61, 0x67, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x26, 0x2e, 0x6b, 0x6f, 0x6e, 0x67, 0x2e, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2e, 0x6d,
	0x6f, 0x64, 0x65, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x50, 0x61, 0x67, 0x69, 0x6e, 0x61, 0x74, 0x69,
	0x6f, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x52, 0x04, 0x70, 0x61, 0x67, 0x65, 0x22,
	0x89, 0x01, 0x0a, 0x15, 0x4c, 0x69, 0x73, 0x74, 0x55, 0x70, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d,
	0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x33, 0x0a, 0x05, 0x69, 0x74, 0x65,
	0x6d, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1d, 0x2e, 0x6b, 0x6f, 0x6e, 0x67, 0x2e,
	0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x55,
	0x70, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x52, 0x05, 0x69, 0x74, 0x65, 0x6d, 0x73, 0x12, 0x3b,
	0x0a, 0x04, 0x70, 0x61, 0x67, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x27, 0x2e, 0x6b,
	0x6f, 0x6e, 0x67, 0x2e, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x2e,
	0x76, 0x31, 0x2e, 0x50, 0x61, 0x67, 0x69, 0x6e, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x52, 0x04, 0x70, 0x61, 0x67, 0x65, 0x32, 0xc8, 0x05, 0x0a, 0x0f,
	0x55, 0x70, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12,
	0x80, 0x01, 0x0a, 0x0b, 0x47, 0x65, 0x74, 0x55, 0x70, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x12,
	0x29, 0x2e, 0x6b, 0x6f, 0x6e, 0x67, 0x2e, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2e, 0x73, 0x65, 0x72,
	0x76, 0x69, 0x63, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x47, 0x65, 0x74, 0x55, 0x70, 0x73, 0x74, 0x72,
	0x65, 0x61, 0x6d, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x2a, 0x2e, 0x6b, 0x6f, 0x6e,
	0x67, 0x2e, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2e,
	0x76, 0x31, 0x2e, 0x47, 0x65, 0x74, 0x55, 0x70, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x1a, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x14, 0x12, 0x12,
	0x2f, 0x76, 0x31, 0x2f, 0x75, 0x70, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x73, 0x2f, 0x7b, 0x69,
	0x64, 0x7d, 0x12, 0x8a, 0x01, 0x0a, 0x0e, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x55, 0x70, 0x73,
	0x74, 0x72, 0x65, 0x61, 0x6d, 0x12, 0x2c, 0x2e, 0x6b, 0x6f, 0x6e, 0x67, 0x2e, 0x61, 0x64, 0x6d,
	0x69, 0x6e, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x43, 0x72,
	0x65, 0x61, 0x74, 0x65, 0x55, 0x70, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x1a, 0x2d, 0x2e, 0x6b, 0x6f, 0x6e, 0x67, 0x2e, 0x61, 0x64, 0x6d, 0x69, 0x6e,
	0x2e, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x43, 0x72, 0x65, 0x61,
	0x74, 0x65, 0x55, 0x70, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x22, 0x1b, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x15, 0x22, 0x0d, 0x2f, 0x76, 0x31, 0x2f,
	0x75, 0x70, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x73, 0x3a, 0x04, 0x69, 0x74, 0x65, 0x6d, 0x12,
	0x94, 0x01, 0x0a, 0x0e, 0x55, 0x70, 0x73, 0x65, 0x72, 0x74, 0x55, 0x70, 0x73, 0x74, 0x72, 0x65,
	0x61, 0x6d, 0x12, 0x2c, 0x2e, 0x6b, 0x6f, 0x6e, 0x67, 0x2e, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2e,
	0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x55, 0x70, 0x73, 0x65, 0x72,
	0x74, 0x55, 0x70, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x1a, 0x2d, 0x2e, 0x6b, 0x6f, 0x6e, 0x67, 0x2e, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2e, 0x73, 0x65,
	0x72, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x55, 0x70, 0x73, 0x65, 0x72, 0x74, 0x55,
	0x70, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22,
	0x25, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x1f, 0x1a, 0x17, 0x2f, 0x76, 0x31, 0x2f, 0x75, 0x70, 0x73,
	0x74, 0x72, 0x65, 0x61, 0x6d, 0x73, 0x2f, 0x7b, 0x69, 0x74, 0x65, 0x6d, 0x2e, 0x69, 0x64, 0x7d,
	0x3a, 0x04, 0x69, 0x74, 0x65, 0x6d, 0x12, 0x89, 0x01, 0x0a, 0x0e, 0x44, 0x65, 0x6c, 0x65, 0x74,
	0x65, 0x55, 0x70, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x12, 0x2c, 0x2e, 0x6b, 0x6f, 0x6e, 0x67,
	0x2e, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x76,
	0x31, 0x2e, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x55, 0x70, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x2d, 0x2e, 0x6b, 0x6f, 0x6e, 0x67, 0x2e, 0x61,
	0x64, 0x6d, 0x69, 0x6e, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x76, 0x31, 0x2e,
	0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x55, 0x70, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x1a, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x14, 0x2a, 0x12,
	0x2f, 0x76, 0x31, 0x2f, 0x75, 0x70, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x73, 0x2f, 0x7b, 0x69,
	0x64, 0x7d, 0x12, 0x81, 0x01, 0x0a, 0x0d, 0x4c, 0x69, 0x73, 0x74, 0x55, 0x70, 0x73, 0x74, 0x72,
	0x65, 0x61, 0x6d, 0x73, 0x12, 0x2b, 0x2e, 0x6b, 0x6f, 0x6e, 0x67, 0x2e, 0x61, 0x64, 0x6d, 0x69,
	0x6e, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x4c, 0x69, 0x73,
	0x74, 0x55, 0x70, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x1a, 0x2c, 0x2e, 0x6b, 0x6f, 0x6e, 0x67, 0x2e, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2e, 0x73,
	0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x4c, 0x69, 0x73, 0x74, 0x55, 0x70,
	0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22,
	0x15, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x0f, 0x12, 0x0d, 0x2f, 0x76, 0x31, 0x2f, 0x75, 0x70, 0x73,
	0x74, 0x72, 0x65, 0x61, 0x6d, 0x73, 0x42, 0x3c, 0x5a, 0x3a, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62,
	0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6b, 0x6f, 0x6e, 0x67, 0x2f, 0x6b, 0x6f, 0x6b, 0x6f, 0x2f, 0x69,
	0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x6b, 0x6f, 0x6e, 0x67,
	0x2f, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2f, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2f, 0x76,
	0x31, 0x3b, 0x76, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_kong_admin_service_v1_upstream_proto_rawDescOnce sync.Once
	file_kong_admin_service_v1_upstream_proto_rawDescData = file_kong_admin_service_v1_upstream_proto_rawDesc
)

func file_kong_admin_service_v1_upstream_proto_rawDescGZIP() []byte {
	file_kong_admin_service_v1_upstream_proto_rawDescOnce.Do(func() {
		file_kong_admin_service_v1_upstream_proto_rawDescData = protoimpl.X.CompressGZIP(file_kong_admin_service_v1_upstream_proto_rawDescData)
	})
	return file_kong_admin_service_v1_upstream_proto_rawDescData
}

var file_kong_admin_service_v1_upstream_proto_msgTypes = make([]protoimpl.MessageInfo, 10)
var file_kong_admin_service_v1_upstream_proto_goTypes = []interface{}{
	(*GetUpstreamRequest)(nil),     // 0: kong.admin.service.v1.GetUpstreamRequest
	(*GetUpstreamResponse)(nil),    // 1: kong.admin.service.v1.GetUpstreamResponse
	(*CreateUpstreamRequest)(nil),  // 2: kong.admin.service.v1.CreateUpstreamRequest
	(*CreateUpstreamResponse)(nil), // 3: kong.admin.service.v1.CreateUpstreamResponse
	(*UpsertUpstreamRequest)(nil),  // 4: kong.admin.service.v1.UpsertUpstreamRequest
	(*UpsertUpstreamResponse)(nil), // 5: kong.admin.service.v1.UpsertUpstreamResponse
	(*DeleteUpstreamRequest)(nil),  // 6: kong.admin.service.v1.DeleteUpstreamRequest
	(*DeleteUpstreamResponse)(nil), // 7: kong.admin.service.v1.DeleteUpstreamResponse
	(*ListUpstreamsRequest)(nil),   // 8: kong.admin.service.v1.ListUpstreamsRequest
	(*ListUpstreamsResponse)(nil),  // 9: kong.admin.service.v1.ListUpstreamsResponse
	(*v1.RequestCluster)(nil),      // 10: kong.admin.model.v1.RequestCluster
	(*v1.Upstream)(nil),            // 11: kong.admin.model.v1.Upstream
	(*v1.PaginationRequest)(nil),   // 12: kong.admin.model.v1.PaginationRequest
	(*v1.PaginationResponse)(nil),  // 13: kong.admin.model.v1.PaginationResponse
}
var file_kong_admin_service_v1_upstream_proto_depIdxs = []int32{
	10, // 0: kong.admin.service.v1.GetUpstreamRequest.cluster:type_name -> kong.admin.model.v1.RequestCluster
	11, // 1: kong.admin.service.v1.GetUpstreamResponse.item:type_name -> kong.admin.model.v1.Upstream
	11, // 2: kong.admin.service.v1.CreateUpstreamRequest.item:type_name -> kong.admin.model.v1.Upstream
	10, // 3: kong.admin.service.v1.CreateUpstreamRequest.cluster:type_name -> kong.admin.model.v1.RequestCluster
	11, // 4: kong.admin.service.v1.CreateUpstreamResponse.item:type_name -> kong.admin.model.v1.Upstream
	11, // 5: kong.admin.service.v1.UpsertUpstreamRequest.item:type_name -> kong.admin.model.v1.Upstream
	10, // 6: kong.admin.service.v1.UpsertUpstreamRequest.cluster:type_name -> kong.admin.model.v1.RequestCluster
	11, // 7: kong.admin.service.v1.UpsertUpstreamResponse.item:type_name -> kong.admin.model.v1.Upstream
	10, // 8: kong.admin.service.v1.DeleteUpstreamRequest.cluster:type_name -> kong.admin.model.v1.RequestCluster
	10, // 9: kong.admin.service.v1.ListUpstreamsRequest.cluster:type_name -> kong.admin.model.v1.RequestCluster
	12, // 10: kong.admin.service.v1.ListUpstreamsRequest.page:type_name -> kong.admin.model.v1.PaginationRequest
	11, // 11: kong.admin.service.v1.ListUpstreamsResponse.items:type_name -> kong.admin.model.v1.Upstream
	13, // 12: kong.admin.service.v1.ListUpstreamsResponse.page:type_name -> kong.admin.model.v1.PaginationResponse
	0,  // 13: kong.admin.service.v1.UpstreamService.GetUpstream:input_type -> kong.admin.service.v1.GetUpstreamRequest
	2,  // 14: kong.admin.service.v1.UpstreamService.CreateUpstream:input_type -> kong.admin.service.v1.CreateUpstreamRequest
	4,  // 15: kong.admin.service.v1.UpstreamService.UpsertUpstream:input_type -> kong.admin.service.v1.UpsertUpstreamRequest
	6,  // 16: kong.admin.service.v1.UpstreamService.DeleteUpstream:input_type -> kong.admin.service.v1.DeleteUpstreamRequest
	8,  // 17: kong.admin.service.v1.UpstreamService.ListUpstreams:input_type -> kong.admin.service.v1.ListUpstreamsRequest
	1,  // 18: kong.admin.service.v1.UpstreamService.GetUpstream:output_type -> kong.admin.service.v1.GetUpstreamResponse
	3,  // 19: kong.admin.service.v1.UpstreamService.CreateUpstream:output_type -> kong.admin.service.v1.CreateUpstreamResponse
	5,  // 20: kong.admin.service.v1.UpstreamService.UpsertUpstream:output_type -> kong.admin.service.v1.UpsertUpstreamResponse
	7,  // 21: kong.admin.service.v1.UpstreamService.DeleteUpstream:output_type -> kong.admin.service.v1.DeleteUpstreamResponse
	9,  // 22: kong.admin.service.v1.UpstreamService.ListUpstreams:output_type -> kong.admin.service.v1.ListUpstreamsResponse
	18, // [18:23] is the sub-list for method output_type
	13, // [13:18] is the sub-list for method input_type
	13, // [13:13] is the sub-list for extension type_name
	13, // [13:13] is the sub-list for extension extendee
	0,  // [0:13] is the sub-list for field type_name
}

func init() { file_kong_admin_service_v1_upstream_proto_init() }
func file_kong_admin_service_v1_upstream_proto_init() {
	if File_kong_admin_service_v1_upstream_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_kong_admin_service_v1_upstream_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetUpstreamRequest); i {
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
		file_kong_admin_service_v1_upstream_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetUpstreamResponse); i {
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
		file_kong_admin_service_v1_upstream_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CreateUpstreamRequest); i {
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
		file_kong_admin_service_v1_upstream_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CreateUpstreamResponse); i {
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
		file_kong_admin_service_v1_upstream_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UpsertUpstreamRequest); i {
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
		file_kong_admin_service_v1_upstream_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UpsertUpstreamResponse); i {
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
		file_kong_admin_service_v1_upstream_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DeleteUpstreamRequest); i {
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
		file_kong_admin_service_v1_upstream_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DeleteUpstreamResponse); i {
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
		file_kong_admin_service_v1_upstream_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ListUpstreamsRequest); i {
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
		file_kong_admin_service_v1_upstream_proto_msgTypes[9].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ListUpstreamsResponse); i {
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
			RawDescriptor: file_kong_admin_service_v1_upstream_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   10,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_kong_admin_service_v1_upstream_proto_goTypes,
		DependencyIndexes: file_kong_admin_service_v1_upstream_proto_depIdxs,
		MessageInfos:      file_kong_admin_service_v1_upstream_proto_msgTypes,
	}.Build()
	File_kong_admin_service_v1_upstream_proto = out.File
	file_kong_admin_service_v1_upstream_proto_rawDesc = nil
	file_kong_admin_service_v1_upstream_proto_goTypes = nil
	file_kong_admin_service_v1_upstream_proto_depIdxs = nil
}
