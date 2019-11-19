// Code generated by protoc-gen-micro. DO NOT EDIT.
// source: vps.proto

package vps

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	math "math"
)

import (
	context "context"
	client "github.com/micro/go-micro/client"
	server "github.com/micro/go-micro/server"
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

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ client.Option
var _ server.Option

// Client API for Vps service

type VpsService interface {
	NewNode(ctx context.Context, in *Request, opts ...client.CallOption) (*Response, error)
	DelNode(ctx context.Context, in *Request, opts ...client.CallOption) (*Response, error)
	ExpandVolume(ctx context.Context, in *VolumeRequest, opts ...client.CallOption) (*Response, error)
}

type vpsService struct {
	c    client.Client
	name string
}

func NewVpsService(name string, c client.Client) VpsService {
	if c == nil {
		c = client.NewClient()
	}
	if len(name) == 0 {
		name = "vps"
	}
	return &vpsService{
		c:    c,
		name: name,
	}
}

func (c *vpsService) NewNode(ctx context.Context, in *Request, opts ...client.CallOption) (*Response, error) {
	req := c.c.NewRequest(c.name, "Vps.NewNode", in)
	out := new(Response)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *vpsService) DelNode(ctx context.Context, in *Request, opts ...client.CallOption) (*Response, error) {
	req := c.c.NewRequest(c.name, "Vps.DelNode", in)
	out := new(Response)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *vpsService) ExpandVolume(ctx context.Context, in *VolumeRequest, opts ...client.CallOption) (*Response, error) {
	req := c.c.NewRequest(c.name, "Vps.ExpandVolume", in)
	out := new(Response)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for Vps service

type VpsHandler interface {
	NewNode(context.Context, *Request, *Response) error
	DelNode(context.Context, *Request, *Response) error
	ExpandVolume(context.Context, *VolumeRequest, *Response) error
}

func RegisterVpsHandler(s server.Server, hdlr VpsHandler, opts ...server.HandlerOption) error {
	type vps interface {
		NewNode(ctx context.Context, in *Request, out *Response) error
		DelNode(ctx context.Context, in *Request, out *Response) error
		ExpandVolume(ctx context.Context, in *VolumeRequest, out *Response) error
	}
	type Vps struct {
		vps
	}
	h := &vpsHandler{hdlr}
	return s.Handle(s.NewHandler(&Vps{h}, opts...))
}

type vpsHandler struct {
	VpsHandler
}

func (h *vpsHandler) NewNode(ctx context.Context, in *Request, out *Response) error {
	return h.VpsHandler.NewNode(ctx, in, out)
}

func (h *vpsHandler) DelNode(ctx context.Context, in *Request, out *Response) error {
	return h.VpsHandler.DelNode(ctx, in, out)
}

func (h *vpsHandler) ExpandVolume(ctx context.Context, in *VolumeRequest, out *Response) error {
	return h.VpsHandler.ExpandVolume(ctx, in, out)
}
