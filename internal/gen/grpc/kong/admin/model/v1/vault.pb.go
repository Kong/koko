// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        (unknown)
// source: kong/admin/model/v1/vault.proto

package v1

import (
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

type Vault struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id          string        `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	CreatedAt   int32         `protobuf:"varint,2,opt,name=created_at,json=createdAt,proto3" json:"created_at,omitempty"`
	UpdatedAt   int32         `protobuf:"varint,3,opt,name=updated_at,json=updatedAt,proto3" json:"updated_at,omitempty"`
	Prefix      string        `protobuf:"bytes,4,opt,name=prefix,proto3" json:"prefix,omitempty"`
	Name        string        `protobuf:"bytes,5,opt,name=name,proto3" json:"name,omitempty"`
	Description string        `protobuf:"bytes,6,opt,name=description,proto3" json:"description,omitempty"`
	Config      *Vault_Config `protobuf:"bytes,7,opt,name=config,proto3" json:"config,omitempty"`
	Tags        []string      `protobuf:"bytes,8,rep,name=tags,proto3" json:"tags,omitempty"`
}

func (x *Vault) Reset() {
	*x = Vault{}
	if protoimpl.UnsafeEnabled {
		mi := &file_kong_admin_model_v1_vault_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Vault) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Vault) ProtoMessage() {}

func (x *Vault) ProtoReflect() protoreflect.Message {
	mi := &file_kong_admin_model_v1_vault_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Vault.ProtoReflect.Descriptor instead.
func (*Vault) Descriptor() ([]byte, []int) {
	return file_kong_admin_model_v1_vault_proto_rawDescGZIP(), []int{0}
}

func (x *Vault) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *Vault) GetCreatedAt() int32 {
	if x != nil {
		return x.CreatedAt
	}
	return 0
}

func (x *Vault) GetUpdatedAt() int32 {
	if x != nil {
		return x.UpdatedAt
	}
	return 0
}

func (x *Vault) GetPrefix() string {
	if x != nil {
		return x.Prefix
	}
	return ""
}

func (x *Vault) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Vault) GetDescription() string {
	if x != nil {
		return x.Description
	}
	return ""
}

func (x *Vault) GetConfig() *Vault_Config {
	if x != nil {
		return x.Config
	}
	return nil
}

func (x *Vault) GetTags() []string {
	if x != nil {
		return x.Tags
	}
	return nil
}

type Vault_EnvConfig struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Prefix string `protobuf:"bytes,1,opt,name=prefix,proto3" json:"prefix,omitempty"`
}

func (x *Vault_EnvConfig) Reset() {
	*x = Vault_EnvConfig{}
	if protoimpl.UnsafeEnabled {
		mi := &file_kong_admin_model_v1_vault_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Vault_EnvConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Vault_EnvConfig) ProtoMessage() {}

func (x *Vault_EnvConfig) ProtoReflect() protoreflect.Message {
	mi := &file_kong_admin_model_v1_vault_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Vault_EnvConfig.ProtoReflect.Descriptor instead.
func (*Vault_EnvConfig) Descriptor() ([]byte, []int) {
	return file_kong_admin_model_v1_vault_proto_rawDescGZIP(), []int{0, 0}
}

func (x *Vault_EnvConfig) GetPrefix() string {
	if x != nil {
		return x.Prefix
	}
	return ""
}

type Vault_AwsConfig struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Region string `protobuf:"bytes,1,opt,name=region,proto3" json:"region,omitempty"`
}

func (x *Vault_AwsConfig) Reset() {
	*x = Vault_AwsConfig{}
	if protoimpl.UnsafeEnabled {
		mi := &file_kong_admin_model_v1_vault_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Vault_AwsConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Vault_AwsConfig) ProtoMessage() {}

func (x *Vault_AwsConfig) ProtoReflect() protoreflect.Message {
	mi := &file_kong_admin_model_v1_vault_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Vault_AwsConfig.ProtoReflect.Descriptor instead.
func (*Vault_AwsConfig) Descriptor() ([]byte, []int) {
	return file_kong_admin_model_v1_vault_proto_rawDescGZIP(), []int{0, 1}
}

func (x *Vault_AwsConfig) GetRegion() string {
	if x != nil {
		return x.Region
	}
	return ""
}

type Vault_GcpConfig struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ProjectId string `protobuf:"bytes,1,opt,name=project_id,json=projectId,proto3" json:"project_id,omitempty"`
}

func (x *Vault_GcpConfig) Reset() {
	*x = Vault_GcpConfig{}
	if protoimpl.UnsafeEnabled {
		mi := &file_kong_admin_model_v1_vault_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Vault_GcpConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Vault_GcpConfig) ProtoMessage() {}

func (x *Vault_GcpConfig) ProtoReflect() protoreflect.Message {
	mi := &file_kong_admin_model_v1_vault_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Vault_GcpConfig.ProtoReflect.Descriptor instead.
func (*Vault_GcpConfig) Descriptor() ([]byte, []int) {
	return file_kong_admin_model_v1_vault_proto_rawDescGZIP(), []int{0, 2}
}

func (x *Vault_GcpConfig) GetProjectId() string {
	if x != nil {
		return x.ProjectId
	}
	return ""
}

type Vault_HcvConfig struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Host     string `protobuf:"bytes,1,opt,name=host,proto3" json:"host,omitempty"`
	Port     int32  `protobuf:"varint,2,opt,name=port,proto3" json:"port,omitempty"`
	Protocol string `protobuf:"bytes,3,opt,name=protocol,proto3" json:"protocol,omitempty"`
	Mount    string `protobuf:"bytes,4,opt,name=mount,proto3" json:"mount,omitempty"`
	Kv       string `protobuf:"bytes,5,opt,name=kv,proto3" json:"kv,omitempty"`
	Token    string `protobuf:"bytes,6,opt,name=token,proto3" json:"token,omitempty"`
}

func (x *Vault_HcvConfig) Reset() {
	*x = Vault_HcvConfig{}
	if protoimpl.UnsafeEnabled {
		mi := &file_kong_admin_model_v1_vault_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Vault_HcvConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Vault_HcvConfig) ProtoMessage() {}

func (x *Vault_HcvConfig) ProtoReflect() protoreflect.Message {
	mi := &file_kong_admin_model_v1_vault_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Vault_HcvConfig.ProtoReflect.Descriptor instead.
func (*Vault_HcvConfig) Descriptor() ([]byte, []int) {
	return file_kong_admin_model_v1_vault_proto_rawDescGZIP(), []int{0, 3}
}

func (x *Vault_HcvConfig) GetHost() string {
	if x != nil {
		return x.Host
	}
	return ""
}

func (x *Vault_HcvConfig) GetPort() int32 {
	if x != nil {
		return x.Port
	}
	return 0
}

func (x *Vault_HcvConfig) GetProtocol() string {
	if x != nil {
		return x.Protocol
	}
	return ""
}

func (x *Vault_HcvConfig) GetMount() string {
	if x != nil {
		return x.Mount
	}
	return ""
}

func (x *Vault_HcvConfig) GetKv() string {
	if x != nil {
		return x.Kv
	}
	return ""
}

func (x *Vault_HcvConfig) GetToken() string {
	if x != nil {
		return x.Token
	}
	return ""
}

type Vault_Config struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to Config:
	//
	//	*Vault_Config_Env
	//	*Vault_Config_Aws
	//	*Vault_Config_Gcp
	//	*Vault_Config_Hcv
	Config isVault_Config_Config `protobuf_oneof:"config"`
}

func (x *Vault_Config) Reset() {
	*x = Vault_Config{}
	if protoimpl.UnsafeEnabled {
		mi := &file_kong_admin_model_v1_vault_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Vault_Config) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Vault_Config) ProtoMessage() {}

func (x *Vault_Config) ProtoReflect() protoreflect.Message {
	mi := &file_kong_admin_model_v1_vault_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Vault_Config.ProtoReflect.Descriptor instead.
func (*Vault_Config) Descriptor() ([]byte, []int) {
	return file_kong_admin_model_v1_vault_proto_rawDescGZIP(), []int{0, 4}
}

func (m *Vault_Config) GetConfig() isVault_Config_Config {
	if m != nil {
		return m.Config
	}
	return nil
}

func (x *Vault_Config) GetEnv() *Vault_EnvConfig {
	if x, ok := x.GetConfig().(*Vault_Config_Env); ok {
		return x.Env
	}
	return nil
}

func (x *Vault_Config) GetAws() *Vault_AwsConfig {
	if x, ok := x.GetConfig().(*Vault_Config_Aws); ok {
		return x.Aws
	}
	return nil
}

func (x *Vault_Config) GetGcp() *Vault_GcpConfig {
	if x, ok := x.GetConfig().(*Vault_Config_Gcp); ok {
		return x.Gcp
	}
	return nil
}

func (x *Vault_Config) GetHcv() *Vault_HcvConfig {
	if x, ok := x.GetConfig().(*Vault_Config_Hcv); ok {
		return x.Hcv
	}
	return nil
}

type isVault_Config_Config interface {
	isVault_Config_Config()
}

type Vault_Config_Env struct {
	Env *Vault_EnvConfig `protobuf:"bytes,1,opt,name=env,proto3,oneof"`
}

type Vault_Config_Aws struct {
	Aws *Vault_AwsConfig `protobuf:"bytes,2,opt,name=aws,proto3,oneof"`
}

type Vault_Config_Gcp struct {
	Gcp *Vault_GcpConfig `protobuf:"bytes,3,opt,name=gcp,proto3,oneof"`
}

type Vault_Config_Hcv struct {
	Hcv *Vault_HcvConfig `protobuf:"bytes,4,opt,name=hcv,proto3,oneof"`
}

func (*Vault_Config_Env) isVault_Config_Config() {}

func (*Vault_Config_Aws) isVault_Config_Config() {}

func (*Vault_Config_Gcp) isVault_Config_Config() {}

func (*Vault_Config_Hcv) isVault_Config_Config() {}

var File_kong_admin_model_v1_vault_proto protoreflect.FileDescriptor

var file_kong_admin_model_v1_vault_proto_rawDesc = []byte{
	0x0a, 0x1f, 0x6b, 0x6f, 0x6e, 0x67, 0x2f, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2f, 0x6d, 0x6f, 0x64,
	0x65, 0x6c, 0x2f, 0x76, 0x31, 0x2f, 0x76, 0x61, 0x75, 0x6c, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x12, 0x13, 0x6b, 0x6f, 0x6e, 0x67, 0x2e, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2e, 0x6d, 0x6f,
	0x64, 0x65, 0x6c, 0x2e, 0x76, 0x31, 0x1a, 0x1f, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x61,
	0x70, 0x69, 0x2f, 0x66, 0x69, 0x65, 0x6c, 0x64, 0x5f, 0x62, 0x65, 0x68, 0x61, 0x76, 0x69, 0x6f,
	0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xff, 0x05, 0x0a, 0x05, 0x56, 0x61, 0x75, 0x6c,
	0x74, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69,
	0x64, 0x12, 0x1d, 0x0a, 0x0a, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x5f, 0x61, 0x74, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x09, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x41, 0x74,
	0x12, 0x1d, 0x0a, 0x0a, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x64, 0x5f, 0x61, 0x74, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x05, 0x52, 0x09, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x64, 0x41, 0x74, 0x12,
	0x1c, 0x0a, 0x06, 0x70, 0x72, 0x65, 0x66, 0x69, 0x78, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x42,
	0x04, 0xe2, 0x41, 0x01, 0x02, 0x52, 0x06, 0x70, 0x72, 0x65, 0x66, 0x69, 0x78, 0x12, 0x18, 0x0a,
	0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x42, 0x04, 0xe2, 0x41, 0x01,
	0x02, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x20, 0x0a, 0x0b, 0x64, 0x65, 0x73, 0x63, 0x72,
	0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x06, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x64, 0x65,
	0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x39, 0x0a, 0x06, 0x63, 0x6f, 0x6e,
	0x66, 0x69, 0x67, 0x18, 0x07, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x21, 0x2e, 0x6b, 0x6f, 0x6e, 0x67,
	0x2e, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x2e, 0x76, 0x31, 0x2e,
	0x56, 0x61, 0x75, 0x6c, 0x74, 0x2e, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x52, 0x06, 0x63, 0x6f,
	0x6e, 0x66, 0x69, 0x67, 0x12, 0x12, 0x0a, 0x04, 0x74, 0x61, 0x67, 0x73, 0x18, 0x08, 0x20, 0x03,
	0x28, 0x09, 0x52, 0x04, 0x74, 0x61, 0x67, 0x73, 0x1a, 0x23, 0x0a, 0x09, 0x45, 0x6e, 0x76, 0x43,
	0x6f, 0x6e, 0x66, 0x69, 0x67, 0x12, 0x16, 0x0a, 0x06, 0x70, 0x72, 0x65, 0x66, 0x69, 0x78, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x70, 0x72, 0x65, 0x66, 0x69, 0x78, 0x1a, 0x23, 0x0a,
	0x09, 0x41, 0x77, 0x73, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x12, 0x16, 0x0a, 0x06, 0x72, 0x65,
	0x67, 0x69, 0x6f, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x72, 0x65, 0x67, 0x69,
	0x6f, 0x6e, 0x1a, 0x2a, 0x0a, 0x09, 0x47, 0x63, 0x70, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x12,
	0x1d, 0x0a, 0x0a, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x09, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x49, 0x64, 0x1a, 0x8b,
	0x01, 0x0a, 0x09, 0x48, 0x63, 0x76, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x12, 0x12, 0x0a, 0x04,
	0x68, 0x6f, 0x73, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x68, 0x6f, 0x73, 0x74,
	0x12, 0x12, 0x0a, 0x04, 0x70, 0x6f, 0x72, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x04,
	0x70, 0x6f, 0x72, 0x74, 0x12, 0x1a, 0x0a, 0x08, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c,
	0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c,
	0x12, 0x14, 0x0a, 0x05, 0x6d, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x05, 0x6d, 0x6f, 0x75, 0x6e, 0x74, 0x12, 0x0e, 0x0a, 0x02, 0x6b, 0x76, 0x18, 0x05, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x02, 0x6b, 0x76, 0x12, 0x14, 0x0a, 0x05, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x18,
	0x06, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x1a, 0xfa, 0x01, 0x0a,
	0x06, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x12, 0x38, 0x0a, 0x03, 0x65, 0x6e, 0x76, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x24, 0x2e, 0x6b, 0x6f, 0x6e, 0x67, 0x2e, 0x61, 0x64, 0x6d, 0x69,
	0x6e, 0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x56, 0x61, 0x75, 0x6c, 0x74,
	0x2e, 0x45, 0x6e, 0x76, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x48, 0x00, 0x52, 0x03, 0x65, 0x6e,
	0x76, 0x12, 0x38, 0x0a, 0x03, 0x61, 0x77, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x24,
	0x2e, 0x6b, 0x6f, 0x6e, 0x67, 0x2e, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2e, 0x6d, 0x6f, 0x64, 0x65,
	0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x56, 0x61, 0x75, 0x6c, 0x74, 0x2e, 0x41, 0x77, 0x73, 0x43, 0x6f,
	0x6e, 0x66, 0x69, 0x67, 0x48, 0x00, 0x52, 0x03, 0x61, 0x77, 0x73, 0x12, 0x38, 0x0a, 0x03, 0x67,
	0x63, 0x70, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x24, 0x2e, 0x6b, 0x6f, 0x6e, 0x67, 0x2e,
	0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x56,
	0x61, 0x75, 0x6c, 0x74, 0x2e, 0x47, 0x63, 0x70, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x48, 0x00,
	0x52, 0x03, 0x67, 0x63, 0x70, 0x12, 0x38, 0x0a, 0x03, 0x68, 0x63, 0x76, 0x18, 0x04, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x24, 0x2e, 0x6b, 0x6f, 0x6e, 0x67, 0x2e, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2e,
	0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x56, 0x61, 0x75, 0x6c, 0x74, 0x2e, 0x48,
	0x63, 0x76, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x48, 0x00, 0x52, 0x03, 0x68, 0x63, 0x76, 0x42,
	0x08, 0x0a, 0x06, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x42, 0x3f, 0x5a, 0x3d, 0x67, 0x69, 0x74,
	0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6b, 0x6f, 0x6e, 0x67, 0x2f, 0x6b, 0x6f, 0x6b,
	0x6f, 0x2f, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x67,
	0x72, 0x70, 0x63, 0x2f, 0x6b, 0x6f, 0x6e, 0x67, 0x2f, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2f, 0x6d,
	0x6f, 0x64, 0x65, 0x6c, 0x2f, 0x76, 0x31, 0x3b, 0x76, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x33,
}

var (
	file_kong_admin_model_v1_vault_proto_rawDescOnce sync.Once
	file_kong_admin_model_v1_vault_proto_rawDescData = file_kong_admin_model_v1_vault_proto_rawDesc
)

func file_kong_admin_model_v1_vault_proto_rawDescGZIP() []byte {
	file_kong_admin_model_v1_vault_proto_rawDescOnce.Do(func() {
		file_kong_admin_model_v1_vault_proto_rawDescData = protoimpl.X.CompressGZIP(file_kong_admin_model_v1_vault_proto_rawDescData)
	})
	return file_kong_admin_model_v1_vault_proto_rawDescData
}

var file_kong_admin_model_v1_vault_proto_msgTypes = make([]protoimpl.MessageInfo, 6)
var file_kong_admin_model_v1_vault_proto_goTypes = []interface{}{
	(*Vault)(nil),           // 0: kong.admin.model.v1.Vault
	(*Vault_EnvConfig)(nil), // 1: kong.admin.model.v1.Vault.EnvConfig
	(*Vault_AwsConfig)(nil), // 2: kong.admin.model.v1.Vault.AwsConfig
	(*Vault_GcpConfig)(nil), // 3: kong.admin.model.v1.Vault.GcpConfig
	(*Vault_HcvConfig)(nil), // 4: kong.admin.model.v1.Vault.HcvConfig
	(*Vault_Config)(nil),    // 5: kong.admin.model.v1.Vault.Config
}
var file_kong_admin_model_v1_vault_proto_depIdxs = []int32{
	5, // 0: kong.admin.model.v1.Vault.config:type_name -> kong.admin.model.v1.Vault.Config
	1, // 1: kong.admin.model.v1.Vault.Config.env:type_name -> kong.admin.model.v1.Vault.EnvConfig
	2, // 2: kong.admin.model.v1.Vault.Config.aws:type_name -> kong.admin.model.v1.Vault.AwsConfig
	3, // 3: kong.admin.model.v1.Vault.Config.gcp:type_name -> kong.admin.model.v1.Vault.GcpConfig
	4, // 4: kong.admin.model.v1.Vault.Config.hcv:type_name -> kong.admin.model.v1.Vault.HcvConfig
	5, // [5:5] is the sub-list for method output_type
	5, // [5:5] is the sub-list for method input_type
	5, // [5:5] is the sub-list for extension type_name
	5, // [5:5] is the sub-list for extension extendee
	0, // [0:5] is the sub-list for field type_name
}

func init() { file_kong_admin_model_v1_vault_proto_init() }
func file_kong_admin_model_v1_vault_proto_init() {
	if File_kong_admin_model_v1_vault_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_kong_admin_model_v1_vault_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Vault); i {
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
		file_kong_admin_model_v1_vault_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Vault_EnvConfig); i {
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
		file_kong_admin_model_v1_vault_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Vault_AwsConfig); i {
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
		file_kong_admin_model_v1_vault_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Vault_GcpConfig); i {
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
		file_kong_admin_model_v1_vault_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Vault_HcvConfig); i {
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
		file_kong_admin_model_v1_vault_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Vault_Config); i {
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
	file_kong_admin_model_v1_vault_proto_msgTypes[5].OneofWrappers = []interface{}{
		(*Vault_Config_Env)(nil),
		(*Vault_Config_Aws)(nil),
		(*Vault_Config_Gcp)(nil),
		(*Vault_Config_Hcv)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_kong_admin_model_v1_vault_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   6,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_kong_admin_model_v1_vault_proto_goTypes,
		DependencyIndexes: file_kong_admin_model_v1_vault_proto_depIdxs,
		MessageInfos:      file_kong_admin_model_v1_vault_proto_msgTypes,
	}.Build()
	File_kong_admin_model_v1_vault_proto = out.File
	file_kong_admin_model_v1_vault_proto_rawDesc = nil
	file_kong_admin_model_v1_vault_proto_goTypes = nil
	file_kong_admin_model_v1_vault_proto_depIdxs = nil
}
