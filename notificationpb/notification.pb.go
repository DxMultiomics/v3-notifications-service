// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.6
// 	protoc        v3.21.7
// source: proto/notification.proto

package notificationpb

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
	unsafe "unsafe"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// The notification message structure
type Notification struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Id            string                 `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Type          string                 `protobuf:"bytes,2,opt,name=type,proto3" json:"type,omitempty"`
	Message       string                 `protobuf:"bytes,3,opt,name=message,proto3" json:"message,omitempty"`
	Timestamp     string                 `protobuf:"bytes,4,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Notification) Reset() {
	*x = Notification{}
	mi := &file_proto_notification_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Notification) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Notification) ProtoMessage() {}

func (x *Notification) ProtoReflect() protoreflect.Message {
	mi := &file_proto_notification_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Notification.ProtoReflect.Descriptor instead.
func (*Notification) Descriptor() ([]byte, []int) {
	return file_proto_notification_proto_rawDescGZIP(), []int{0}
}

func (x *Notification) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *Notification) GetType() string {
	if x != nil {
		return x.Type
	}
	return ""
}

func (x *Notification) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

func (x *Notification) GetTimestamp() string {
	if x != nil {
		return x.Timestamp
	}
	return ""
}

// Request to start streaming notifications
type StreamNotificationsRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	UserId        string                 `protobuf:"bytes,1,opt,name=user_id,json=userId,proto3" json:"user_id,omitempty"` // To identify the client
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *StreamNotificationsRequest) Reset() {
	*x = StreamNotificationsRequest{}
	mi := &file_proto_notification_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *StreamNotificationsRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StreamNotificationsRequest) ProtoMessage() {}

func (x *StreamNotificationsRequest) ProtoReflect() protoreflect.Message {
	mi := &file_proto_notification_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StreamNotificationsRequest.ProtoReflect.Descriptor instead.
func (*StreamNotificationsRequest) Descriptor() ([]byte, []int) {
	return file_proto_notification_proto_rawDescGZIP(), []int{1}
}

func (x *StreamNotificationsRequest) GetUserId() string {
	if x != nil {
		return x.UserId
	}
	return ""
}

var File_proto_notification_proto protoreflect.FileDescriptor

const file_proto_notification_proto_rawDesc = "" +
	"\n" +
	"\x18proto/notification.proto\x12\fnotification\"j\n" +
	"\fNotification\x12\x0e\n" +
	"\x02id\x18\x01 \x01(\tR\x02id\x12\x12\n" +
	"\x04type\x18\x02 \x01(\tR\x04type\x12\x18\n" +
	"\amessage\x18\x03 \x01(\tR\amessage\x12\x1c\n" +
	"\ttimestamp\x18\x04 \x01(\tR\ttimestamp\"5\n" +
	"\x1aStreamNotificationsRequest\x12\x17\n" +
	"\auser_id\x18\x01 \x01(\tR\x06userId2v\n" +
	"\x13NotificationService\x12_\n" +
	"\x13StreamNotifications\x12(.notification.StreamNotificationsRequest\x1a\x1a.notification.Notification\"\x000\x01B6Z4dxmultiomics.com/service-notification/notificationpbb\x06proto3"

var (
	file_proto_notification_proto_rawDescOnce sync.Once
	file_proto_notification_proto_rawDescData []byte
)

func file_proto_notification_proto_rawDescGZIP() []byte {
	file_proto_notification_proto_rawDescOnce.Do(func() {
		file_proto_notification_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_proto_notification_proto_rawDesc), len(file_proto_notification_proto_rawDesc)))
	})
	return file_proto_notification_proto_rawDescData
}

var file_proto_notification_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_proto_notification_proto_goTypes = []any{
	(*Notification)(nil),               // 0: notification.Notification
	(*StreamNotificationsRequest)(nil), // 1: notification.StreamNotificationsRequest
}
var file_proto_notification_proto_depIdxs = []int32{
	1, // 0: notification.NotificationService.StreamNotifications:input_type -> notification.StreamNotificationsRequest
	0, // 1: notification.NotificationService.StreamNotifications:output_type -> notification.Notification
	1, // [1:2] is the sub-list for method output_type
	0, // [0:1] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_proto_notification_proto_init() }
func file_proto_notification_proto_init() {
	if File_proto_notification_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_proto_notification_proto_rawDesc), len(file_proto_notification_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_proto_notification_proto_goTypes,
		DependencyIndexes: file_proto_notification_proto_depIdxs,
		MessageInfos:      file_proto_notification_proto_msgTypes,
	}.Build()
	File_proto_notification_proto = out.File
	file_proto_notification_proto_goTypes = nil
	file_proto_notification_proto_depIdxs = nil
}
