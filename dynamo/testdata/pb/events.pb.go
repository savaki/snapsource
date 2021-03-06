// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: events.proto

package pb

import proto "github.com/gogo/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion2 // please upgrade the proto package

type UserCreated struct {
	ID                   string   `protobuf:"bytes,1,opt,name=ID,json=iD,proto3" json:"ID,omitempty"`
	Version              int32    `protobuf:"varint,2,opt,name=Version,json=version,proto3" json:"Version,omitempty"`
	At                   int64    `protobuf:"varint,3,opt,name=At,json=at,proto3" json:"At,omitempty"`
	Name                 string   `protobuf:"bytes,4,opt,name=Name,json=name,proto3" json:"Name,omitempty"`
	Email                string   `protobuf:"bytes,5,opt,name=Email,json=email,proto3" json:"Email,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *UserCreated) Reset()         { *m = UserCreated{} }
func (m *UserCreated) String() string { return proto.CompactTextString(m) }
func (*UserCreated) ProtoMessage()    {}
func (*UserCreated) Descriptor() ([]byte, []int) {
	return fileDescriptor_events_90a1e70da3b6589d, []int{0}
}
func (m *UserCreated) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_UserCreated.Unmarshal(m, b)
}
func (m *UserCreated) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_UserCreated.Marshal(b, m, deterministic)
}
func (dst *UserCreated) XXX_Merge(src proto.Message) {
	xxx_messageInfo_UserCreated.Merge(dst, src)
}
func (m *UserCreated) XXX_Size() int {
	return xxx_messageInfo_UserCreated.Size(m)
}
func (m *UserCreated) XXX_DiscardUnknown() {
	xxx_messageInfo_UserCreated.DiscardUnknown(m)
}

var xxx_messageInfo_UserCreated proto.InternalMessageInfo

func (m *UserCreated) GetID() string {
	if m != nil {
		return m.ID
	}
	return ""
}

func (m *UserCreated) GetVersion() int32 {
	if m != nil {
		return m.Version
	}
	return 0
}

func (m *UserCreated) GetAt() int64 {
	if m != nil {
		return m.At
	}
	return 0
}

func (m *UserCreated) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *UserCreated) GetEmail() string {
	if m != nil {
		return m.Email
	}
	return ""
}

type EmailUpdated struct {
	ID                   string   `protobuf:"bytes,1,opt,name=ID,json=iD,proto3" json:"ID,omitempty"`
	Version              int32    `protobuf:"varint,2,opt,name=Version,json=version,proto3" json:"Version,omitempty"`
	At                   int64    `protobuf:"varint,3,opt,name=At,json=at,proto3" json:"At,omitempty"`
	OldEmail             string   `protobuf:"bytes,4,opt,name=OldEmail,json=oldEmail,proto3" json:"OldEmail,omitempty"`
	NewEmail             string   `protobuf:"bytes,5,opt,name=NewEmail,json=newEmail,proto3" json:"NewEmail,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *EmailUpdated) Reset()         { *m = EmailUpdated{} }
func (m *EmailUpdated) String() string { return proto.CompactTextString(m) }
func (*EmailUpdated) ProtoMessage()    {}
func (*EmailUpdated) Descriptor() ([]byte, []int) {
	return fileDescriptor_events_90a1e70da3b6589d, []int{1}
}
func (m *EmailUpdated) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_EmailUpdated.Unmarshal(m, b)
}
func (m *EmailUpdated) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_EmailUpdated.Marshal(b, m, deterministic)
}
func (dst *EmailUpdated) XXX_Merge(src proto.Message) {
	xxx_messageInfo_EmailUpdated.Merge(dst, src)
}
func (m *EmailUpdated) XXX_Size() int {
	return xxx_messageInfo_EmailUpdated.Size(m)
}
func (m *EmailUpdated) XXX_DiscardUnknown() {
	xxx_messageInfo_EmailUpdated.DiscardUnknown(m)
}

var xxx_messageInfo_EmailUpdated proto.InternalMessageInfo

func (m *EmailUpdated) GetID() string {
	if m != nil {
		return m.ID
	}
	return ""
}

func (m *EmailUpdated) GetVersion() int32 {
	if m != nil {
		return m.Version
	}
	return 0
}

func (m *EmailUpdated) GetAt() int64 {
	if m != nil {
		return m.At
	}
	return 0
}

func (m *EmailUpdated) GetOldEmail() string {
	if m != nil {
		return m.OldEmail
	}
	return ""
}

func (m *EmailUpdated) GetNewEmail() string {
	if m != nil {
		return m.NewEmail
	}
	return ""
}

type Payload struct {
	Type                 int32         `protobuf:"varint,1,opt,name=Type,json=type,proto3" json:"Type,omitempty"`
	T1                   *UserCreated  `protobuf:"bytes,2,opt,name=T1,json=t1" json:"T1,omitempty"`
	T2                   *EmailUpdated `protobuf:"bytes,3,opt,name=T2,json=t2" json:"T2,omitempty"`
	XXX_NoUnkeyedLiteral struct{}      `json:"-"`
	XXX_unrecognized     []byte        `json:"-"`
	XXX_sizecache        int32         `json:"-"`
}

func (m *Payload) Reset()         { *m = Payload{} }
func (m *Payload) String() string { return proto.CompactTextString(m) }
func (*Payload) ProtoMessage()    {}
func (*Payload) Descriptor() ([]byte, []int) {
	return fileDescriptor_events_90a1e70da3b6589d, []int{2}
}
func (m *Payload) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Payload.Unmarshal(m, b)
}
func (m *Payload) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Payload.Marshal(b, m, deterministic)
}
func (dst *Payload) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Payload.Merge(dst, src)
}
func (m *Payload) XXX_Size() int {
	return xxx_messageInfo_Payload.Size(m)
}
func (m *Payload) XXX_DiscardUnknown() {
	xxx_messageInfo_Payload.DiscardUnknown(m)
}

var xxx_messageInfo_Payload proto.InternalMessageInfo

func (m *Payload) GetType() int32 {
	if m != nil {
		return m.Type
	}
	return 0
}

func (m *Payload) GetT1() *UserCreated {
	if m != nil {
		return m.T1
	}
	return nil
}

func (m *Payload) GetT2() *EmailUpdated {
	if m != nil {
		return m.T2
	}
	return nil
}

func init() {
	proto.RegisterType((*UserCreated)(nil), "pb.UserCreated")
	proto.RegisterType((*EmailUpdated)(nil), "pb.EmailUpdated")
	proto.RegisterType((*Payload)(nil), "pb.Payload")
}

func init() { proto.RegisterFile("events.proto", fileDescriptor_events_90a1e70da3b6589d) }

var fileDescriptor_events_90a1e70da3b6589d = []byte{
	// 243 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xa4, 0x90, 0xb1, 0x4e, 0xc3, 0x30,
	0x10, 0x86, 0xe5, 0x6b, 0x4c, 0xc3, 0xa5, 0x02, 0x64, 0x31, 0x58, 0x2c, 0x44, 0x99, 0x32, 0x45,
	0x6a, 0x78, 0x02, 0x44, 0x19, 0x58, 0x0a, 0xb2, 0x52, 0x66, 0x1c, 0xe5, 0x86, 0x48, 0x49, 0x6c,
	0x12, 0xab, 0x28, 0x1b, 0x8f, 0x8e, 0xec, 0x74, 0xc8, 0xce, 0xf8, 0x9d, 0xcf, 0xfa, 0xfe, 0xfb,
	0x71, 0x47, 0x67, 0x1a, 0xdc, 0x54, 0xd8, 0xd1, 0x38, 0x23, 0xc0, 0xd6, 0xd9, 0x37, 0x26, 0xa7,
	0x89, 0xc6, 0x97, 0x91, 0xb4, 0xa3, 0x46, 0xdc, 0x20, 0xbc, 0x1d, 0x24, 0x4b, 0x59, 0x7e, 0xad,
	0xa0, 0x3d, 0x08, 0x89, 0xdb, 0x4f, 0x1a, 0xa7, 0xd6, 0x0c, 0x12, 0x52, 0x96, 0x73, 0xb5, 0x3d,
	0x2f, 0xe8, 0x37, 0x9f, 0x9d, 0xdc, 0xa4, 0x2c, 0xdf, 0x28, 0xd0, 0x4e, 0x08, 0x8c, 0x8e, 0xba,
	0x27, 0x19, 0x85, 0xbf, 0xd1, 0xa0, 0x7b, 0x12, 0xf7, 0xc8, 0x5f, 0x7b, 0xdd, 0x76, 0x92, 0x87,
	0x21, 0x27, 0x0f, 0xd9, 0x2f, 0xc3, 0x5d, 0x18, 0x9f, 0x6c, 0xf3, 0x4f, 0xe9, 0x03, 0xc6, 0xef,
	0x5d, 0xb3, 0x38, 0x16, 0x71, 0x6c, 0x2e, 0xec, 0xdf, 0x8e, 0xf4, 0xb3, 0xf6, 0xc7, 0xc3, 0x85,
	0xb3, 0x2f, 0xdc, 0x7e, 0xe8, 0xb9, 0x33, 0xba, 0xf1, 0xb9, 0xab, 0xd9, 0x52, 0xd0, 0x73, 0x15,
	0xb9, 0xd9, 0x92, 0x78, 0x44, 0xa8, 0xf6, 0xc1, 0x9d, 0x94, 0xb7, 0x85, 0xad, 0x8b, 0x55, 0x45,
	0x0a, 0xdc, 0x5e, 0xa4, 0x08, 0x55, 0x19, 0x72, 0x24, 0xe5, 0x9d, 0x5f, 0x58, 0xdf, 0xa3, 0xc0,
	0x95, 0xf5, 0x55, 0xa8, 0xf8, 0xe9, 0x2f, 0x00, 0x00, 0xff, 0xff, 0x4f, 0x2d, 0xd2, 0xdd, 0x72,
	0x01, 0x00, 0x00,
}
