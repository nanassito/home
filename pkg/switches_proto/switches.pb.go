// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v3.12.4
// source: switches.proto

package switches_proto

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

type ReqList struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *ReqList) Reset() {
	*x = ReqList{}
	if protoimpl.UnsafeEnabled {
		mi := &file_switches_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ReqList) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ReqList) ProtoMessage() {}

func (x *ReqList) ProtoReflect() protoreflect.Message {
	mi := &file_switches_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ReqList.ProtoReflect.Descriptor instead.
func (*ReqList) Descriptor() ([]byte, []int) {
	return file_switches_proto_rawDescGZIP(), []int{0}
}

type RspList struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	SwitchIDs []string `protobuf:"bytes,1,rep,name=SwitchIDs,proto3" json:"SwitchIDs,omitempty"`
}

func (x *RspList) Reset() {
	*x = RspList{}
	if protoimpl.UnsafeEnabled {
		mi := &file_switches_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RspList) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RspList) ProtoMessage() {}

func (x *RspList) ProtoReflect() protoreflect.Message {
	mi := &file_switches_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RspList.ProtoReflect.Descriptor instead.
func (*RspList) Descriptor() ([]byte, []int) {
	return file_switches_proto_rawDescGZIP(), []int{1}
}

func (x *RspList) GetSwitchIDs() []string {
	if x != nil {
		return x.SwitchIDs
	}
	return nil
}

type ReqActivate struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	SwitchID        string `protobuf:"bytes,1,opt,name=SwitchID,proto3" json:"SwitchID,omitempty"`
	DurationSeconds int64  `protobuf:"varint,2,opt,name=DurationSeconds,proto3" json:"DurationSeconds,omitempty"`
	ClientID        string `protobuf:"bytes,3,opt,name=ClientID,proto3" json:"ClientID,omitempty"`
}

func (x *ReqActivate) Reset() {
	*x = ReqActivate{}
	if protoimpl.UnsafeEnabled {
		mi := &file_switches_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ReqActivate) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ReqActivate) ProtoMessage() {}

func (x *ReqActivate) ProtoReflect() protoreflect.Message {
	mi := &file_switches_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ReqActivate.ProtoReflect.Descriptor instead.
func (*ReqActivate) Descriptor() ([]byte, []int) {
	return file_switches_proto_rawDescGZIP(), []int{2}
}

func (x *ReqActivate) GetSwitchID() string {
	if x != nil {
		return x.SwitchID
	}
	return ""
}

func (x *ReqActivate) GetDurationSeconds() int64 {
	if x != nil {
		return x.DurationSeconds
	}
	return 0
}

func (x *ReqActivate) GetClientID() string {
	if x != nil {
		return x.ClientID
	}
	return ""
}

type RspActivate struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	SwitchID    string `protobuf:"bytes,1,opt,name=SwitchID,proto3" json:"SwitchID,omitempty"`
	ActiveUntil int64  `protobuf:"varint,2,opt,name=ActiveUntil,proto3" json:"ActiveUntil,omitempty"`
}

func (x *RspActivate) Reset() {
	*x = RspActivate{}
	if protoimpl.UnsafeEnabled {
		mi := &file_switches_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RspActivate) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RspActivate) ProtoMessage() {}

func (x *RspActivate) ProtoReflect() protoreflect.Message {
	mi := &file_switches_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RspActivate.ProtoReflect.Descriptor instead.
func (*RspActivate) Descriptor() ([]byte, []int) {
	return file_switches_proto_rawDescGZIP(), []int{3}
}

func (x *RspActivate) GetSwitchID() string {
	if x != nil {
		return x.SwitchID
	}
	return ""
}

func (x *RspActivate) GetActiveUntil() int64 {
	if x != nil {
		return x.ActiveUntil
	}
	return 0
}

type ReqStatus struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	SwitchID string `protobuf:"bytes,1,opt,name=SwitchID,proto3" json:"SwitchID,omitempty"`
}

func (x *ReqStatus) Reset() {
	*x = ReqStatus{}
	if protoimpl.UnsafeEnabled {
		mi := &file_switches_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ReqStatus) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ReqStatus) ProtoMessage() {}

func (x *ReqStatus) ProtoReflect() protoreflect.Message {
	mi := &file_switches_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ReqStatus.ProtoReflect.Descriptor instead.
func (*ReqStatus) Descriptor() ([]byte, []int) {
	return file_switches_proto_rawDescGZIP(), []int{4}
}

func (x *ReqStatus) GetSwitchID() string {
	if x != nil {
		return x.SwitchID
	}
	return ""
}

type RspStatus struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	SwitchID    string               `protobuf:"bytes,1,opt,name=SwitchID,proto3" json:"SwitchID,omitempty"`
	IsActive    bool                 `protobuf:"varint,2,opt,name=isActive,proto3" json:"isActive,omitempty"`
	ActiveUntil *int64               `protobuf:"varint,3,opt,name=ActiveUntil,proto3,oneof" json:"ActiveUntil,omitempty"`
	Requests    []*RspStatus_Request `protobuf:"bytes,4,rep,name=Requests,proto3" json:"Requests,omitempty"`
}

func (x *RspStatus) Reset() {
	*x = RspStatus{}
	if protoimpl.UnsafeEnabled {
		mi := &file_switches_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RspStatus) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RspStatus) ProtoMessage() {}

func (x *RspStatus) ProtoReflect() protoreflect.Message {
	mi := &file_switches_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RspStatus.ProtoReflect.Descriptor instead.
func (*RspStatus) Descriptor() ([]byte, []int) {
	return file_switches_proto_rawDescGZIP(), []int{5}
}

func (x *RspStatus) GetSwitchID() string {
	if x != nil {
		return x.SwitchID
	}
	return ""
}

func (x *RspStatus) GetIsActive() bool {
	if x != nil {
		return x.IsActive
	}
	return false
}

func (x *RspStatus) GetActiveUntil() int64 {
	if x != nil && x.ActiveUntil != nil {
		return *x.ActiveUntil
	}
	return 0
}

func (x *RspStatus) GetRequests() []*RspStatus_Request {
	if x != nil {
		return x.Requests
	}
	return nil
}

// For internal use only. Used to load init data from disk.
type SwitchConfig struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Mqtt       *SwitchConfig_MqttConfig       `protobuf:"bytes,1,opt,name=Mqtt,proto3" json:"Mqtt,omitempty"`
	Prometheus *SwitchConfig_PrometheusConfig `protobuf:"bytes,2,opt,name=Prometheus,proto3" json:"Prometheus,omitempty"`
}

func (x *SwitchConfig) Reset() {
	*x = SwitchConfig{}
	if protoimpl.UnsafeEnabled {
		mi := &file_switches_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SwitchConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SwitchConfig) ProtoMessage() {}

func (x *SwitchConfig) ProtoReflect() protoreflect.Message {
	mi := &file_switches_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SwitchConfig.ProtoReflect.Descriptor instead.
func (*SwitchConfig) Descriptor() ([]byte, []int) {
	return file_switches_proto_rawDescGZIP(), []int{6}
}

func (x *SwitchConfig) GetMqtt() *SwitchConfig_MqttConfig {
	if x != nil {
		return x.Mqtt
	}
	return nil
}

func (x *SwitchConfig) GetPrometheus() *SwitchConfig_PrometheusConfig {
	if x != nil {
		return x.Prometheus
	}
	return nil
}

type RspStatus_Request struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ClientID string `protobuf:"bytes,1,opt,name=ClientID,proto3" json:"ClientID,omitempty"`
	From     int64  `protobuf:"varint,2,opt,name=From,proto3" json:"From,omitempty"`
	To       int64  `protobuf:"varint,3,opt,name=To,proto3" json:"To,omitempty"`
}

func (x *RspStatus_Request) Reset() {
	*x = RspStatus_Request{}
	if protoimpl.UnsafeEnabled {
		mi := &file_switches_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RspStatus_Request) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RspStatus_Request) ProtoMessage() {}

func (x *RspStatus_Request) ProtoReflect() protoreflect.Message {
	mi := &file_switches_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RspStatus_Request.ProtoReflect.Descriptor instead.
func (*RspStatus_Request) Descriptor() ([]byte, []int) {
	return file_switches_proto_rawDescGZIP(), []int{5, 0}
}

func (x *RspStatus_Request) GetClientID() string {
	if x != nil {
		return x.ClientID
	}
	return ""
}

func (x *RspStatus_Request) GetFrom() int64 {
	if x != nil {
		return x.From
	}
	return 0
}

func (x *RspStatus_Request) GetTo() int64 {
	if x != nil {
		return x.To
	}
	return 0
}

type SwitchConfig_MqttConfig struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	SetTopic  string `protobuf:"bytes,1,opt,name=SetTopic,proto3" json:"SetTopic,omitempty"`
	MsgRest   string `protobuf:"bytes,2,opt,name=MsgRest,proto3" json:"MsgRest,omitempty"`
	MsgActive string `protobuf:"bytes,3,opt,name=MsgActive,proto3" json:"MsgActive,omitempty"`
}

func (x *SwitchConfig_MqttConfig) Reset() {
	*x = SwitchConfig_MqttConfig{}
	if protoimpl.UnsafeEnabled {
		mi := &file_switches_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SwitchConfig_MqttConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SwitchConfig_MqttConfig) ProtoMessage() {}

func (x *SwitchConfig_MqttConfig) ProtoReflect() protoreflect.Message {
	mi := &file_switches_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SwitchConfig_MqttConfig.ProtoReflect.Descriptor instead.
func (*SwitchConfig_MqttConfig) Descriptor() ([]byte, []int) {
	return file_switches_proto_rawDescGZIP(), []int{6, 0}
}

func (x *SwitchConfig_MqttConfig) GetSetTopic() string {
	if x != nil {
		return x.SetTopic
	}
	return ""
}

func (x *SwitchConfig_MqttConfig) GetMsgRest() string {
	if x != nil {
		return x.MsgRest
	}
	return ""
}

func (x *SwitchConfig_MqttConfig) GetMsgActive() string {
	if x != nil {
		return x.MsgActive
	}
	return ""
}

type SwitchConfig_PrometheusConfig struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Metric      string            `protobuf:"bytes,1,opt,name=Metric,proto3" json:"Metric,omitempty"`
	Labels      map[string]string `protobuf:"bytes,2,rep,name=Labels,proto3" json:"Labels,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	ValueRest   int32             `protobuf:"varint,3,opt,name=ValueRest,proto3" json:"ValueRest,omitempty"`
	ValueActive int32             `protobuf:"varint,4,opt,name=ValueActive,proto3" json:"ValueActive,omitempty"`
}

func (x *SwitchConfig_PrometheusConfig) Reset() {
	*x = SwitchConfig_PrometheusConfig{}
	if protoimpl.UnsafeEnabled {
		mi := &file_switches_proto_msgTypes[9]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SwitchConfig_PrometheusConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SwitchConfig_PrometheusConfig) ProtoMessage() {}

func (x *SwitchConfig_PrometheusConfig) ProtoReflect() protoreflect.Message {
	mi := &file_switches_proto_msgTypes[9]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SwitchConfig_PrometheusConfig.ProtoReflect.Descriptor instead.
func (*SwitchConfig_PrometheusConfig) Descriptor() ([]byte, []int) {
	return file_switches_proto_rawDescGZIP(), []int{6, 1}
}

func (x *SwitchConfig_PrometheusConfig) GetMetric() string {
	if x != nil {
		return x.Metric
	}
	return ""
}

func (x *SwitchConfig_PrometheusConfig) GetLabels() map[string]string {
	if x != nil {
		return x.Labels
	}
	return nil
}

func (x *SwitchConfig_PrometheusConfig) GetValueRest() int32 {
	if x != nil {
		return x.ValueRest
	}
	return 0
}

func (x *SwitchConfig_PrometheusConfig) GetValueActive() int32 {
	if x != nil {
		return x.ValueActive
	}
	return 0
}

var File_switches_proto protoreflect.FileDescriptor

var file_switches_proto_rawDesc = []byte{
	0x0a, 0x0e, 0x73, 0x77, 0x69, 0x74, 0x63, 0x68, 0x65, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x12, 0x0e, 0x73, 0x77, 0x69, 0x74, 0x63, 0x68, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65,
	0x1a, 0x1c, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x61, 0x6e, 0x6e,
	0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x09,
	0x0a, 0x07, 0x52, 0x65, 0x71, 0x4c, 0x69, 0x73, 0x74, 0x22, 0x27, 0x0a, 0x07, 0x52, 0x73, 0x70,
	0x4c, 0x69, 0x73, 0x74, 0x12, 0x1c, 0x0a, 0x09, 0x53, 0x77, 0x69, 0x74, 0x63, 0x68, 0x49, 0x44,
	0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x09, 0x52, 0x09, 0x53, 0x77, 0x69, 0x74, 0x63, 0x68, 0x49,
	0x44, 0x73, 0x22, 0x6f, 0x0a, 0x0b, 0x52, 0x65, 0x71, 0x41, 0x63, 0x74, 0x69, 0x76, 0x61, 0x74,
	0x65, 0x12, 0x1a, 0x0a, 0x08, 0x53, 0x77, 0x69, 0x74, 0x63, 0x68, 0x49, 0x44, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x08, 0x53, 0x77, 0x69, 0x74, 0x63, 0x68, 0x49, 0x44, 0x12, 0x28, 0x0a,
	0x0f, 0x44, 0x75, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x53, 0x65, 0x63, 0x6f, 0x6e, 0x64, 0x73,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x0f, 0x44, 0x75, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e,
	0x53, 0x65, 0x63, 0x6f, 0x6e, 0x64, 0x73, 0x12, 0x1a, 0x0a, 0x08, 0x43, 0x6c, 0x69, 0x65, 0x6e,
	0x74, 0x49, 0x44, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x43, 0x6c, 0x69, 0x65, 0x6e,
	0x74, 0x49, 0x44, 0x22, 0x4b, 0x0a, 0x0b, 0x52, 0x73, 0x70, 0x41, 0x63, 0x74, 0x69, 0x76, 0x61,
	0x74, 0x65, 0x12, 0x1a, 0x0a, 0x08, 0x53, 0x77, 0x69, 0x74, 0x63, 0x68, 0x49, 0x44, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x53, 0x77, 0x69, 0x74, 0x63, 0x68, 0x49, 0x44, 0x12, 0x20,
	0x0a, 0x0b, 0x41, 0x63, 0x74, 0x69, 0x76, 0x65, 0x55, 0x6e, 0x74, 0x69, 0x6c, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x03, 0x52, 0x0b, 0x41, 0x63, 0x74, 0x69, 0x76, 0x65, 0x55, 0x6e, 0x74, 0x69, 0x6c,
	0x22, 0x27, 0x0a, 0x09, 0x52, 0x65, 0x71, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x1a, 0x0a,
	0x08, 0x53, 0x77, 0x69, 0x74, 0x63, 0x68, 0x49, 0x44, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x08, 0x53, 0x77, 0x69, 0x74, 0x63, 0x68, 0x49, 0x44, 0x22, 0x84, 0x02, 0x0a, 0x09, 0x52, 0x73,
	0x70, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x1a, 0x0a, 0x08, 0x53, 0x77, 0x69, 0x74, 0x63,
	0x68, 0x49, 0x44, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x53, 0x77, 0x69, 0x74, 0x63,
	0x68, 0x49, 0x44, 0x12, 0x1a, 0x0a, 0x08, 0x69, 0x73, 0x41, 0x63, 0x74, 0x69, 0x76, 0x65, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x08, 0x52, 0x08, 0x69, 0x73, 0x41, 0x63, 0x74, 0x69, 0x76, 0x65, 0x12,
	0x25, 0x0a, 0x0b, 0x41, 0x63, 0x74, 0x69, 0x76, 0x65, 0x55, 0x6e, 0x74, 0x69, 0x6c, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x03, 0x48, 0x00, 0x52, 0x0b, 0x41, 0x63, 0x74, 0x69, 0x76, 0x65, 0x55, 0x6e,
	0x74, 0x69, 0x6c, 0x88, 0x01, 0x01, 0x12, 0x3d, 0x0a, 0x08, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x73, 0x18, 0x04, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x21, 0x2e, 0x73, 0x77, 0x69, 0x74, 0x63,
	0x68, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x52, 0x73, 0x70, 0x53, 0x74, 0x61,
	0x74, 0x75, 0x73, 0x2e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x52, 0x08, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x73, 0x1a, 0x49, 0x0a, 0x07, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x12, 0x1a, 0x0a, 0x08, 0x43, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x49, 0x44, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x08, 0x43, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x49, 0x44, 0x12, 0x12, 0x0a, 0x04,
	0x46, 0x72, 0x6f, 0x6d, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x04, 0x46, 0x72, 0x6f, 0x6d,
	0x12, 0x0e, 0x0a, 0x02, 0x54, 0x6f, 0x18, 0x03, 0x20, 0x01, 0x28, 0x03, 0x52, 0x02, 0x54, 0x6f,
	0x42, 0x0e, 0x0a, 0x0c, 0x5f, 0x41, 0x63, 0x74, 0x69, 0x76, 0x65, 0x55, 0x6e, 0x74, 0x69, 0x6c,
	0x22, 0xf7, 0x03, 0x0a, 0x0c, 0x53, 0x77, 0x69, 0x74, 0x63, 0x68, 0x43, 0x6f, 0x6e, 0x66, 0x69,
	0x67, 0x12, 0x3b, 0x0a, 0x04, 0x4d, 0x71, 0x74, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x27, 0x2e, 0x73, 0x77, 0x69, 0x74, 0x63, 0x68, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65,
	0x2e, 0x53, 0x77, 0x69, 0x74, 0x63, 0x68, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x2e, 0x4d, 0x71,
	0x74, 0x74, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x52, 0x04, 0x4d, 0x71, 0x74, 0x74, 0x12, 0x4d,
	0x0a, 0x0a, 0x50, 0x72, 0x6f, 0x6d, 0x65, 0x74, 0x68, 0x65, 0x75, 0x73, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x2d, 0x2e, 0x73, 0x77, 0x69, 0x74, 0x63, 0x68, 0x2e, 0x73, 0x65, 0x72, 0x76,
	0x69, 0x63, 0x65, 0x2e, 0x53, 0x77, 0x69, 0x74, 0x63, 0x68, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67,
	0x2e, 0x50, 0x72, 0x6f, 0x6d, 0x65, 0x74, 0x68, 0x65, 0x75, 0x73, 0x43, 0x6f, 0x6e, 0x66, 0x69,
	0x67, 0x52, 0x0a, 0x50, 0x72, 0x6f, 0x6d, 0x65, 0x74, 0x68, 0x65, 0x75, 0x73, 0x1a, 0x60, 0x0a,
	0x0a, 0x4d, 0x71, 0x74, 0x74, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x12, 0x1a, 0x0a, 0x08, 0x53,
	0x65, 0x74, 0x54, 0x6f, 0x70, 0x69, 0x63, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x53,
	0x65, 0x74, 0x54, 0x6f, 0x70, 0x69, 0x63, 0x12, 0x18, 0x0a, 0x07, 0x4d, 0x73, 0x67, 0x52, 0x65,
	0x73, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x4d, 0x73, 0x67, 0x52, 0x65, 0x73,
	0x74, 0x12, 0x1c, 0x0a, 0x09, 0x4d, 0x73, 0x67, 0x41, 0x63, 0x74, 0x69, 0x76, 0x65, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x4d, 0x73, 0x67, 0x41, 0x63, 0x74, 0x69, 0x76, 0x65, 0x1a,
	0xf8, 0x01, 0x0a, 0x10, 0x50, 0x72, 0x6f, 0x6d, 0x65, 0x74, 0x68, 0x65, 0x75, 0x73, 0x43, 0x6f,
	0x6e, 0x66, 0x69, 0x67, 0x12, 0x16, 0x0a, 0x06, 0x4d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x4d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x12, 0x51, 0x0a, 0x06,
	0x4c, 0x61, 0x62, 0x65, 0x6c, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x39, 0x2e, 0x73,
	0x77, 0x69, 0x74, 0x63, 0x68, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x53, 0x77,
	0x69, 0x74, 0x63, 0x68, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x2e, 0x50, 0x72, 0x6f, 0x6d, 0x65,
	0x74, 0x68, 0x65, 0x75, 0x73, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x2e, 0x4c, 0x61, 0x62, 0x65,
	0x6c, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x06, 0x4c, 0x61, 0x62, 0x65, 0x6c, 0x73, 0x12,
	0x1c, 0x0a, 0x09, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x52, 0x65, 0x73, 0x74, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x05, 0x52, 0x09, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x52, 0x65, 0x73, 0x74, 0x12, 0x20, 0x0a,
	0x0b, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x41, 0x63, 0x74, 0x69, 0x76, 0x65, 0x18, 0x04, 0x20, 0x01,
	0x28, 0x05, 0x52, 0x0b, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x41, 0x63, 0x74, 0x69, 0x76, 0x65, 0x1a,
	0x39, 0x0a, 0x0b, 0x4c, 0x61, 0x62, 0x65, 0x6c, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10,
	0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79,
	0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x32, 0xfe, 0x01, 0x0a, 0x09, 0x53,
	0x77, 0x69, 0x74, 0x63, 0x68, 0x53, 0x76, 0x63, 0x12, 0x47, 0x0a, 0x04, 0x4c, 0x69, 0x73, 0x74,
	0x12, 0x17, 0x2e, 0x73, 0x77, 0x69, 0x74, 0x63, 0x68, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63,
	0x65, 0x2e, 0x52, 0x65, 0x71, 0x4c, 0x69, 0x73, 0x74, 0x1a, 0x17, 0x2e, 0x73, 0x77, 0x69, 0x74,
	0x63, 0x68, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x52, 0x73, 0x70, 0x4c, 0x69,
	0x73, 0x74, 0x22, 0x0d, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x07, 0x12, 0x05, 0x2f, 0x6c, 0x69, 0x73,
	0x74, 0x12, 0x57, 0x0a, 0x08, 0x41, 0x63, 0x74, 0x69, 0x76, 0x61, 0x74, 0x65, 0x12, 0x1b, 0x2e,
	0x73, 0x77, 0x69, 0x74, 0x63, 0x68, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x52,
	0x65, 0x71, 0x41, 0x63, 0x74, 0x69, 0x76, 0x61, 0x74, 0x65, 0x1a, 0x1b, 0x2e, 0x73, 0x77, 0x69,
	0x74, 0x63, 0x68, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x52, 0x73, 0x70, 0x41,
	0x63, 0x74, 0x69, 0x76, 0x61, 0x74, 0x65, 0x22, 0x11, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x0b, 0x22,
	0x09, 0x2f, 0x61, 0x63, 0x74, 0x69, 0x76, 0x61, 0x74, 0x65, 0x12, 0x4f, 0x0a, 0x06, 0x53, 0x74,
	0x61, 0x74, 0x75, 0x73, 0x12, 0x19, 0x2e, 0x73, 0x77, 0x69, 0x74, 0x63, 0x68, 0x2e, 0x73, 0x65,
	0x72, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x52, 0x65, 0x71, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x1a,
	0x19, 0x2e, 0x73, 0x77, 0x69, 0x74, 0x63, 0x68, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65,
	0x2e, 0x52, 0x73, 0x70, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x22, 0x0f, 0x82, 0xd3, 0xe4, 0x93,
	0x02, 0x09, 0x12, 0x07, 0x2f, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x42, 0x3d, 0x5a, 0x3b, 0x67,
	0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6e, 0x61, 0x6e, 0x61, 0x73, 0x73,
	0x69, 0x74, 0x6f, 0x2f, 0x68, 0x6f, 0x6d, 0x65, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x2f, 0x73, 0x77, 0x69, 0x74, 0x63, 0x68, 0x65, 0x73, 0x3b, 0x73, 0x77, 0x69, 0x74,
	0x63, 0x68, 0x65, 0x73, 0x5f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x33,
}

var (
	file_switches_proto_rawDescOnce sync.Once
	file_switches_proto_rawDescData = file_switches_proto_rawDesc
)

func file_switches_proto_rawDescGZIP() []byte {
	file_switches_proto_rawDescOnce.Do(func() {
		file_switches_proto_rawDescData = protoimpl.X.CompressGZIP(file_switches_proto_rawDescData)
	})
	return file_switches_proto_rawDescData
}

var file_switches_proto_msgTypes = make([]protoimpl.MessageInfo, 11)
var file_switches_proto_goTypes = []interface{}{
	(*ReqList)(nil),                       // 0: switch.service.ReqList
	(*RspList)(nil),                       // 1: switch.service.RspList
	(*ReqActivate)(nil),                   // 2: switch.service.ReqActivate
	(*RspActivate)(nil),                   // 3: switch.service.RspActivate
	(*ReqStatus)(nil),                     // 4: switch.service.ReqStatus
	(*RspStatus)(nil),                     // 5: switch.service.RspStatus
	(*SwitchConfig)(nil),                  // 6: switch.service.SwitchConfig
	(*RspStatus_Request)(nil),             // 7: switch.service.RspStatus.Request
	(*SwitchConfig_MqttConfig)(nil),       // 8: switch.service.SwitchConfig.MqttConfig
	(*SwitchConfig_PrometheusConfig)(nil), // 9: switch.service.SwitchConfig.PrometheusConfig
	nil,                                   // 10: switch.service.SwitchConfig.PrometheusConfig.LabelsEntry
}
var file_switches_proto_depIdxs = []int32{
	7,  // 0: switch.service.RspStatus.Requests:type_name -> switch.service.RspStatus.Request
	8,  // 1: switch.service.SwitchConfig.Mqtt:type_name -> switch.service.SwitchConfig.MqttConfig
	9,  // 2: switch.service.SwitchConfig.Prometheus:type_name -> switch.service.SwitchConfig.PrometheusConfig
	10, // 3: switch.service.SwitchConfig.PrometheusConfig.Labels:type_name -> switch.service.SwitchConfig.PrometheusConfig.LabelsEntry
	0,  // 4: switch.service.SwitchSvc.List:input_type -> switch.service.ReqList
	2,  // 5: switch.service.SwitchSvc.Activate:input_type -> switch.service.ReqActivate
	4,  // 6: switch.service.SwitchSvc.Status:input_type -> switch.service.ReqStatus
	1,  // 7: switch.service.SwitchSvc.List:output_type -> switch.service.RspList
	3,  // 8: switch.service.SwitchSvc.Activate:output_type -> switch.service.RspActivate
	5,  // 9: switch.service.SwitchSvc.Status:output_type -> switch.service.RspStatus
	7,  // [7:10] is the sub-list for method output_type
	4,  // [4:7] is the sub-list for method input_type
	4,  // [4:4] is the sub-list for extension type_name
	4,  // [4:4] is the sub-list for extension extendee
	0,  // [0:4] is the sub-list for field type_name
}

func init() { file_switches_proto_init() }
func file_switches_proto_init() {
	if File_switches_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_switches_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ReqList); i {
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
		file_switches_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RspList); i {
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
		file_switches_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ReqActivate); i {
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
		file_switches_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RspActivate); i {
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
		file_switches_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ReqStatus); i {
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
		file_switches_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RspStatus); i {
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
		file_switches_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SwitchConfig); i {
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
		file_switches_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RspStatus_Request); i {
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
		file_switches_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SwitchConfig_MqttConfig); i {
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
		file_switches_proto_msgTypes[9].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SwitchConfig_PrometheusConfig); i {
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
	file_switches_proto_msgTypes[5].OneofWrappers = []interface{}{}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_switches_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   11,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_switches_proto_goTypes,
		DependencyIndexes: file_switches_proto_depIdxs,
		MessageInfos:      file_switches_proto_msgTypes,
	}.Build()
	File_switches_proto = out.File
	file_switches_proto_rawDesc = nil
	file_switches_proto_goTypes = nil
	file_switches_proto_depIdxs = nil
}
