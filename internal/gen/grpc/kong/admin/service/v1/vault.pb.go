// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1-devel
// 	protoc        (unknown)
// source: kong/admin/service/v1/vault.proto

package v1

import (
	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	_ "google.golang.org/genproto/googleapis/api/annotations"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	_ "google.golang.org/protobuf/types/descriptorpb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type GetVaultRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id      string             `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Cluster *v1.RequestCluster `protobuf:"bytes,2,opt,name=cluster,proto3" json:"cluster,omitempty"`
}

func (x *GetVaultRequest) Reset() {
	*x = GetVaultRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_kong_admin_service_v1_vault_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetVaultRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetVaultRequest) ProtoMessage() {}

func (x *GetVaultRequest) ProtoReflect() protoreflect.Message {
	mi := &file_kong_admin_service_v1_vault_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetVaultRequest.ProtoReflect.Descriptor instead.
func (*GetVaultRequest) Descriptor() ([]byte, []int) {
	return file_kong_admin_service_v1_vault_proto_rawDescGZIP(), []int{0}
}

func (x *GetVaultRequest) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *GetVaultRequest) GetCluster() *v1.RequestCluster {
	if x != nil {
		return x.Cluster
	}
	return nil
}

type GetVaultResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Item *v1.Vault `protobuf:"bytes,1,opt,name=item,proto3" json:"item,omitempty"`
}

func (x *GetVaultResponse) Reset() {
	*x = GetVaultResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_kong_admin_service_v1_vault_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetVaultResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetVaultResponse) ProtoMessage() {}

func (x *GetVaultResponse) ProtoReflect() protoreflect.Message {
	mi := &file_kong_admin_service_v1_vault_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetVaultResponse.ProtoReflect.Descriptor instead.
func (*GetVaultResponse) Descriptor() ([]byte, []int) {
	return file_kong_admin_service_v1_vault_proto_rawDescGZIP(), []int{1}
}

func (x *GetVaultResponse) GetItem() *v1.Vault {
	if x != nil {
		return x.Item
	}
	return nil
}

type CreateVaultRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Item    *v1.Vault          `protobuf:"bytes,1,opt,name=item,proto3" json:"item,omitempty"`
	Cluster *v1.RequestCluster `protobuf:"bytes,2,opt,name=cluster,proto3" json:"cluster,omitempty"`
}

func (x *CreateVaultRequest) Reset() {
	*x = CreateVaultRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_kong_admin_service_v1_vault_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CreateVaultRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CreateVaultRequest) ProtoMessage() {}

func (x *CreateVaultRequest) ProtoReflect() protoreflect.Message {
	mi := &file_kong_admin_service_v1_vault_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CreateVaultRequest.ProtoReflect.Descriptor instead.
func (*CreateVaultRequest) Descriptor() ([]byte, []int) {
	return file_kong_admin_service_v1_vault_proto_rawDescGZIP(), []int{2}
}

func (x *CreateVaultRequest) GetItem() *v1.Vault {
	if x != nil {
		return x.Item
	}
	return nil
}

func (x *CreateVaultRequest) GetCluster() *v1.RequestCluster {
	if x != nil {
		return x.Cluster
	}
	return nil
}

type CreateVaultResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Item *v1.Vault `protobuf:"bytes,1,opt,name=item,proto3" json:"item,omitempty"`
}

func (x *CreateVaultResponse) Reset() {
	*x = CreateVaultResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_kong_admin_service_v1_vault_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CreateVaultResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CreateVaultResponse) ProtoMessage() {}

func (x *CreateVaultResponse) ProtoReflect() protoreflect.Message {
	mi := &file_kong_admin_service_v1_vault_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CreateVaultResponse.ProtoReflect.Descriptor instead.
func (*CreateVaultResponse) Descriptor() ([]byte, []int) {
	return file_kong_admin_service_v1_vault_proto_rawDescGZIP(), []int{3}
}

func (x *CreateVaultResponse) GetItem() *v1.Vault {
	if x != nil {
		return x.Item
	}
	return nil
}

type UpsertVaultRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Item    *v1.Vault          `protobuf:"bytes,1,opt,name=item,proto3" json:"item,omitempty"`
	Cluster *v1.RequestCluster `protobuf:"bytes,2,opt,name=cluster,proto3" json:"cluster,omitempty"`
}

func (x *UpsertVaultRequest) Reset() {
	*x = UpsertVaultRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_kong_admin_service_v1_vault_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UpsertVaultRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpsertVaultRequest) ProtoMessage() {}

func (x *UpsertVaultRequest) ProtoReflect() protoreflect.Message {
	mi := &file_kong_admin_service_v1_vault_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UpsertVaultRequest.ProtoReflect.Descriptor instead.
func (*UpsertVaultRequest) Descriptor() ([]byte, []int) {
	return file_kong_admin_service_v1_vault_proto_rawDescGZIP(), []int{4}
}

func (x *UpsertVaultRequest) GetItem() *v1.Vault {
	if x != nil {
		return x.Item
	}
	return nil
}

func (x *UpsertVaultRequest) GetCluster() *v1.RequestCluster {
	if x != nil {
		return x.Cluster
	}
	return nil
}

type UpsertVaultResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Item *v1.Vault `protobuf:"bytes,1,opt,name=item,proto3" json:"item,omitempty"`
}

func (x *UpsertVaultResponse) Reset() {
	*x = UpsertVaultResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_kong_admin_service_v1_vault_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UpsertVaultResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpsertVaultResponse) ProtoMessage() {}

func (x *UpsertVaultResponse) ProtoReflect() protoreflect.Message {
	mi := &file_kong_admin_service_v1_vault_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UpsertVaultResponse.ProtoReflect.Descriptor instead.
func (*UpsertVaultResponse) Descriptor() ([]byte, []int) {
	return file_kong_admin_service_v1_vault_proto_rawDescGZIP(), []int{5}
}

func (x *UpsertVaultResponse) GetItem() *v1.Vault {
	if x != nil {
		return x.Item
	}
	return nil
}

type DeleteVaultRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id      string             `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Cluster *v1.RequestCluster `protobuf:"bytes,2,opt,name=cluster,proto3" json:"cluster,omitempty"`
}

func (x *DeleteVaultRequest) Reset() {
	*x = DeleteVaultRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_kong_admin_service_v1_vault_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DeleteVaultRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DeleteVaultRequest) ProtoMessage() {}

func (x *DeleteVaultRequest) ProtoReflect() protoreflect.Message {
	mi := &file_kong_admin_service_v1_vault_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DeleteVaultRequest.ProtoReflect.Descriptor instead.
func (*DeleteVaultRequest) Descriptor() ([]byte, []int) {
	return file_kong_admin_service_v1_vault_proto_rawDescGZIP(), []int{6}
}

func (x *DeleteVaultRequest) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *DeleteVaultRequest) GetCluster() *v1.RequestCluster {
	if x != nil {
		return x.Cluster
	}
	return nil
}

type DeleteVaultResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *DeleteVaultResponse) Reset() {
	*x = DeleteVaultResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_kong_admin_service_v1_vault_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DeleteVaultResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DeleteVaultResponse) ProtoMessage() {}

func (x *DeleteVaultResponse) ProtoReflect() protoreflect.Message {
	mi := &file_kong_admin_service_v1_vault_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DeleteVaultResponse.ProtoReflect.Descriptor instead.
func (*DeleteVaultResponse) Descriptor() ([]byte, []int) {
	return file_kong_admin_service_v1_vault_proto_rawDescGZIP(), []int{7}
}

type ListVaultsRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Cluster *v1.RequestCluster    `protobuf:"bytes,1,opt,name=cluster,proto3" json:"cluster,omitempty"`
	Page    *v1.PaginationRequest `protobuf:"bytes,2,opt,name=page,proto3" json:"page,omitempty"`
}

func (x *ListVaultsRequest) Reset() {
	*x = ListVaultsRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_kong_admin_service_v1_vault_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ListVaultsRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListVaultsRequest) ProtoMessage() {}

func (x *ListVaultsRequest) ProtoReflect() protoreflect.Message {
	mi := &file_kong_admin_service_v1_vault_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListVaultsRequest.ProtoReflect.Descriptor instead.
func (*ListVaultsRequest) Descriptor() ([]byte, []int) {
	return file_kong_admin_service_v1_vault_proto_rawDescGZIP(), []int{8}
}

func (x *ListVaultsRequest) GetCluster() *v1.RequestCluster {
	if x != nil {
		return x.Cluster
	}
	return nil
}

func (x *ListVaultsRequest) GetPage() *v1.PaginationRequest {
	if x != nil {
		return x.Page
	}
	return nil
}

type ListVaultsResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Items []*v1.Vault            `protobuf:"bytes,1,rep,name=items,proto3" json:"items,omitempty"`
	Page  *v1.PaginationResponse `protobuf:"bytes,2,opt,name=page,proto3" json:"page,omitempty"`
}

func (x *ListVaultsResponse) Reset() {
	*x = ListVaultsResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_kong_admin_service_v1_vault_proto_msgTypes[9]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ListVaultsResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListVaultsResponse) ProtoMessage() {}

func (x *ListVaultsResponse) ProtoReflect() protoreflect.Message {
	mi := &file_kong_admin_service_v1_vault_proto_msgTypes[9]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListVaultsResponse.ProtoReflect.Descriptor instead.
func (*ListVaultsResponse) Descriptor() ([]byte, []int) {
	return file_kong_admin_service_v1_vault_proto_rawDescGZIP(), []int{9}
}

func (x *ListVaultsResponse) GetItems() []*v1.Vault {
	if x != nil {
		return x.Items
	}
	return nil
}

func (x *ListVaultsResponse) GetPage() *v1.PaginationResponse {
	if x != nil {
		return x.Page
	}
	return nil
}

var File_kong_admin_service_v1_vault_proto protoreflect.FileDescriptor

var file_kong_admin_service_v1_vault_proto_rawDesc = []byte{
	0x0a, 0x21, 0x6b, 0x6f, 0x6e, 0x67, 0x2f, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2f, 0x73, 0x65, 0x72,
	0x76, 0x69, 0x63, 0x65, 0x2f, 0x76, 0x31, 0x2f, 0x76, 0x61, 0x75, 0x6c, 0x74, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x12, 0x15, 0x6b, 0x6f, 0x6e, 0x67, 0x2e, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2e,
	0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x76, 0x31, 0x1a, 0x1c, 0x67, 0x6f, 0x6f, 0x67,
	0x6c, 0x65, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f,
	0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x20, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65,
	0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69,
	0x70, 0x74, 0x6f, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x21, 0x6b, 0x6f, 0x6e, 0x67,
	0x2f, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2f, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x2f, 0x76, 0x31, 0x2f,
	0x63, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x24, 0x6b,
	0x6f, 0x6e, 0x67, 0x2f, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2f, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x2f,
	0x76, 0x31, 0x2f, 0x70, 0x61, 0x67, 0x69, 0x6e, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x1a, 0x1f, 0x6b, 0x6f, 0x6e, 0x67, 0x2f, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2f,
	0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x2f, 0x76, 0x31, 0x2f, 0x76, 0x61, 0x75, 0x6c, 0x74, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x22, 0x60, 0x0a, 0x0f, 0x47, 0x65, 0x74, 0x56, 0x61, 0x75, 0x6c, 0x74,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12, 0x3d, 0x0a, 0x07, 0x63, 0x6c, 0x75, 0x73, 0x74,
	0x65, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x23, 0x2e, 0x6b, 0x6f, 0x6e, 0x67, 0x2e,
	0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x43, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x52, 0x07, 0x63,
	0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x22, 0x42, 0x0a, 0x10, 0x47, 0x65, 0x74, 0x56, 0x61, 0x75,
	0x6c, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x2e, 0x0a, 0x04, 0x69, 0x74,
	0x65, 0x6d, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x6b, 0x6f, 0x6e, 0x67, 0x2e,
	0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x56,
	0x61, 0x75, 0x6c, 0x74, 0x52, 0x04, 0x69, 0x74, 0x65, 0x6d, 0x22, 0x83, 0x01, 0x0a, 0x12, 0x43,
	0x72, 0x65, 0x61, 0x74, 0x65, 0x56, 0x61, 0x75, 0x6c, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x12, 0x2e, 0x0a, 0x04, 0x69, 0x74, 0x65, 0x6d, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x1a, 0x2e, 0x6b, 0x6f, 0x6e, 0x67, 0x2e, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2e, 0x6d, 0x6f, 0x64,
	0x65, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x56, 0x61, 0x75, 0x6c, 0x74, 0x52, 0x04, 0x69, 0x74, 0x65,
	0x6d, 0x12, 0x3d, 0x0a, 0x07, 0x63, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x23, 0x2e, 0x6b, 0x6f, 0x6e, 0x67, 0x2e, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2e,
	0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x43, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x52, 0x07, 0x63, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72,
	0x22, 0x45, 0x0a, 0x13, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x56, 0x61, 0x75, 0x6c, 0x74, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x2e, 0x0a, 0x04, 0x69, 0x74, 0x65, 0x6d, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x6b, 0x6f, 0x6e, 0x67, 0x2e, 0x61, 0x64, 0x6d,
	0x69, 0x6e, 0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x56, 0x61, 0x75, 0x6c,
	0x74, 0x52, 0x04, 0x69, 0x74, 0x65, 0x6d, 0x22, 0x83, 0x01, 0x0a, 0x12, 0x55, 0x70, 0x73, 0x65,
	0x72, 0x74, 0x56, 0x61, 0x75, 0x6c, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x2e,
	0x0a, 0x04, 0x69, 0x74, 0x65, 0x6d, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x6b,
	0x6f, 0x6e, 0x67, 0x2e, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x2e,
	0x76, 0x31, 0x2e, 0x56, 0x61, 0x75, 0x6c, 0x74, 0x52, 0x04, 0x69, 0x74, 0x65, 0x6d, 0x12, 0x3d,
	0x0a, 0x07, 0x63, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x23, 0x2e, 0x6b, 0x6f, 0x6e, 0x67, 0x2e, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2e, 0x6d, 0x6f, 0x64,
	0x65, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x43, 0x6c, 0x75,
	0x73, 0x74, 0x65, 0x72, 0x52, 0x07, 0x63, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x22, 0x45, 0x0a,
	0x13, 0x55, 0x70, 0x73, 0x65, 0x72, 0x74, 0x56, 0x61, 0x75, 0x6c, 0x74, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x12, 0x2e, 0x0a, 0x04, 0x69, 0x74, 0x65, 0x6d, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x6b, 0x6f, 0x6e, 0x67, 0x2e, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2e,
	0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x56, 0x61, 0x75, 0x6c, 0x74, 0x52, 0x04,
	0x69, 0x74, 0x65, 0x6d, 0x22, 0x63, 0x0a, 0x12, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x56, 0x61,
	0x75, 0x6c, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12, 0x3d, 0x0a, 0x07, 0x63, 0x6c,
	0x75, 0x73, 0x74, 0x65, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x23, 0x2e, 0x6b, 0x6f,
	0x6e, 0x67, 0x2e, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x2e, 0x76,
	0x31, 0x2e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x43, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72,
	0x52, 0x07, 0x63, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x22, 0x15, 0x0a, 0x13, 0x44, 0x65, 0x6c,
	0x65, 0x74, 0x65, 0x56, 0x61, 0x75, 0x6c, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x22, 0x8e, 0x01, 0x0a, 0x11, 0x4c, 0x69, 0x73, 0x74, 0x56, 0x61, 0x75, 0x6c, 0x74, 0x73, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x3d, 0x0a, 0x07, 0x63, 0x6c, 0x75, 0x73, 0x74, 0x65,
	0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x23, 0x2e, 0x6b, 0x6f, 0x6e, 0x67, 0x2e, 0x61,
	0x64, 0x6d, 0x69, 0x6e, 0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x43, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x52, 0x07, 0x63, 0x6c,
	0x75, 0x73, 0x74, 0x65, 0x72, 0x12, 0x3a, 0x0a, 0x04, 0x70, 0x61, 0x67, 0x65, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x26, 0x2e, 0x6b, 0x6f, 0x6e, 0x67, 0x2e, 0x61, 0x64, 0x6d, 0x69, 0x6e,
	0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x50, 0x61, 0x67, 0x69, 0x6e, 0x61,
	0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x52, 0x04, 0x70, 0x61, 0x67,
	0x65, 0x22, 0x83, 0x01, 0x0a, 0x12, 0x4c, 0x69, 0x73, 0x74, 0x56, 0x61, 0x75, 0x6c, 0x74, 0x73,
	0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x30, 0x0a, 0x05, 0x69, 0x74, 0x65, 0x6d,
	0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x6b, 0x6f, 0x6e, 0x67, 0x2e, 0x61,
	0x64, 0x6d, 0x69, 0x6e, 0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x56, 0x61,
	0x75, 0x6c, 0x74, 0x52, 0x05, 0x69, 0x74, 0x65, 0x6d, 0x73, 0x12, 0x3b, 0x0a, 0x04, 0x70, 0x61,
	0x67, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x27, 0x2e, 0x6b, 0x6f, 0x6e, 0x67, 0x2e,
	0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x50,
	0x61, 0x67, 0x69, 0x6e, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x52, 0x04, 0x70, 0x61, 0x67, 0x65, 0x32, 0x85, 0x05, 0x0a, 0x0c, 0x56, 0x61, 0x75, 0x6c,
	0x74, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x74, 0x0a, 0x08, 0x47, 0x65, 0x74, 0x56,
	0x61, 0x75, 0x6c, 0x74, 0x12, 0x26, 0x2e, 0x6b, 0x6f, 0x6e, 0x67, 0x2e, 0x61, 0x64, 0x6d, 0x69,
	0x6e, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x47, 0x65, 0x74,
	0x56, 0x61, 0x75, 0x6c, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x27, 0x2e, 0x6b,
	0x6f, 0x6e, 0x67, 0x2e, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63,
	0x65, 0x2e, 0x76, 0x31, 0x2e, 0x47, 0x65, 0x74, 0x56, 0x61, 0x75, 0x6c, 0x74, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x17, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x11, 0x12, 0x0f, 0x2f,
	0x76, 0x31, 0x2f, 0x76, 0x61, 0x75, 0x6c, 0x74, 0x73, 0x2f, 0x7b, 0x69, 0x64, 0x7d, 0x12, 0x7e,
	0x0a, 0x0b, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x56, 0x61, 0x75, 0x6c, 0x74, 0x12, 0x29, 0x2e,
	0x6b, 0x6f, 0x6e, 0x67, 0x2e, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x69,
	0x63, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x56, 0x61, 0x75, 0x6c,
	0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x2a, 0x2e, 0x6b, 0x6f, 0x6e, 0x67, 0x2e,
	0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x76, 0x31,
	0x2e, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x56, 0x61, 0x75, 0x6c, 0x74, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x22, 0x18, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x12, 0x3a, 0x04, 0x69, 0x74,
	0x65, 0x6d, 0x22, 0x0a, 0x2f, 0x76, 0x31, 0x2f, 0x76, 0x61, 0x75, 0x6c, 0x74, 0x73, 0x12, 0x88,
	0x01, 0x0a, 0x0b, 0x55, 0x70, 0x73, 0x65, 0x72, 0x74, 0x56, 0x61, 0x75, 0x6c, 0x74, 0x12, 0x29,
	0x2e, 0x6b, 0x6f, 0x6e, 0x67, 0x2e, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2e, 0x73, 0x65, 0x72, 0x76,
	0x69, 0x63, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x55, 0x70, 0x73, 0x65, 0x72, 0x74, 0x56, 0x61, 0x75,
	0x6c, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x2a, 0x2e, 0x6b, 0x6f, 0x6e, 0x67,
	0x2e, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x76,
	0x31, 0x2e, 0x55, 0x70, 0x73, 0x65, 0x72, 0x74, 0x56, 0x61, 0x75, 0x6c, 0x74, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x22, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x1c, 0x3a, 0x04, 0x69,
	0x74, 0x65, 0x6d, 0x1a, 0x14, 0x2f, 0x76, 0x31, 0x2f, 0x76, 0x61, 0x75, 0x6c, 0x74, 0x73, 0x2f,
	0x7b, 0x69, 0x74, 0x65, 0x6d, 0x2e, 0x69, 0x64, 0x7d, 0x12, 0x7d, 0x0a, 0x0b, 0x44, 0x65, 0x6c,
	0x65, 0x74, 0x65, 0x56, 0x61, 0x75, 0x6c, 0x74, 0x12, 0x29, 0x2e, 0x6b, 0x6f, 0x6e, 0x67, 0x2e,
	0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x76, 0x31,
	0x2e, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x56, 0x61, 0x75, 0x6c, 0x74, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x1a, 0x2a, 0x2e, 0x6b, 0x6f, 0x6e, 0x67, 0x2e, 0x61, 0x64, 0x6d, 0x69, 0x6e,
	0x2e, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x44, 0x65, 0x6c, 0x65,
	0x74, 0x65, 0x56, 0x61, 0x75, 0x6c, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22,
	0x17, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x11, 0x2a, 0x0f, 0x2f, 0x76, 0x31, 0x2f, 0x76, 0x61, 0x75,
	0x6c, 0x74, 0x73, 0x2f, 0x7b, 0x69, 0x64, 0x7d, 0x12, 0x75, 0x0a, 0x0a, 0x4c, 0x69, 0x73, 0x74,
	0x56, 0x61, 0x75, 0x6c, 0x74, 0x73, 0x12, 0x28, 0x2e, 0x6b, 0x6f, 0x6e, 0x67, 0x2e, 0x61, 0x64,
	0x6d, 0x69, 0x6e, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x4c,
	0x69, 0x73, 0x74, 0x56, 0x61, 0x75, 0x6c, 0x74, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x1a, 0x29, 0x2e, 0x6b, 0x6f, 0x6e, 0x67, 0x2e, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2e, 0x73, 0x65,
	0x72, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x4c, 0x69, 0x73, 0x74, 0x56, 0x61, 0x75,
	0x6c, 0x74, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x12, 0x82, 0xd3, 0xe4,
	0x93, 0x02, 0x0c, 0x12, 0x0a, 0x2f, 0x76, 0x31, 0x2f, 0x76, 0x61, 0x75, 0x6c, 0x74, 0x73, 0x42,
	0x3c, 0x5a, 0x3a, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6b, 0x6f,
	0x6e, 0x67, 0x2f, 0x6b, 0x6f, 0x6b, 0x6f, 0x2f, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c,
	0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x6b, 0x6f, 0x6e, 0x67, 0x2f, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2f,
	0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2f, 0x76, 0x31, 0x3b, 0x76, 0x31, 0x62, 0x06, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_kong_admin_service_v1_vault_proto_rawDescOnce sync.Once
	file_kong_admin_service_v1_vault_proto_rawDescData = file_kong_admin_service_v1_vault_proto_rawDesc
)

func file_kong_admin_service_v1_vault_proto_rawDescGZIP() []byte {
	file_kong_admin_service_v1_vault_proto_rawDescOnce.Do(func() {
		file_kong_admin_service_v1_vault_proto_rawDescData = protoimpl.X.CompressGZIP(file_kong_admin_service_v1_vault_proto_rawDescData)
	})
	return file_kong_admin_service_v1_vault_proto_rawDescData
}

var file_kong_admin_service_v1_vault_proto_msgTypes = make([]protoimpl.MessageInfo, 10)
var file_kong_admin_service_v1_vault_proto_goTypes = []interface{}{
	(*GetVaultRequest)(nil),       // 0: kong.admin.service.v1.GetVaultRequest
	(*GetVaultResponse)(nil),      // 1: kong.admin.service.v1.GetVaultResponse
	(*CreateVaultRequest)(nil),    // 2: kong.admin.service.v1.CreateVaultRequest
	(*CreateVaultResponse)(nil),   // 3: kong.admin.service.v1.CreateVaultResponse
	(*UpsertVaultRequest)(nil),    // 4: kong.admin.service.v1.UpsertVaultRequest
	(*UpsertVaultResponse)(nil),   // 5: kong.admin.service.v1.UpsertVaultResponse
	(*DeleteVaultRequest)(nil),    // 6: kong.admin.service.v1.DeleteVaultRequest
	(*DeleteVaultResponse)(nil),   // 7: kong.admin.service.v1.DeleteVaultResponse
	(*ListVaultsRequest)(nil),     // 8: kong.admin.service.v1.ListVaultsRequest
	(*ListVaultsResponse)(nil),    // 9: kong.admin.service.v1.ListVaultsResponse
	(*v1.RequestCluster)(nil),     // 10: kong.admin.model.v1.RequestCluster
	(*v1.Vault)(nil),              // 11: kong.admin.model.v1.Vault
	(*v1.PaginationRequest)(nil),  // 12: kong.admin.model.v1.PaginationRequest
	(*v1.PaginationResponse)(nil), // 13: kong.admin.model.v1.PaginationResponse
}
var file_kong_admin_service_v1_vault_proto_depIdxs = []int32{
	10, // 0: kong.admin.service.v1.GetVaultRequest.cluster:type_name -> kong.admin.model.v1.RequestCluster
	11, // 1: kong.admin.service.v1.GetVaultResponse.item:type_name -> kong.admin.model.v1.Vault
	11, // 2: kong.admin.service.v1.CreateVaultRequest.item:type_name -> kong.admin.model.v1.Vault
	10, // 3: kong.admin.service.v1.CreateVaultRequest.cluster:type_name -> kong.admin.model.v1.RequestCluster
	11, // 4: kong.admin.service.v1.CreateVaultResponse.item:type_name -> kong.admin.model.v1.Vault
	11, // 5: kong.admin.service.v1.UpsertVaultRequest.item:type_name -> kong.admin.model.v1.Vault
	10, // 6: kong.admin.service.v1.UpsertVaultRequest.cluster:type_name -> kong.admin.model.v1.RequestCluster
	11, // 7: kong.admin.service.v1.UpsertVaultResponse.item:type_name -> kong.admin.model.v1.Vault
	10, // 8: kong.admin.service.v1.DeleteVaultRequest.cluster:type_name -> kong.admin.model.v1.RequestCluster
	10, // 9: kong.admin.service.v1.ListVaultsRequest.cluster:type_name -> kong.admin.model.v1.RequestCluster
	12, // 10: kong.admin.service.v1.ListVaultsRequest.page:type_name -> kong.admin.model.v1.PaginationRequest
	11, // 11: kong.admin.service.v1.ListVaultsResponse.items:type_name -> kong.admin.model.v1.Vault
	13, // 12: kong.admin.service.v1.ListVaultsResponse.page:type_name -> kong.admin.model.v1.PaginationResponse
	0,  // 13: kong.admin.service.v1.VaultService.GetVault:input_type -> kong.admin.service.v1.GetVaultRequest
	2,  // 14: kong.admin.service.v1.VaultService.CreateVault:input_type -> kong.admin.service.v1.CreateVaultRequest
	4,  // 15: kong.admin.service.v1.VaultService.UpsertVault:input_type -> kong.admin.service.v1.UpsertVaultRequest
	6,  // 16: kong.admin.service.v1.VaultService.DeleteVault:input_type -> kong.admin.service.v1.DeleteVaultRequest
	8,  // 17: kong.admin.service.v1.VaultService.ListVaults:input_type -> kong.admin.service.v1.ListVaultsRequest
	1,  // 18: kong.admin.service.v1.VaultService.GetVault:output_type -> kong.admin.service.v1.GetVaultResponse
	3,  // 19: kong.admin.service.v1.VaultService.CreateVault:output_type -> kong.admin.service.v1.CreateVaultResponse
	5,  // 20: kong.admin.service.v1.VaultService.UpsertVault:output_type -> kong.admin.service.v1.UpsertVaultResponse
	7,  // 21: kong.admin.service.v1.VaultService.DeleteVault:output_type -> kong.admin.service.v1.DeleteVaultResponse
	9,  // 22: kong.admin.service.v1.VaultService.ListVaults:output_type -> kong.admin.service.v1.ListVaultsResponse
	18, // [18:23] is the sub-list for method output_type
	13, // [13:18] is the sub-list for method input_type
	13, // [13:13] is the sub-list for extension type_name
	13, // [13:13] is the sub-list for extension extendee
	0,  // [0:13] is the sub-list for field type_name
}

func init() { file_kong_admin_service_v1_vault_proto_init() }
func file_kong_admin_service_v1_vault_proto_init() {
	if File_kong_admin_service_v1_vault_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_kong_admin_service_v1_vault_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetVaultRequest); i {
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
		file_kong_admin_service_v1_vault_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetVaultResponse); i {
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
		file_kong_admin_service_v1_vault_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CreateVaultRequest); i {
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
		file_kong_admin_service_v1_vault_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CreateVaultResponse); i {
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
		file_kong_admin_service_v1_vault_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UpsertVaultRequest); i {
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
		file_kong_admin_service_v1_vault_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UpsertVaultResponse); i {
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
		file_kong_admin_service_v1_vault_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DeleteVaultRequest); i {
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
		file_kong_admin_service_v1_vault_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DeleteVaultResponse); i {
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
		file_kong_admin_service_v1_vault_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ListVaultsRequest); i {
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
		file_kong_admin_service_v1_vault_proto_msgTypes[9].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ListVaultsResponse); i {
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
			RawDescriptor: file_kong_admin_service_v1_vault_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   10,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_kong_admin_service_v1_vault_proto_goTypes,
		DependencyIndexes: file_kong_admin_service_v1_vault_proto_depIdxs,
		MessageInfos:      file_kong_admin_service_v1_vault_proto_msgTypes,
	}.Build()
	File_kong_admin_service_v1_vault_proto = out.File
	file_kong_admin_service_v1_vault_proto_rawDesc = nil
	file_kong_admin_service_v1_vault_proto_goTypes = nil
	file_kong_admin_service_v1_vault_proto_depIdxs = nil
}
