// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: pkg/logproto/indexgateway.proto

package logproto

import (
	context "context"
	fmt "fmt"
	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/gogo/protobuf/proto"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
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
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

func init() { proto.RegisterFile("pkg/logproto/indexgateway.proto", fileDescriptor_d27585148d0a52c8) }

var fileDescriptor_d27585148d0a52c8 = []byte{
	// 351 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x92, 0xc1, 0x4a, 0xfb, 0x30,
	0x1c, 0xc7, 0x13, 0xfe, 0xf0, 0x47, 0xa3, 0x78, 0x08, 0xc2, 0xa4, 0xd3, 0x08, 0xe2, 0x41, 0x2f,
	0xab, 0xe8, 0x1b, 0x28, 0xac, 0x0c, 0xa6, 0xe0, 0x04, 0x0f, 0x3b, 0x88, 0xe9, 0xfc, 0x2d, 0x2b,
	0xeb, 0x9a, 0xda, 0xa6, 0xe8, 0x6e, 0x3e, 0x82, 0x8f, 0xe1, 0xa3, 0x78, 0xdc, 0x71, 0x47, 0x97,
	0x5d, 0x3c, 0xee, 0xec, 0x49, 0xda, 0xd2, 0x2d, 0x1b, 0x1d, 0x78, 0x6a, 0xfa, 0xfd, 0x7e, 0xf2,
	0xf9, 0xd1, 0x26, 0xe4, 0x30, 0xec, 0x0b, 0xdb, 0x97, 0x22, 0x8c, 0xa4, 0x92, 0xb6, 0x17, 0x3c,
	0xc1, 0xab, 0xe0, 0x0a, 0x5e, 0xf8, 0xb0, 0x96, 0x45, 0x74, 0xc7, 0xcc, 0x42, 0xd7, 0xda, 0x15,
	0x52, 0xc8, 0x9c, 0x4e, 0x57, 0x39, 0x65, 0x55, 0x97, 0x34, 0xc5, 0x22, 0x2f, 0xcf, 0x7f, 0xfe,
	0x91, 0xed, 0x46, 0x6a, 0x71, 0x72, 0x0b, 0x6d, 0x10, 0x72, 0x9b, 0x40, 0x34, 0xcc, 0x42, 0x5a,
	0xad, 0xcd, 0xf9, 0x45, 0xda, 0x82, 0xe7, 0x04, 0x62, 0x65, 0xed, 0x97, 0x97, 0x71, 0x28, 0x83,
	0x18, 0xce, 0x30, 0x6d, 0x92, 0x2d, 0x07, 0xd4, 0x55, 0x2f, 0x09, 0xfa, 0x2d, 0xe8, 0x52, 0x03,
	0x37, 0xe2, 0x42, 0x76, 0xb0, 0xa6, 0xcd, 0x6d, 0x47, 0x88, 0xd6, 0xc9, 0xa6, 0x03, 0xea, 0x0e,
	0x22, 0x0f, 0x62, 0x6a, 0x2d, 0xd1, 0x79, 0x58, 0x98, 0xaa, 0xa5, 0xdd, 0xdc, 0xf3, 0x40, 0x2a,
	0x4d, 0xee, 0x82, 0x7f, 0xc3, 0x07, 0x10, 0xd7, 0x65, 0x74, 0x0d, 0x2a, 0xf2, 0x3a, 0xe9, 0x1b,
	0x3d, 0x59, 0xec, 0x5c, 0x83, 0x14, 0x33, 0x2a, 0x2b, 0xa4, 0xe1, 0x7f, 0x24, 0x7b, 0x59, 0x74,
	0xcf, 0xfd, 0x64, 0x75, 0xc0, 0xe9, 0xca, 0xb6, 0x12, 0xe6, 0x0f, 0x13, 0x1c, 0xb2, 0x91, 0x7e,
	0x98, 0xe2, 0x2a, 0x36, 0x0f, 0x28, 0xfb, 0xfd, 0x59, 0x5a, 0x72, 0x40, 0x66, 0x59, 0x88, 0x2e,
	0xdb, 0xa3, 0x09, 0x43, 0xe3, 0x09, 0x43, 0xb3, 0x09, 0xc3, 0x6f, 0x9a, 0xe1, 0x0f, 0xcd, 0xf0,
	0xa7, 0x66, 0x78, 0xa4, 0x19, 0xfe, 0xd2, 0x0c, 0x7f, 0x6b, 0x86, 0x66, 0x9a, 0xe1, 0xf7, 0x29,
	0x43, 0xa3, 0x29, 0x43, 0xe3, 0x29, 0x43, 0xed, 0x63, 0xe1, 0xa9, 0x5e, 0xe2, 0xd6, 0x3a, 0x72,
	0x60, 0x8b, 0x88, 0x77, 0x79, 0xc0, 0x6d, 0x5f, 0xf6, 0x3d, 0xdb, 0xbc, 0x67, 0xee, 0xff, 0xec,
	0x71, 0xf1, 0x1b, 0x00, 0x00, 0xff, 0xff, 0xbb, 0x22, 0x04, 0xe7, 0xc5, 0x02, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// IndexGatewayClient is the client API for IndexGateway service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type IndexGatewayClient interface {
	/// QueryIndex reads the indexes required for given query & sends back the batch of rows
	/// in rpc streams
	QueryIndex(ctx context.Context, in *QueryIndexRequest, opts ...grpc.CallOption) (IndexGateway_QueryIndexClient, error)
	/// GetChunkRef returns chunk reference that match the provided label matchers
	GetChunkRef(ctx context.Context, in *GetChunkRefRequest, opts ...grpc.CallOption) (*GetChunkRefResponse, error)
	GetSeries(ctx context.Context, in *GetSeriesRequest, opts ...grpc.CallOption) (*GetSeriesResponse, error)
	LabelNamesForMetricName(ctx context.Context, in *LabelNamesForMetricNameRequest, opts ...grpc.CallOption) (*LabelResponse, error)
	LabelValuesForMetricName(ctx context.Context, in *LabelValuesForMetricNameRequest, opts ...grpc.CallOption) (*LabelResponse, error)
	// Note: this MUST be the same as the variant defined in
	// logproto.proto on the Querier service.
	GetStats(ctx context.Context, in *IndexStatsRequest, opts ...grpc.CallOption) (*IndexStatsResponse, error)
}

type indexGatewayClient struct {
	cc *grpc.ClientConn
}

func NewIndexGatewayClient(cc *grpc.ClientConn) IndexGatewayClient {
	return &indexGatewayClient{cc}
}

func (c *indexGatewayClient) QueryIndex(ctx context.Context, in *QueryIndexRequest, opts ...grpc.CallOption) (IndexGateway_QueryIndexClient, error) {
	stream, err := c.cc.NewStream(ctx, &_IndexGateway_serviceDesc.Streams[0], "/indexgatewaypb.IndexGateway/QueryIndex", opts...)
	if err != nil {
		return nil, err
	}
	x := &indexGatewayQueryIndexClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type IndexGateway_QueryIndexClient interface {
	Recv() (*QueryIndexResponse, error)
	grpc.ClientStream
}

type indexGatewayQueryIndexClient struct {
	grpc.ClientStream
}

func (x *indexGatewayQueryIndexClient) Recv() (*QueryIndexResponse, error) {
	m := new(QueryIndexResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *indexGatewayClient) GetChunkRef(ctx context.Context, in *GetChunkRefRequest, opts ...grpc.CallOption) (*GetChunkRefResponse, error) {
	out := new(GetChunkRefResponse)
	err := c.cc.Invoke(ctx, "/indexgatewaypb.IndexGateway/GetChunkRef", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *indexGatewayClient) GetSeries(ctx context.Context, in *GetSeriesRequest, opts ...grpc.CallOption) (*GetSeriesResponse, error) {
	out := new(GetSeriesResponse)
	err := c.cc.Invoke(ctx, "/indexgatewaypb.IndexGateway/GetSeries", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *indexGatewayClient) LabelNamesForMetricName(ctx context.Context, in *LabelNamesForMetricNameRequest, opts ...grpc.CallOption) (*LabelResponse, error) {
	out := new(LabelResponse)
	err := c.cc.Invoke(ctx, "/indexgatewaypb.IndexGateway/LabelNamesForMetricName", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *indexGatewayClient) LabelValuesForMetricName(ctx context.Context, in *LabelValuesForMetricNameRequest, opts ...grpc.CallOption) (*LabelResponse, error) {
	out := new(LabelResponse)
	err := c.cc.Invoke(ctx, "/indexgatewaypb.IndexGateway/LabelValuesForMetricName", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *indexGatewayClient) GetStats(ctx context.Context, in *IndexStatsRequest, opts ...grpc.CallOption) (*IndexStatsResponse, error) {
	out := new(IndexStatsResponse)
	err := c.cc.Invoke(ctx, "/indexgatewaypb.IndexGateway/GetStats", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// IndexGatewayServer is the server API for IndexGateway service.
type IndexGatewayServer interface {
	/// QueryIndex reads the indexes required for given query & sends back the batch of rows
	/// in rpc streams
	QueryIndex(*QueryIndexRequest, IndexGateway_QueryIndexServer) error
	/// GetChunkRef returns chunk reference that match the provided label matchers
	GetChunkRef(context.Context, *GetChunkRefRequest) (*GetChunkRefResponse, error)
	GetSeries(context.Context, *GetSeriesRequest) (*GetSeriesResponse, error)
	LabelNamesForMetricName(context.Context, *LabelNamesForMetricNameRequest) (*LabelResponse, error)
	LabelValuesForMetricName(context.Context, *LabelValuesForMetricNameRequest) (*LabelResponse, error)
	// Note: this MUST be the same as the variant defined in
	// logproto.proto on the Querier service.
	GetStats(context.Context, *IndexStatsRequest) (*IndexStatsResponse, error)
}

// UnimplementedIndexGatewayServer can be embedded to have forward compatible implementations.
type UnimplementedIndexGatewayServer struct {
}

func (*UnimplementedIndexGatewayServer) QueryIndex(req *QueryIndexRequest, srv IndexGateway_QueryIndexServer) error {
	return status.Errorf(codes.Unimplemented, "method QueryIndex not implemented")
}
func (*UnimplementedIndexGatewayServer) GetChunkRef(ctx context.Context, req *GetChunkRefRequest) (*GetChunkRefResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetChunkRef not implemented")
}
func (*UnimplementedIndexGatewayServer) GetSeries(ctx context.Context, req *GetSeriesRequest) (*GetSeriesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetSeries not implemented")
}
func (*UnimplementedIndexGatewayServer) LabelNamesForMetricName(ctx context.Context, req *LabelNamesForMetricNameRequest) (*LabelResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method LabelNamesForMetricName not implemented")
}
func (*UnimplementedIndexGatewayServer) LabelValuesForMetricName(ctx context.Context, req *LabelValuesForMetricNameRequest) (*LabelResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method LabelValuesForMetricName not implemented")
}
func (*UnimplementedIndexGatewayServer) GetStats(ctx context.Context, req *IndexStatsRequest) (*IndexStatsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetStats not implemented")
}

func RegisterIndexGatewayServer(s *grpc.Server, srv IndexGatewayServer) {
	s.RegisterService(&_IndexGateway_serviceDesc, srv)
}

func _IndexGateway_QueryIndex_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(QueryIndexRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(IndexGatewayServer).QueryIndex(m, &indexGatewayQueryIndexServer{stream})
}

type IndexGateway_QueryIndexServer interface {
	Send(*QueryIndexResponse) error
	grpc.ServerStream
}

type indexGatewayQueryIndexServer struct {
	grpc.ServerStream
}

func (x *indexGatewayQueryIndexServer) Send(m *QueryIndexResponse) error {
	return x.ServerStream.SendMsg(m)
}

func _IndexGateway_GetChunkRef_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetChunkRefRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(IndexGatewayServer).GetChunkRef(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/indexgatewaypb.IndexGateway/GetChunkRef",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(IndexGatewayServer).GetChunkRef(ctx, req.(*GetChunkRefRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _IndexGateway_GetSeries_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetSeriesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(IndexGatewayServer).GetSeries(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/indexgatewaypb.IndexGateway/GetSeries",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(IndexGatewayServer).GetSeries(ctx, req.(*GetSeriesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _IndexGateway_LabelNamesForMetricName_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(LabelNamesForMetricNameRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(IndexGatewayServer).LabelNamesForMetricName(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/indexgatewaypb.IndexGateway/LabelNamesForMetricName",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(IndexGatewayServer).LabelNamesForMetricName(ctx, req.(*LabelNamesForMetricNameRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _IndexGateway_LabelValuesForMetricName_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(LabelValuesForMetricNameRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(IndexGatewayServer).LabelValuesForMetricName(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/indexgatewaypb.IndexGateway/LabelValuesForMetricName",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(IndexGatewayServer).LabelValuesForMetricName(ctx, req.(*LabelValuesForMetricNameRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _IndexGateway_GetStats_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(IndexStatsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(IndexGatewayServer).GetStats(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/indexgatewaypb.IndexGateway/GetStats",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(IndexGatewayServer).GetStats(ctx, req.(*IndexStatsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _IndexGateway_serviceDesc = grpc.ServiceDesc{
	ServiceName: "indexgatewaypb.IndexGateway",
	HandlerType: (*IndexGatewayServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetChunkRef",
			Handler:    _IndexGateway_GetChunkRef_Handler,
		},
		{
			MethodName: "GetSeries",
			Handler:    _IndexGateway_GetSeries_Handler,
		},
		{
			MethodName: "LabelNamesForMetricName",
			Handler:    _IndexGateway_LabelNamesForMetricName_Handler,
		},
		{
			MethodName: "LabelValuesForMetricName",
			Handler:    _IndexGateway_LabelValuesForMetricName_Handler,
		},
		{
			MethodName: "GetStats",
			Handler:    _IndexGateway_GetStats_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "QueryIndex",
			Handler:       _IndexGateway_QueryIndex_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "pkg/logproto/indexgateway.proto",
}
