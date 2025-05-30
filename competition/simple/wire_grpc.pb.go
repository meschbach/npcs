// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v3.17.3
// source: competition/simple/wire.proto

package simple

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
	SimpleGame_Joined_FullMethodName = "/SimpleGame/joined"
)

// SimpleGameClient is the client API for SimpleGame service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type SimpleGameClient interface {
	Joined(ctx context.Context, in *JoinedIn, opts ...grpc.CallOption) (*JoinedOut, error)
}

type simpleGameClient struct {
	cc grpc.ClientConnInterface
}

func NewSimpleGameClient(cc grpc.ClientConnInterface) SimpleGameClient {
	return &simpleGameClient{cc}
}

func (c *simpleGameClient) Joined(ctx context.Context, in *JoinedIn, opts ...grpc.CallOption) (*JoinedOut, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(JoinedOut)
	err := c.cc.Invoke(ctx, SimpleGame_Joined_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// SimpleGameServer is the server API for SimpleGame service.
// All implementations must embed UnimplementedSimpleGameServer
// for forward compatibility.
type SimpleGameServer interface {
	Joined(context.Context, *JoinedIn) (*JoinedOut, error)
	mustEmbedUnimplementedSimpleGameServer()
}

// UnimplementedSimpleGameServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedSimpleGameServer struct{}

func (UnimplementedSimpleGameServer) Joined(context.Context, *JoinedIn) (*JoinedOut, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Joined not implemented")
}
func (UnimplementedSimpleGameServer) mustEmbedUnimplementedSimpleGameServer() {}
func (UnimplementedSimpleGameServer) testEmbeddedByValue()                    {}

// UnsafeSimpleGameServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to SimpleGameServer will
// result in compilation errors.
type UnsafeSimpleGameServer interface {
	mustEmbedUnimplementedSimpleGameServer()
}

func RegisterSimpleGameServer(s grpc.ServiceRegistrar, srv SimpleGameServer) {
	// If the following call pancis, it indicates UnimplementedSimpleGameServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&SimpleGame_ServiceDesc, srv)
}

func _SimpleGame_Joined_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(JoinedIn)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SimpleGameServer).Joined(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: SimpleGame_Joined_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SimpleGameServer).Joined(ctx, req.(*JoinedIn))
	}
	return interceptor(ctx, in, info, handler)
}

// SimpleGame_ServiceDesc is the grpc.ServiceDesc for SimpleGame service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var SimpleGame_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "SimpleGame",
	HandlerType: (*SimpleGameServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "joined",
			Handler:    _SimpleGame_Joined_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "competition/simple/wire.proto",
}
