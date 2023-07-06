// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.30.0
// 	protoc        v3.21.12
// source: providers.proto

package proto

import (
	llx "go.mondoo.com/cnquery/llx"
	v1 "go.mondoo.com/cnquery/motor/inventory/v1"
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

type ParseCLIReq struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Connector string                    `protobuf:"bytes,1,opt,name=connector,proto3" json:"connector,omitempty"`
	Args      []string                  `protobuf:"bytes,2,rep,name=args,proto3" json:"args,omitempty"`
	Flags     map[string]*llx.Primitive `protobuf:"bytes,3,rep,name=flags,proto3" json:"flags,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (x *ParseCLIReq) Reset() {
	*x = ParseCLIReq{}
	if protoimpl.UnsafeEnabled {
		mi := &file_providers_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ParseCLIReq) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ParseCLIReq) ProtoMessage() {}

func (x *ParseCLIReq) ProtoReflect() protoreflect.Message {
	mi := &file_providers_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ParseCLIReq.ProtoReflect.Descriptor instead.
func (*ParseCLIReq) Descriptor() ([]byte, []int) {
	return file_providers_proto_rawDescGZIP(), []int{0}
}

func (x *ParseCLIReq) GetConnector() string {
	if x != nil {
		return x.Connector
	}
	return ""
}

func (x *ParseCLIReq) GetArgs() []string {
	if x != nil {
		return x.Args
	}
	return nil
}

func (x *ParseCLIReq) GetFlags() map[string]*llx.Primitive {
	if x != nil {
		return x.Flags
	}
	return nil
}

type ParseCLIRes struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// full inventory of everything that was requested
	Inventory *v1.Inventory `protobuf:"bytes,1,opt,name=inventory,proto3" json:"inventory,omitempty"`
}

func (x *ParseCLIRes) Reset() {
	*x = ParseCLIRes{}
	if protoimpl.UnsafeEnabled {
		mi := &file_providers_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ParseCLIRes) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ParseCLIRes) ProtoMessage() {}

func (x *ParseCLIRes) ProtoReflect() protoreflect.Message {
	mi := &file_providers_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ParseCLIRes.ProtoReflect.Descriptor instead.
func (*ParseCLIRes) Descriptor() ([]byte, []int) {
	return file_providers_proto_rawDescGZIP(), []int{1}
}

func (x *ParseCLIRes) GetInventory() *v1.Inventory {
	if x != nil {
		return x.Inventory
	}
	return nil
}

type ConnectReq struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Features []byte `protobuf:"bytes,2,opt,name=features,proto3" json:"features,omitempty"`
	// Asset is one target that is being connected to
	Asset *v1.Inventory `protobuf:"bytes,3,opt,name=asset,proto3" json:"asset,omitempty"`
}

func (x *ConnectReq) Reset() {
	*x = ConnectReq{}
	if protoimpl.UnsafeEnabled {
		mi := &file_providers_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ConnectReq) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ConnectReq) ProtoMessage() {}

func (x *ConnectReq) ProtoReflect() protoreflect.Message {
	mi := &file_providers_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ConnectReq.ProtoReflect.Descriptor instead.
func (*ConnectReq) Descriptor() ([]byte, []int) {
	return file_providers_proto_rawDescGZIP(), []int{2}
}

func (x *ConnectReq) GetFeatures() []byte {
	if x != nil {
		return x.Features
	}
	return nil
}

func (x *ConnectReq) GetAsset() *v1.Inventory {
	if x != nil {
		return x.Asset
	}
	return nil
}

type Connection struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id uint32 `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
}

func (x *Connection) Reset() {
	*x = Connection{}
	if protoimpl.UnsafeEnabled {
		mi := &file_providers_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Connection) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Connection) ProtoMessage() {}

func (x *Connection) ProtoReflect() protoreflect.Message {
	mi := &file_providers_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Connection.ProtoReflect.Descriptor instead.
func (*Connection) Descriptor() ([]byte, []int) {
	return file_providers_proto_rawDescGZIP(), []int{3}
}

func (x *Connection) GetId() uint32 {
	if x != nil {
		return x.Id
	}
	return 0
}

type DataReq struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Connection     uint32                    `protobuf:"varint,1,opt,name=connection,proto3" json:"connection,omitempty"`
	CallbackServer uint32                    `protobuf:"varint,2,opt,name=callback_server,json=callbackServer,proto3" json:"callback_server,omitempty"`
	Resource       string                    `protobuf:"bytes,3,opt,name=resource,proto3" json:"resource,omitempty"`
	ResourceId     string                    `protobuf:"bytes,4,opt,name=resource_id,json=resourceId,proto3" json:"resource_id,omitempty"`
	Field          string                    `protobuf:"bytes,5,opt,name=field,proto3" json:"field,omitempty"`
	Args           map[string]*llx.Primitive `protobuf:"bytes,6,rep,name=args,proto3" json:"args,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (x *DataReq) Reset() {
	*x = DataReq{}
	if protoimpl.UnsafeEnabled {
		mi := &file_providers_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DataReq) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DataReq) ProtoMessage() {}

func (x *DataReq) ProtoReflect() protoreflect.Message {
	mi := &file_providers_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DataReq.ProtoReflect.Descriptor instead.
func (*DataReq) Descriptor() ([]byte, []int) {
	return file_providers_proto_rawDescGZIP(), []int{4}
}

func (x *DataReq) GetConnection() uint32 {
	if x != nil {
		return x.Connection
	}
	return 0
}

func (x *DataReq) GetCallbackServer() uint32 {
	if x != nil {
		return x.CallbackServer
	}
	return 0
}

func (x *DataReq) GetResource() string {
	if x != nil {
		return x.Resource
	}
	return ""
}

func (x *DataReq) GetResourceId() string {
	if x != nil {
		return x.ResourceId
	}
	return ""
}

func (x *DataReq) GetField() string {
	if x != nil {
		return x.Field
	}
	return ""
}

func (x *DataReq) GetArgs() map[string]*llx.Primitive {
	if x != nil {
		return x.Args
	}
	return nil
}

type DataRes struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Data  *llx.Primitive `protobuf:"bytes,1,opt,name=data,proto3" json:"data,omitempty"`
	Error string         `protobuf:"bytes,2,opt,name=error,proto3" json:"error,omitempty"`
	// The ID uniquely identifies this request and all associated callbacks
	Id string `protobuf:"bytes,3,opt,name=id,proto3" json:"id,omitempty"`
}

func (x *DataRes) Reset() {
	*x = DataRes{}
	if protoimpl.UnsafeEnabled {
		mi := &file_providers_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DataRes) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DataRes) ProtoMessage() {}

func (x *DataRes) ProtoReflect() protoreflect.Message {
	mi := &file_providers_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DataRes.ProtoReflect.Descriptor instead.
func (*DataRes) Descriptor() ([]byte, []int) {
	return file_providers_proto_rawDescGZIP(), []int{5}
}

func (x *DataRes) GetData() *llx.Primitive {
	if x != nil {
		return x.Data
	}
	return nil
}

func (x *DataRes) GetError() string {
	if x != nil {
		return x.Error
	}
	return ""
}

func (x *DataRes) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

type CollectRes struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *CollectRes) Reset() {
	*x = CollectRes{}
	if protoimpl.UnsafeEnabled {
		mi := &file_providers_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CollectRes) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CollectRes) ProtoMessage() {}

func (x *CollectRes) ProtoReflect() protoreflect.Message {
	mi := &file_providers_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CollectRes.ProtoReflect.Descriptor instead.
func (*CollectRes) Descriptor() ([]byte, []int) {
	return file_providers_proto_rawDescGZIP(), []int{6}
}

var File_providers_proto protoreflect.FileDescriptor

var file_providers_proto_rawDesc = []byte{
	0x0a, 0x0f, 0x70, 0x72, 0x6f, 0x76, 0x69, 0x64, 0x65, 0x72, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x12, 0x05, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x22, 0x6d, 0x6f, 0x74, 0x6f, 0x72, 0x2f,
	0x69, 0x6e, 0x76, 0x65, 0x6e, 0x74, 0x6f, 0x72, 0x79, 0x2f, 0x76, 0x31, 0x2f, 0x69, 0x6e, 0x76,
	0x65, 0x6e, 0x74, 0x6f, 0x72, 0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x0d, 0x6c, 0x6c,
	0x78, 0x2f, 0x6c, 0x6c, 0x78, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xc6, 0x01, 0x0a, 0x0b,
	0x50, 0x61, 0x72, 0x73, 0x65, 0x43, 0x4c, 0x49, 0x52, 0x65, 0x71, 0x12, 0x1c, 0x0a, 0x09, 0x63,
	0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09,
	0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x12, 0x12, 0x0a, 0x04, 0x61, 0x72, 0x67,
	0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x09, 0x52, 0x04, 0x61, 0x72, 0x67, 0x73, 0x12, 0x33, 0x0a,
	0x05, 0x66, 0x6c, 0x61, 0x67, 0x73, 0x18, 0x03, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1d, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x50, 0x61, 0x72, 0x73, 0x65, 0x43, 0x4c, 0x49, 0x52, 0x65, 0x71,
	0x2e, 0x46, 0x6c, 0x61, 0x67, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x05, 0x66, 0x6c, 0x61,
	0x67, 0x73, 0x1a, 0x50, 0x0a, 0x0a, 0x46, 0x6c, 0x61, 0x67, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79,
	0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b,
	0x65, 0x79, 0x12, 0x2c, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x16, 0x2e, 0x63, 0x6e, 0x71, 0x75, 0x65, 0x72, 0x79, 0x2e, 0x6c, 0x6c, 0x78, 0x2e,
	0x50, 0x72, 0x69, 0x6d, 0x69, 0x74, 0x69, 0x76, 0x65, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65,
	0x3a, 0x02, 0x38, 0x01, 0x22, 0x52, 0x0a, 0x0b, 0x50, 0x61, 0x72, 0x73, 0x65, 0x43, 0x4c, 0x49,
	0x52, 0x65, 0x73, 0x12, 0x43, 0x0a, 0x09, 0x69, 0x6e, 0x76, 0x65, 0x6e, 0x74, 0x6f, 0x72, 0x79,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x25, 0x2e, 0x63, 0x6e, 0x71, 0x75, 0x65, 0x72, 0x79,
	0x2e, 0x6d, 0x6f, 0x74, 0x6f, 0x72, 0x2e, 0x69, 0x6e, 0x76, 0x65, 0x6e, 0x74, 0x6f, 0x72, 0x79,
	0x2e, 0x76, 0x31, 0x2e, 0x49, 0x6e, 0x76, 0x65, 0x6e, 0x74, 0x6f, 0x72, 0x79, 0x52, 0x09, 0x69,
	0x6e, 0x76, 0x65, 0x6e, 0x74, 0x6f, 0x72, 0x79, 0x22, 0x65, 0x0a, 0x0a, 0x43, 0x6f, 0x6e, 0x6e,
	0x65, 0x63, 0x74, 0x52, 0x65, 0x71, 0x12, 0x1a, 0x0a, 0x08, 0x66, 0x65, 0x61, 0x74, 0x75, 0x72,
	0x65, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x08, 0x66, 0x65, 0x61, 0x74, 0x75, 0x72,
	0x65, 0x73, 0x12, 0x3b, 0x0a, 0x05, 0x61, 0x73, 0x73, 0x65, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x25, 0x2e, 0x63, 0x6e, 0x71, 0x75, 0x65, 0x72, 0x79, 0x2e, 0x6d, 0x6f, 0x74, 0x6f,
	0x72, 0x2e, 0x69, 0x6e, 0x76, 0x65, 0x6e, 0x74, 0x6f, 0x72, 0x79, 0x2e, 0x76, 0x31, 0x2e, 0x49,
	0x6e, 0x76, 0x65, 0x6e, 0x74, 0x6f, 0x72, 0x79, 0x52, 0x05, 0x61, 0x73, 0x73, 0x65, 0x74, 0x22,
	0x1c, 0x0a, 0x0a, 0x43, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x0e, 0x0a,
	0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x02, 0x69, 0x64, 0x22, 0xa4, 0x02,
	0x0a, 0x07, 0x44, 0x61, 0x74, 0x61, 0x52, 0x65, 0x71, 0x12, 0x1e, 0x0a, 0x0a, 0x63, 0x6f, 0x6e,
	0x6e, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x0a, 0x63,
	0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x27, 0x0a, 0x0f, 0x63, 0x61, 0x6c,
	0x6c, 0x62, 0x61, 0x63, 0x6b, 0x5f, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x0d, 0x52, 0x0e, 0x63, 0x61, 0x6c, 0x6c, 0x62, 0x61, 0x63, 0x6b, 0x53, 0x65, 0x72, 0x76,
	0x65, 0x72, 0x12, 0x1a, 0x0a, 0x08, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x12, 0x1f,
	0x0a, 0x0b, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x5f, 0x69, 0x64, 0x18, 0x04, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x0a, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x49, 0x64, 0x12,
	0x14, 0x0a, 0x05, 0x66, 0x69, 0x65, 0x6c, 0x64, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05,
	0x66, 0x69, 0x65, 0x6c, 0x64, 0x12, 0x2c, 0x0a, 0x04, 0x61, 0x72, 0x67, 0x73, 0x18, 0x06, 0x20,
	0x03, 0x28, 0x0b, 0x32, 0x18, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x44, 0x61, 0x74, 0x61,
	0x52, 0x65, 0x71, 0x2e, 0x41, 0x72, 0x67, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x04, 0x61,
	0x72, 0x67, 0x73, 0x1a, 0x4f, 0x0a, 0x09, 0x41, 0x72, 0x67, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79,
	0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b,
	0x65, 0x79, 0x12, 0x2c, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x16, 0x2e, 0x63, 0x6e, 0x71, 0x75, 0x65, 0x72, 0x79, 0x2e, 0x6c, 0x6c, 0x78, 0x2e,
	0x50, 0x72, 0x69, 0x6d, 0x69, 0x74, 0x69, 0x76, 0x65, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65,
	0x3a, 0x02, 0x38, 0x01, 0x22, 0x5b, 0x0a, 0x07, 0x44, 0x61, 0x74, 0x61, 0x52, 0x65, 0x73, 0x12,
	0x2a, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x61, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x16, 0x2e,
	0x63, 0x6e, 0x71, 0x75, 0x65, 0x72, 0x79, 0x2e, 0x6c, 0x6c, 0x78, 0x2e, 0x50, 0x72, 0x69, 0x6d,
	0x69, 0x74, 0x69, 0x76, 0x65, 0x52, 0x04, 0x64, 0x61, 0x74, 0x61, 0x12, 0x14, 0x0a, 0x05, 0x65,
	0x72, 0x72, 0x6f, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x65, 0x72, 0x72, 0x6f,
	0x72, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69,
	0x64, 0x22, 0x0c, 0x0a, 0x0a, 0x43, 0x6f, 0x6c, 0x6c, 0x65, 0x63, 0x74, 0x52, 0x65, 0x73, 0x32,
	0xa0, 0x01, 0x0a, 0x0e, 0x50, 0x72, 0x6f, 0x76, 0x69, 0x64, 0x65, 0x72, 0x50, 0x6c, 0x75, 0x67,
	0x69, 0x6e, 0x12, 0x32, 0x0a, 0x08, 0x50, 0x61, 0x72, 0x73, 0x65, 0x43, 0x4c, 0x49, 0x12, 0x12,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x50, 0x61, 0x72, 0x73, 0x65, 0x43, 0x4c, 0x49, 0x52,
	0x65, 0x71, 0x1a, 0x12, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x50, 0x61, 0x72, 0x73, 0x65,
	0x43, 0x4c, 0x49, 0x52, 0x65, 0x73, 0x12, 0x2f, 0x0a, 0x07, 0x43, 0x6f, 0x6e, 0x6e, 0x65, 0x63,
	0x74, 0x12, 0x11, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x43, 0x6f, 0x6e, 0x6e, 0x65, 0x63,
	0x74, 0x52, 0x65, 0x71, 0x1a, 0x11, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x43, 0x6f, 0x6e,
	0x6e, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x29, 0x0a, 0x07, 0x47, 0x65, 0x74, 0x44, 0x61,
	0x74, 0x61, 0x12, 0x0e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x44, 0x61, 0x74, 0x61, 0x52,
	0x65, 0x71, 0x1a, 0x0e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x44, 0x61, 0x74, 0x61, 0x52,
	0x65, 0x73, 0x32, 0x40, 0x0a, 0x10, 0x50, 0x72, 0x6f, 0x76, 0x69, 0x64, 0x65, 0x72, 0x43, 0x61,
	0x6c, 0x6c, 0x62, 0x61, 0x63, 0x6b, 0x12, 0x2c, 0x0a, 0x07, 0x43, 0x6f, 0x6c, 0x6c, 0x65, 0x63,
	0x74, 0x12, 0x0e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x44, 0x61, 0x74, 0x61, 0x52, 0x65,
	0x73, 0x1a, 0x11, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x43, 0x6f, 0x6c, 0x6c, 0x65, 0x63,
	0x74, 0x52, 0x65, 0x73, 0x42, 0x27, 0x5a, 0x25, 0x67, 0x6f, 0x2e, 0x6d, 0x6f, 0x6e, 0x64, 0x6f,
	0x6f, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x63, 0x6e, 0x71, 0x75, 0x65, 0x72, 0x79, 0x2f, 0x70, 0x72,
	0x6f, 0x76, 0x69, 0x64, 0x65, 0x72, 0x73, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x06, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_providers_proto_rawDescOnce sync.Once
	file_providers_proto_rawDescData = file_providers_proto_rawDesc
)

func file_providers_proto_rawDescGZIP() []byte {
	file_providers_proto_rawDescOnce.Do(func() {
		file_providers_proto_rawDescData = protoimpl.X.CompressGZIP(file_providers_proto_rawDescData)
	})
	return file_providers_proto_rawDescData
}

var file_providers_proto_msgTypes = make([]protoimpl.MessageInfo, 9)
var file_providers_proto_goTypes = []interface{}{
	(*ParseCLIReq)(nil),   // 0: proto.ParseCLIReq
	(*ParseCLIRes)(nil),   // 1: proto.ParseCLIRes
	(*ConnectReq)(nil),    // 2: proto.ConnectReq
	(*Connection)(nil),    // 3: proto.Connection
	(*DataReq)(nil),       // 4: proto.DataReq
	(*DataRes)(nil),       // 5: proto.DataRes
	(*CollectRes)(nil),    // 6: proto.CollectRes
	nil,                   // 7: proto.ParseCLIReq.FlagsEntry
	nil,                   // 8: proto.DataReq.ArgsEntry
	(*v1.Inventory)(nil),  // 9: cnquery.motor.inventory.v1.Inventory
	(*llx.Primitive)(nil), // 10: cnquery.llx.Primitive
}
var file_providers_proto_depIdxs = []int32{
	7,  // 0: proto.ParseCLIReq.flags:type_name -> proto.ParseCLIReq.FlagsEntry
	9,  // 1: proto.ParseCLIRes.inventory:type_name -> cnquery.motor.inventory.v1.Inventory
	9,  // 2: proto.ConnectReq.asset:type_name -> cnquery.motor.inventory.v1.Inventory
	8,  // 3: proto.DataReq.args:type_name -> proto.DataReq.ArgsEntry
	10, // 4: proto.DataRes.data:type_name -> cnquery.llx.Primitive
	10, // 5: proto.ParseCLIReq.FlagsEntry.value:type_name -> cnquery.llx.Primitive
	10, // 6: proto.DataReq.ArgsEntry.value:type_name -> cnquery.llx.Primitive
	0,  // 7: proto.ProviderPlugin.ParseCLI:input_type -> proto.ParseCLIReq
	2,  // 8: proto.ProviderPlugin.Connect:input_type -> proto.ConnectReq
	4,  // 9: proto.ProviderPlugin.GetData:input_type -> proto.DataReq
	5,  // 10: proto.ProviderCallback.Collect:input_type -> proto.DataRes
	1,  // 11: proto.ProviderPlugin.ParseCLI:output_type -> proto.ParseCLIRes
	3,  // 12: proto.ProviderPlugin.Connect:output_type -> proto.Connection
	5,  // 13: proto.ProviderPlugin.GetData:output_type -> proto.DataRes
	6,  // 14: proto.ProviderCallback.Collect:output_type -> proto.CollectRes
	11, // [11:15] is the sub-list for method output_type
	7,  // [7:11] is the sub-list for method input_type
	7,  // [7:7] is the sub-list for extension type_name
	7,  // [7:7] is the sub-list for extension extendee
	0,  // [0:7] is the sub-list for field type_name
}

func init() { file_providers_proto_init() }
func file_providers_proto_init() {
	if File_providers_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_providers_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ParseCLIReq); i {
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
		file_providers_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ParseCLIRes); i {
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
		file_providers_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ConnectReq); i {
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
		file_providers_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Connection); i {
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
		file_providers_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DataReq); i {
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
		file_providers_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DataRes); i {
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
		file_providers_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CollectRes); i {
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
			RawDescriptor: file_providers_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   9,
			NumExtensions: 0,
			NumServices:   2,
		},
		GoTypes:           file_providers_proto_goTypes,
		DependencyIndexes: file_providers_proto_depIdxs,
		MessageInfos:      file_providers_proto_msgTypes,
	}.Build()
	File_providers_proto = out.File
	file_providers_proto_rawDesc = nil
	file_providers_proto_goTypes = nil
	file_providers_proto_depIdxs = nil
}
