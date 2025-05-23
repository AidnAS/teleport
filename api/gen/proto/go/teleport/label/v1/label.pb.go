// Copyright 2024 Gravitational, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.6
// 	protoc        (unknown)
// source: teleport/label/v1/label.proto

package labelv1

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

// Label represents a single label key along with a set of possible values for it.
type Label struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// The name of the label.
	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	// The values associated with the label.
	Values        []string `protobuf:"bytes,2,rep,name=values,proto3" json:"values,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Label) Reset() {
	*x = Label{}
	mi := &file_teleport_label_v1_label_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Label) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Label) ProtoMessage() {}

func (x *Label) ProtoReflect() protoreflect.Message {
	mi := &file_teleport_label_v1_label_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Label.ProtoReflect.Descriptor instead.
func (*Label) Descriptor() ([]byte, []int) {
	return file_teleport_label_v1_label_proto_rawDescGZIP(), []int{0}
}

func (x *Label) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Label) GetValues() []string {
	if x != nil {
		return x.Values
	}
	return nil
}

var File_teleport_label_v1_label_proto protoreflect.FileDescriptor

const file_teleport_label_v1_label_proto_rawDesc = "" +
	"\n" +
	"\x1dteleport/label/v1/label.proto\x12\x11teleport.label.v1\"3\n" +
	"\x05Label\x12\x12\n" +
	"\x04name\x18\x01 \x01(\tR\x04name\x12\x16\n" +
	"\x06values\x18\x02 \x03(\tR\x06valuesBNZLgithub.com/gravitational/teleport/api/gen/proto/go/teleport/label/v1;labelv1b\x06proto3"

var (
	file_teleport_label_v1_label_proto_rawDescOnce sync.Once
	file_teleport_label_v1_label_proto_rawDescData []byte
)

func file_teleport_label_v1_label_proto_rawDescGZIP() []byte {
	file_teleport_label_v1_label_proto_rawDescOnce.Do(func() {
		file_teleport_label_v1_label_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_teleport_label_v1_label_proto_rawDesc), len(file_teleport_label_v1_label_proto_rawDesc)))
	})
	return file_teleport_label_v1_label_proto_rawDescData
}

var file_teleport_label_v1_label_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_teleport_label_v1_label_proto_goTypes = []any{
	(*Label)(nil), // 0: teleport.label.v1.Label
}
var file_teleport_label_v1_label_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_teleport_label_v1_label_proto_init() }
func file_teleport_label_v1_label_proto_init() {
	if File_teleport_label_v1_label_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_teleport_label_v1_label_proto_rawDesc), len(file_teleport_label_v1_label_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_teleport_label_v1_label_proto_goTypes,
		DependencyIndexes: file_teleport_label_v1_label_proto_depIdxs,
		MessageInfos:      file_teleport_label_v1_label_proto_msgTypes,
	}.Build()
	File_teleport_label_v1_label_proto = out.File
	file_teleport_label_v1_label_proto_goTypes = nil
	file_teleport_label_v1_label_proto_depIdxs = nil
}
