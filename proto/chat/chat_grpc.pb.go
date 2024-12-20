// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v5.28.3
// source: proto/chat/chat.proto

package chat

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	ChannelService_CreateChannel_FullMethodName = "/chat.ChannelService/CreateChannel"
)

// ChannelServiceClient is the client API for ChannelService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ChannelServiceClient interface {
	CreateChannel(ctx context.Context, in *CreateChannelRequest, opts ...grpc.CallOption) (*CreateChannelResponse, error)
}

type channelServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewChannelServiceClient(cc grpc.ClientConnInterface) ChannelServiceClient {
	return &channelServiceClient{cc}
}

func (c *channelServiceClient) CreateChannel(ctx context.Context, in *CreateChannelRequest, opts ...grpc.CallOption) (*CreateChannelResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(CreateChannelResponse)
	err := c.cc.Invoke(ctx, ChannelService_CreateChannel_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ChannelServiceServer is the server API for ChannelService service.
// All implementations must embed UnimplementedChannelServiceServer
// for forward compatibility.
type ChannelServiceServer interface {
	CreateChannel(context.Context, *CreateChannelRequest) (*CreateChannelResponse, error)
	mustEmbedUnimplementedChannelServiceServer()
}

// UnimplementedChannelServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedChannelServiceServer struct{}

func (UnimplementedChannelServiceServer) CreateChannel(context.Context, *CreateChannelRequest) (*CreateChannelResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateChannel not implemented")
}
func (UnimplementedChannelServiceServer) mustEmbedUnimplementedChannelServiceServer() {}
func (UnimplementedChannelServiceServer) testEmbeddedByValue()                        {}

// UnsafeChannelServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ChannelServiceServer will
// result in compilation errors.
type UnsafeChannelServiceServer interface {
	mustEmbedUnimplementedChannelServiceServer()
}

func RegisterChannelServiceServer(s grpc.ServiceRegistrar, srv ChannelServiceServer) {
	// If the following call pancis, it indicates UnimplementedChannelServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&ChannelService_ServiceDesc, srv)
}

func _ChannelService_CreateChannel_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateChannelRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ChannelServiceServer).CreateChannel(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ChannelService_CreateChannel_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ChannelServiceServer).CreateChannel(ctx, req.(*CreateChannelRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// ChannelService_ServiceDesc is the grpc.ServiceDesc for ChannelService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var ChannelService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "chat.ChannelService",
	HandlerType: (*ChannelServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CreateChannel",
			Handler:    _ChannelService_CreateChannel_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "proto/chat/chat.proto",
}
