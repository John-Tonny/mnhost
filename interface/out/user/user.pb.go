// Code generated by protoc-gen-go. DO NOT EDIT.
// source: user.proto

package user

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

// 用户信息
type User struct {
	Id                   string   `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Password             string   `protobuf:"bytes,3,opt,name=password,proto3" json:"password,omitempty"`
	Mobile               string   `protobuf:"bytes,4,opt,name=mobile,proto3" json:"mobile,omitempty"`
	Name                 string   `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	RealName             string   `protobuf:"bytes,5,opt,name=realName,proto3" json:"realName,omitempty"`
	IdCard               string   `protobuf:"bytes,6,opt,name=idCard,proto3" json:"idCard,omitempty"`
	RewardAddress        string   `protobuf:"bytes,7,opt,name=rewardAddress,proto3" json:"rewardAddress,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *User) Reset()         { *m = User{} }
func (m *User) String() string { return proto.CompactTextString(m) }
func (*User) ProtoMessage()    {}
func (*User) Descriptor() ([]byte, []int) {
	return fileDescriptor_116e343673f7ffaf, []int{0}
}

func (m *User) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_User.Unmarshal(m, b)
}
func (m *User) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_User.Marshal(b, m, deterministic)
}
func (m *User) XXX_Merge(src proto.Message) {
	xxx_messageInfo_User.Merge(m, src)
}
func (m *User) XXX_Size() int {
	return xxx_messageInfo_User.Size(m)
}
func (m *User) XXX_DiscardUnknown() {
	xxx_messageInfo_User.DiscardUnknown(m)
}

var xxx_messageInfo_User proto.InternalMessageInfo

func (m *User) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *User) GetPassword() string {
	if m != nil {
		return m.Password
	}
	return ""
}

func (m *User) GetMobile() string {
	if m != nil {
		return m.Mobile
	}
	return ""
}

func (m *User) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *User) GetRealName() string {
	if m != nil {
		return m.RealName
	}
	return ""
}

func (m *User) GetIdCard() string {
	if m != nil {
		return m.IdCard
	}
	return ""
}

func (m *User) GetRewardAddress() string {
	if m != nil {
		return m.RewardAddress
	}
	return ""
}

type Request struct {
	Mobile               string   `protobuf:"bytes,1,opt,name=mobile,proto3" json:"mobile,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Request) Reset()         { *m = Request{} }
func (m *Request) String() string { return proto.CompactTextString(m) }
func (*Request) ProtoMessage()    {}
func (*Request) Descriptor() ([]byte, []int) {
	return fileDescriptor_116e343673f7ffaf, []int{1}
}

func (m *Request) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Request.Unmarshal(m, b)
}
func (m *Request) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Request.Marshal(b, m, deterministic)
}
func (m *Request) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Request.Merge(m, src)
}
func (m *Request) XXX_Size() int {
	return xxx_messageInfo_Request.Size(m)
}
func (m *Request) XXX_DiscardUnknown() {
	xxx_messageInfo_Request.DiscardUnknown(m)
}

var xxx_messageInfo_Request proto.InternalMessageInfo

func (m *Request) GetMobile() string {
	if m != nil {
		return m.Mobile
	}
	return ""
}

type Response struct {
	User                 *User    `protobuf:"bytes,1,opt,name=user,proto3" json:"user,omitempty"`
	Users                []*User  `protobuf:"bytes,2,rep,name=users,proto3" json:"users,omitempty"`
	Errors               []*Error `protobuf:"bytes,3,rep,name=errors,proto3" json:"errors,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Response) Reset()         { *m = Response{} }
func (m *Response) String() string { return proto.CompactTextString(m) }
func (*Response) ProtoMessage()    {}
func (*Response) Descriptor() ([]byte, []int) {
	return fileDescriptor_116e343673f7ffaf, []int{2}
}

func (m *Response) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Response.Unmarshal(m, b)
}
func (m *Response) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Response.Marshal(b, m, deterministic)
}
func (m *Response) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Response.Merge(m, src)
}
func (m *Response) XXX_Size() int {
	return xxx_messageInfo_Response.Size(m)
}
func (m *Response) XXX_DiscardUnknown() {
	xxx_messageInfo_Response.DiscardUnknown(m)
}

var xxx_messageInfo_Response proto.InternalMessageInfo

func (m *Response) GetUser() *User {
	if m != nil {
		return m.User
	}
	return nil
}

func (m *Response) GetUsers() []*User {
	if m != nil {
		return m.Users
	}
	return nil
}

func (m *Response) GetErrors() []*Error {
	if m != nil {
		return m.Errors
	}
	return nil
}

type Token struct {
	Token                string   `protobuf:"bytes,1,opt,name=token,proto3" json:"token,omitempty"`
	Valid                bool     `protobuf:"varint,2,opt,name=valid,proto3" json:"valid,omitempty"`
	Errors               *Error   `protobuf:"bytes,3,opt,name=errors,proto3" json:"errors,omitempty"`
	UserId               string   `protobuf:"bytes,4,opt,name=user_id,json=userId,proto3" json:"user_id,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Token) Reset()         { *m = Token{} }
func (m *Token) String() string { return proto.CompactTextString(m) }
func (*Token) ProtoMessage()    {}
func (*Token) Descriptor() ([]byte, []int) {
	return fileDescriptor_116e343673f7ffaf, []int{3}
}

func (m *Token) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Token.Unmarshal(m, b)
}
func (m *Token) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Token.Marshal(b, m, deterministic)
}
func (m *Token) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Token.Merge(m, src)
}
func (m *Token) XXX_Size() int {
	return xxx_messageInfo_Token.Size(m)
}
func (m *Token) XXX_DiscardUnknown() {
	xxx_messageInfo_Token.DiscardUnknown(m)
}

var xxx_messageInfo_Token proto.InternalMessageInfo

func (m *Token) GetToken() string {
	if m != nil {
		return m.Token
	}
	return ""
}

func (m *Token) GetValid() bool {
	if m != nil {
		return m.Valid
	}
	return false
}

func (m *Token) GetErrors() *Error {
	if m != nil {
		return m.Errors
	}
	return nil
}

func (m *Token) GetUserId() string {
	if m != nil {
		return m.UserId
	}
	return ""
}

type Error struct {
	Code                 int32    `protobuf:"varint,1,opt,name=code,proto3" json:"code,omitempty"`
	Description          string   `protobuf:"bytes,2,opt,name=description,proto3" json:"description,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Error) Reset()         { *m = Error{} }
func (m *Error) String() string { return proto.CompactTextString(m) }
func (*Error) ProtoMessage()    {}
func (*Error) Descriptor() ([]byte, []int) {
	return fileDescriptor_116e343673f7ffaf, []int{4}
}

func (m *Error) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Error.Unmarshal(m, b)
}
func (m *Error) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Error.Marshal(b, m, deterministic)
}
func (m *Error) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Error.Merge(m, src)
}
func (m *Error) XXX_Size() int {
	return xxx_messageInfo_Error.Size(m)
}
func (m *Error) XXX_DiscardUnknown() {
	xxx_messageInfo_Error.DiscardUnknown(m)
}

var xxx_messageInfo_Error proto.InternalMessageInfo

func (m *Error) GetCode() int32 {
	if m != nil {
		return m.Code
	}
	return 0
}

func (m *Error) GetDescription() string {
	if m != nil {
		return m.Description
	}
	return ""
}

func init() {
	proto.RegisterType((*User)(nil), "user.User")
	proto.RegisterType((*Request)(nil), "user.Request")
	proto.RegisterType((*Response)(nil), "user.Response")
	proto.RegisterType((*Token)(nil), "user.Token")
	proto.RegisterType((*Error)(nil), "user.Error")
}

func init() { proto.RegisterFile("user.proto", fileDescriptor_116e343673f7ffaf) }

var fileDescriptor_116e343673f7ffaf = []byte{
	// 395 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x84, 0x52, 0xcd, 0x8a, 0xd4, 0x40,
	0x10, 0xde, 0xfc, 0xce, 0x58, 0x61, 0xf6, 0x50, 0x88, 0x36, 0x7b, 0x90, 0x6c, 0x56, 0x44, 0x11,
	0xf6, 0xb0, 0x9e, 0x3d, 0x0c, 0x8b, 0x2c, 0x5e, 0x3c, 0xc4, 0x9f, 0xab, 0xf4, 0x4e, 0x17, 0xd8,
	0x98, 0x4d, 0x67, 0xba, 0x7b, 0x66, 0x5e, 0xcd, 0x17, 0xf1, 0x7d, 0xa4, 0xab, 0x13, 0xc9, 0x0c,
	0xca, 0x9e, 0x52, 0xdf, 0x0f, 0x1f, 0x55, 0x5f, 0x07, 0x60, 0xe7, 0xc8, 0x5e, 0x0f, 0xd6, 0x78,
	0x83, 0x79, 0x98, 0x9b, 0x5f, 0x09, 0xe4, 0x5f, 0x1d, 0x59, 0x3c, 0x87, 0x54, 0x2b, 0x91, 0xd4,
	0xc9, 0xeb, 0x27, 0x6d, 0xaa, 0x15, 0x5e, 0xc0, 0x72, 0x90, 0xce, 0x1d, 0x8c, 0x55, 0x22, 0x63,
	0xf6, 0x2f, 0xc6, 0x67, 0x50, 0x3e, 0x98, 0x7b, 0xdd, 0x91, 0xc8, 0x59, 0x19, 0x11, 0x22, 0xe4,
	0xbd, 0x7c, 0x20, 0x91, 0x32, 0xcb, 0x73, 0xc8, 0xb1, 0x24, 0xbb, 0x4f, 0x81, 0x2f, 0x62, 0xce,
	0x84, 0x43, 0x8e, 0x56, 0xb7, 0xd2, 0x2a, 0x51, 0xc6, 0x9c, 0x88, 0xf0, 0x25, 0xac, 0x2c, 0x1d,
	0xa4, 0x55, 0x6b, 0xa5, 0x2c, 0x39, 0x27, 0x16, 0x2c, 0x1f, 0x93, 0xcd, 0x25, 0x2c, 0x5a, 0xda,
	0xee, 0xc8, 0xf9, 0xd9, 0x42, 0xc9, 0x7c, 0xa1, 0x66, 0x0b, 0xcb, 0x96, 0xdc, 0x60, 0x7a, 0x47,
	0xf8, 0x02, 0xf8, 0x62, 0x76, 0x54, 0x37, 0x70, 0xcd, 0x55, 0x84, 0xd3, 0x5b, 0xe6, 0xb1, 0x86,
	0x22, 0x7c, 0x9d, 0x48, 0xeb, 0xec, 0xc4, 0x10, 0x05, 0xbc, 0x82, 0x92, 0xac, 0x35, 0xd6, 0x89,
	0x8c, 0x2d, 0x55, 0xb4, 0x7c, 0x08, 0x5c, 0x3b, 0x4a, 0xcd, 0x16, 0x8a, 0x2f, 0xe6, 0x27, 0xf5,
	0xf8, 0x14, 0x0a, 0x1f, 0x86, 0x71, 0xa5, 0x08, 0x02, 0xbb, 0x97, 0x9d, 0x56, 0xdc, 0xd1, 0xb2,
	0x8d, 0xe0, 0x28, 0x39, 0xf9, 0x4f, 0x32, 0x3e, 0x87, 0x45, 0x60, 0xbf, 0x6b, 0x35, 0xd5, 0x1e,
	0xe0, 0x47, 0xd5, 0xbc, 0x87, 0x82, 0x9d, 0xa1, 0xff, 0x8d, 0x51, 0xb1, 0x84, 0xa2, 0xe5, 0x19,
	0x6b, 0xa8, 0x14, 0xb9, 0x8d, 0xd5, 0x83, 0xd7, 0xa6, 0x1f, 0x9f, 0x66, 0x4e, 0xdd, 0xfc, 0x4e,
	0xa0, 0x0a, 0x67, 0x7e, 0x26, 0xbb, 0xd7, 0x1b, 0xc2, 0x57, 0x50, 0xde, 0x5a, 0x92, 0x9e, 0x70,
	0xd6, 0xc1, 0xc5, 0x79, 0x9c, 0xa7, 0x3a, 0x9b, 0x33, 0xbc, 0x82, 0xec, 0x8e, 0xfc, 0x23, 0xa6,
	0x37, 0x50, 0xde, 0x91, 0x5f, 0x77, 0x1d, 0xae, 0x26, 0x8d, 0x9f, 0xec, 0x1f, 0xd6, 0x4b, 0xc8,
	0xd7, 0x3b, 0xff, 0xe3, 0x28, 0x70, 0x2c, 0x82, 0x1b, 0x6d, 0xce, 0xf0, 0x2d, 0xac, 0xbe, 0x85,
	0xc2, 0xa4, 0xa7, 0x58, 0xf2, 0x5c, 0x3f, 0x31, 0xdf, 0x97, 0xfc, 0x9f, 0xbf, 0xfb, 0x13, 0x00,
	0x00, 0xff, 0xff, 0xd1, 0xef, 0x5d, 0x1f, 0xf5, 0x02, 0x00, 0x00,
}
