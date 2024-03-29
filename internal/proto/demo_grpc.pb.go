// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v4.23.0
// source: proto/demo.proto

package proto

import (
	context "context"

	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

const (
	Metrics_UpdateMetric_FullMethodName      = "/grpc_server.Metrics/UpdateMetric"
	Metrics_UpdateManyMetrics_FullMethodName = "/grpc_server.Metrics/UpdateManyMetrics"
	Metrics_GetMetric_FullMethodName         = "/grpc_server.Metrics/GetMetric"
	Metrics_GetAllMetrics_FullMethodName     = "/grpc_server.Metrics/GetAllMetrics"
	Metrics_PingDatabase_FullMethodName      = "/grpc_server.Metrics/PingDatabase"
)

// MetricsClient is the client API for Metrics service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type MetricsClient interface {
	UpdateMetric(ctx context.Context, in *UpdateMetricRequest, opts ...grpc.CallOption) (*UpdateMetricResponse, error)
	UpdateManyMetrics(ctx context.Context, in *UpdateManyMetricsRequest, opts ...grpc.CallOption) (*UpdateManyMetricsResponse, error)
	GetMetric(ctx context.Context, in *GetMetricRequest, opts ...grpc.CallOption) (*GetMetricResponse, error)
	GetAllMetrics(ctx context.Context, in *GetAllMetricsRequest, opts ...grpc.CallOption) (*GetAllMetricsResponse, error)
	PingDatabase(ctx context.Context, in *PingDatabaseRequest, opts ...grpc.CallOption) (*PingDatabaseResponse, error)
}

type metricsClient struct {
	cc grpc.ClientConnInterface
}

func NewMetricsClient(cc grpc.ClientConnInterface) MetricsClient {
	return &metricsClient{cc}
}

func (c *metricsClient) UpdateMetric(ctx context.Context, in *UpdateMetricRequest, opts ...grpc.CallOption) (*UpdateMetricResponse, error) {
	out := new(UpdateMetricResponse)
	err := c.cc.Invoke(ctx, Metrics_UpdateMetric_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *metricsClient) UpdateManyMetrics(ctx context.Context, in *UpdateManyMetricsRequest, opts ...grpc.CallOption) (*UpdateManyMetricsResponse, error) {
	out := new(UpdateManyMetricsResponse)
	err := c.cc.Invoke(ctx, Metrics_UpdateManyMetrics_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *metricsClient) GetMetric(ctx context.Context, in *GetMetricRequest, opts ...grpc.CallOption) (*GetMetricResponse, error) {
	out := new(GetMetricResponse)
	err := c.cc.Invoke(ctx, Metrics_GetMetric_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *metricsClient) GetAllMetrics(ctx context.Context, in *GetAllMetricsRequest, opts ...grpc.CallOption) (*GetAllMetricsResponse, error) {
	out := new(GetAllMetricsResponse)
	err := c.cc.Invoke(ctx, Metrics_GetAllMetrics_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *metricsClient) PingDatabase(ctx context.Context, in *PingDatabaseRequest, opts ...grpc.CallOption) (*PingDatabaseResponse, error) {
	out := new(PingDatabaseResponse)
	err := c.cc.Invoke(ctx, Metrics_PingDatabase_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MetricsServer is the server API for Metrics service.
// All implementations must embed UnimplementedMetricsServer
// for forward compatibility
type MetricsServer interface {
	UpdateMetric(context.Context, *UpdateMetricRequest) (*UpdateMetricResponse, error)
	UpdateManyMetrics(context.Context, *UpdateManyMetricsRequest) (*UpdateManyMetricsResponse, error)
	GetMetric(context.Context, *GetMetricRequest) (*GetMetricResponse, error)
	GetAllMetrics(context.Context, *GetAllMetricsRequest) (*GetAllMetricsResponse, error)
	PingDatabase(context.Context, *PingDatabaseRequest) (*PingDatabaseResponse, error)
	mustEmbedUnimplementedMetricsServer()
}

// UnimplementedMetricsServer must be embedded to have forward compatible implementations.
type UnimplementedMetricsServer struct {
}

func (UnimplementedMetricsServer) UpdateMetric(context.Context, *UpdateMetricRequest) (*UpdateMetricResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateMetric not implemented")
}
func (UnimplementedMetricsServer) UpdateManyMetrics(context.Context, *UpdateManyMetricsRequest) (*UpdateManyMetricsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateManyMetrics not implemented")
}
func (UnimplementedMetricsServer) GetMetric(context.Context, *GetMetricRequest) (*GetMetricResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetMetric not implemented")
}
func (UnimplementedMetricsServer) GetAllMetrics(context.Context, *GetAllMetricsRequest) (*GetAllMetricsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetAllMetrics not implemented")
}
func (UnimplementedMetricsServer) PingDatabase(context.Context, *PingDatabaseRequest) (*PingDatabaseResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PingDatabase not implemented")
}
func (UnimplementedMetricsServer) mustEmbedUnimplementedMetricsServer() {}

// UnsafeMetricsServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to MetricsServer will
// result in compilation errors.
type UnsafeMetricsServer interface {
	mustEmbedUnimplementedMetricsServer()
}

func RegisterMetricsServer(s grpc.ServiceRegistrar, srv MetricsServer) {
	s.RegisterService(&Metrics_ServiceDesc, srv)
}

func _Metrics_UpdateMetric_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateMetricRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MetricsServer).UpdateMetric(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Metrics_UpdateMetric_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MetricsServer).UpdateMetric(ctx, req.(*UpdateMetricRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Metrics_UpdateManyMetrics_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateManyMetricsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MetricsServer).UpdateManyMetrics(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Metrics_UpdateManyMetrics_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MetricsServer).UpdateManyMetrics(ctx, req.(*UpdateManyMetricsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Metrics_GetMetric_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetMetricRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MetricsServer).GetMetric(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Metrics_GetMetric_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MetricsServer).GetMetric(ctx, req.(*GetMetricRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Metrics_GetAllMetrics_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetAllMetricsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MetricsServer).GetAllMetrics(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Metrics_GetAllMetrics_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MetricsServer).GetAllMetrics(ctx, req.(*GetAllMetricsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Metrics_PingDatabase_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PingDatabaseRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MetricsServer).PingDatabase(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Metrics_PingDatabase_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MetricsServer).PingDatabase(ctx, req.(*PingDatabaseRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Metrics_ServiceDesc is the grpc.ServiceDesc for Metrics service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Metrics_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "grpc_server.Metrics",
	HandlerType: (*MetricsServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "UpdateMetric",
			Handler:    _Metrics_UpdateMetric_Handler,
		},
		{
			MethodName: "UpdateManyMetrics",
			Handler:    _Metrics_UpdateManyMetrics_Handler,
		},
		{
			MethodName: "GetMetric",
			Handler:    _Metrics_GetMetric_Handler,
		},
		{
			MethodName: "GetAllMetrics",
			Handler:    _Metrics_GetAllMetrics_Handler,
		},
		{
			MethodName: "PingDatabase",
			Handler:    _Metrics_PingDatabase_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "proto/demo.proto",
}
