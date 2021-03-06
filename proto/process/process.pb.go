// Code generated by protoc-gen-go. DO NOT EDIT.
// source: process.proto

/*
Package proto is a generated protocol buffer package.

It is generated from these files:
	process.proto

It has these top-level messages:
	KeyValuePair
	ConfigureProcessReq
	ConfigureProcessResp
	RestartProcessReq
	RestartProcessResp
	GetProcessStateReq
	GetProcessStateResp
*/
package proto

import proto1 "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto1.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto1.ProtoPackageIsVersion2 // please upgrade the proto package

type ProcessState int32

const (
	ProcessState_NOT_RUNNING ProcessState = 0
	ProcessState_STARTING    ProcessState = 1
	ProcessState_RUNNING     ProcessState = 2
	ProcessState_FINISHED    ProcessState = 3
)

var ProcessState_name = map[int32]string{
	0: "NOT_RUNNING",
	1: "STARTING",
	2: "RUNNING",
	3: "FINISHED",
}
var ProcessState_value = map[string]int32{
	"NOT_RUNNING": 0,
	"STARTING":    1,
	"RUNNING":     2,
	"FINISHED":    3,
}

func (x ProcessState) String() string {
	return proto1.EnumName(ProcessState_name, int32(x))
}
func (ProcessState) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

type KeyValuePair struct {
	Key   string `protobuf:"bytes,1,opt,name=key" json:"key,omitempty"`
	Value string `protobuf:"bytes,2,opt,name=value" json:"value,omitempty"`
}

func (m *KeyValuePair) Reset()                    { *m = KeyValuePair{} }
func (m *KeyValuePair) String() string            { return proto1.CompactTextString(m) }
func (*KeyValuePair) ProtoMessage()               {}
func (*KeyValuePair) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *KeyValuePair) GetKey() string {
	if m != nil {
		return m.Key
	}
	return ""
}

func (m *KeyValuePair) GetValue() string {
	if m != nil {
		return m.Value
	}
	return ""
}

type ConfigureProcessReq struct {
	ProcessName   string          `protobuf:"bytes,1,opt,name=process_name,json=processName" json:"process_name,omitempty"`
	StartCmd      []string        `protobuf:"bytes,2,rep,name=start_cmd,json=startCmd" json:"start_cmd,omitempty"`
	RunPath       string          `protobuf:"bytes,3,opt,name=run_path,json=runPath" json:"run_path,omitempty"`
	Enabled       bool            `protobuf:"varint,4,opt,name=enabled" json:"enabled,omitempty"`
	ListeningPort uint32          `protobuf:"varint,5,opt,name=listening_port,json=listeningPort" json:"listening_port,omitempty"`
	EnvVars       []*KeyValuePair `protobuf:"bytes,6,rep,name=env_vars,json=envVars" json:"env_vars,omitempty"`
}

func (m *ConfigureProcessReq) Reset()                    { *m = ConfigureProcessReq{} }
func (m *ConfigureProcessReq) String() string            { return proto1.CompactTextString(m) }
func (*ConfigureProcessReq) ProtoMessage()               {}
func (*ConfigureProcessReq) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *ConfigureProcessReq) GetProcessName() string {
	if m != nil {
		return m.ProcessName
	}
	return ""
}

func (m *ConfigureProcessReq) GetStartCmd() []string {
	if m != nil {
		return m.StartCmd
	}
	return nil
}

func (m *ConfigureProcessReq) GetRunPath() string {
	if m != nil {
		return m.RunPath
	}
	return ""
}

func (m *ConfigureProcessReq) GetEnabled() bool {
	if m != nil {
		return m.Enabled
	}
	return false
}

func (m *ConfigureProcessReq) GetListeningPort() uint32 {
	if m != nil {
		return m.ListeningPort
	}
	return 0
}

func (m *ConfigureProcessReq) GetEnvVars() []*KeyValuePair {
	if m != nil {
		return m.EnvVars
	}
	return nil
}

type ConfigureProcessResp struct {
}

func (m *ConfigureProcessResp) Reset()                    { *m = ConfigureProcessResp{} }
func (m *ConfigureProcessResp) String() string            { return proto1.CompactTextString(m) }
func (*ConfigureProcessResp) ProtoMessage()               {}
func (*ConfigureProcessResp) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

type RestartProcessReq struct {
	ProcessName string          `protobuf:"bytes,1,opt,name=process_name,json=processName" json:"process_name,omitempty"`
	StartCmd    []string        `protobuf:"bytes,2,rep,name=start_cmd,json=startCmd" json:"start_cmd,omitempty"`
	EnvVars     []*KeyValuePair `protobuf:"bytes,3,rep,name=env_vars,json=envVars" json:"env_vars,omitempty"`
}

func (m *RestartProcessReq) Reset()                    { *m = RestartProcessReq{} }
func (m *RestartProcessReq) String() string            { return proto1.CompactTextString(m) }
func (*RestartProcessReq) ProtoMessage()               {}
func (*RestartProcessReq) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func (m *RestartProcessReq) GetProcessName() string {
	if m != nil {
		return m.ProcessName
	}
	return ""
}

func (m *RestartProcessReq) GetStartCmd() []string {
	if m != nil {
		return m.StartCmd
	}
	return nil
}

func (m *RestartProcessReq) GetEnvVars() []*KeyValuePair {
	if m != nil {
		return m.EnvVars
	}
	return nil
}

type RestartProcessResp struct {
}

func (m *RestartProcessResp) Reset()                    { *m = RestartProcessResp{} }
func (m *RestartProcessResp) String() string            { return proto1.CompactTextString(m) }
func (*RestartProcessResp) ProtoMessage()               {}
func (*RestartProcessResp) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

type GetProcessStateReq struct {
	ProcessName string `protobuf:"bytes,1,opt,name=process_name,json=processName" json:"process_name,omitempty"`
}

func (m *GetProcessStateReq) Reset()                    { *m = GetProcessStateReq{} }
func (m *GetProcessStateReq) String() string            { return proto1.CompactTextString(m) }
func (*GetProcessStateReq) ProtoMessage()               {}
func (*GetProcessStateReq) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{5} }

func (m *GetProcessStateReq) GetProcessName() string {
	if m != nil {
		return m.ProcessName
	}
	return ""
}

type GetProcessStateResp struct {
	ProcessState ProcessState `protobuf:"varint,1,opt,name=process_state,json=processState,enum=proto.ProcessState" json:"process_state,omitempty"`
}

func (m *GetProcessStateResp) Reset()                    { *m = GetProcessStateResp{} }
func (m *GetProcessStateResp) String() string            { return proto1.CompactTextString(m) }
func (*GetProcessStateResp) ProtoMessage()               {}
func (*GetProcessStateResp) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{6} }

func (m *GetProcessStateResp) GetProcessState() ProcessState {
	if m != nil {
		return m.ProcessState
	}
	return ProcessState_NOT_RUNNING
}

func init() {
	proto1.RegisterType((*KeyValuePair)(nil), "proto.KeyValuePair")
	proto1.RegisterType((*ConfigureProcessReq)(nil), "proto.ConfigureProcessReq")
	proto1.RegisterType((*ConfigureProcessResp)(nil), "proto.ConfigureProcessResp")
	proto1.RegisterType((*RestartProcessReq)(nil), "proto.RestartProcessReq")
	proto1.RegisterType((*RestartProcessResp)(nil), "proto.RestartProcessResp")
	proto1.RegisterType((*GetProcessStateReq)(nil), "proto.GetProcessStateReq")
	proto1.RegisterType((*GetProcessStateResp)(nil), "proto.GetProcessStateResp")
	proto1.RegisterEnum("proto.ProcessState", ProcessState_name, ProcessState_value)
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for ProcessService service

type ProcessServiceClient interface {
	ConfigureProcess(ctx context.Context, in *ConfigureProcessReq, opts ...grpc.CallOption) (*ConfigureProcessResp, error)
	RestartProcess(ctx context.Context, in *RestartProcessReq, opts ...grpc.CallOption) (*RestartProcessResp, error)
	GetProcessState(ctx context.Context, in *GetProcessStateReq, opts ...grpc.CallOption) (*GetProcessStateResp, error)
}

type processServiceClient struct {
	cc *grpc.ClientConn
}

func NewProcessServiceClient(cc *grpc.ClientConn) ProcessServiceClient {
	return &processServiceClient{cc}
}

func (c *processServiceClient) ConfigureProcess(ctx context.Context, in *ConfigureProcessReq, opts ...grpc.CallOption) (*ConfigureProcessResp, error) {
	out := new(ConfigureProcessResp)
	err := grpc.Invoke(ctx, "/proto.ProcessService/ConfigureProcess", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *processServiceClient) RestartProcess(ctx context.Context, in *RestartProcessReq, opts ...grpc.CallOption) (*RestartProcessResp, error) {
	out := new(RestartProcessResp)
	err := grpc.Invoke(ctx, "/proto.ProcessService/RestartProcess", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *processServiceClient) GetProcessState(ctx context.Context, in *GetProcessStateReq, opts ...grpc.CallOption) (*GetProcessStateResp, error) {
	out := new(GetProcessStateResp)
	err := grpc.Invoke(ctx, "/proto.ProcessService/GetProcessState", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for ProcessService service

type ProcessServiceServer interface {
	ConfigureProcess(context.Context, *ConfigureProcessReq) (*ConfigureProcessResp, error)
	RestartProcess(context.Context, *RestartProcessReq) (*RestartProcessResp, error)
	GetProcessState(context.Context, *GetProcessStateReq) (*GetProcessStateResp, error)
}

func RegisterProcessServiceServer(s *grpc.Server, srv ProcessServiceServer) {
	s.RegisterService(&_ProcessService_serviceDesc, srv)
}

func _ProcessService_ConfigureProcess_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ConfigureProcessReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ProcessServiceServer).ConfigureProcess(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.ProcessService/ConfigureProcess",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ProcessServiceServer).ConfigureProcess(ctx, req.(*ConfigureProcessReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _ProcessService_RestartProcess_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RestartProcessReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ProcessServiceServer).RestartProcess(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.ProcessService/RestartProcess",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ProcessServiceServer).RestartProcess(ctx, req.(*RestartProcessReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _ProcessService_GetProcessState_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetProcessStateReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ProcessServiceServer).GetProcessState(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.ProcessService/GetProcessState",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ProcessServiceServer).GetProcessState(ctx, req.(*GetProcessStateReq))
	}
	return interceptor(ctx, in, info, handler)
}

var _ProcessService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "proto.ProcessService",
	HandlerType: (*ProcessServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ConfigureProcess",
			Handler:    _ProcessService_ConfigureProcess_Handler,
		},
		{
			MethodName: "RestartProcess",
			Handler:    _ProcessService_RestartProcess_Handler,
		},
		{
			MethodName: "GetProcessState",
			Handler:    _ProcessService_GetProcessState_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "process.proto",
}

func init() { proto1.RegisterFile("process.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 444 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xac, 0x52, 0x4d, 0x6f, 0xd3, 0x40,
	0x10, 0xad, 0x63, 0xd2, 0x38, 0x93, 0x8f, 0x86, 0x49, 0x84, 0xb6, 0xee, 0xc5, 0x58, 0x42, 0xb2,
	0x38, 0xe4, 0x50, 0x24, 0xe0, 0x8a, 0x0a, 0xa4, 0x01, 0xe1, 0x46, 0x4e, 0xe8, 0xd5, 0xda, 0x26,
	0x43, 0x6b, 0x91, 0xd8, 0x66, 0x77, 0x63, 0xa9, 0x67, 0x7e, 0x29, 0xff, 0x81, 0x1f, 0x80, 0xbc,
	0xb1, 0xa3, 0x34, 0x26, 0x52, 0x0f, 0x3d, 0xd9, 0xf3, 0xde, 0xce, 0xdb, 0xb7, 0x6f, 0x06, 0x3a,
	0xa9, 0x48, 0xe6, 0x24, 0xe5, 0x30, 0x15, 0x89, 0x4a, 0xb0, 0xae, 0x3f, 0xee, 0x5b, 0x68, 0x7f,
	0xa5, 0xfb, 0x6b, 0xbe, 0x5c, 0xd3, 0x84, 0x47, 0x02, 0x7b, 0x60, 0xfe, 0xa4, 0x7b, 0x66, 0x38,
	0x86, 0xd7, 0x0c, 0xf2, 0x5f, 0x1c, 0x40, 0x3d, 0xcb, 0x69, 0x56, 0xd3, 0xd8, 0xa6, 0x70, 0xff,
	0x18, 0xd0, 0xbf, 0x48, 0xe2, 0x1f, 0xd1, 0xed, 0x5a, 0xd0, 0x64, 0xa3, 0x1c, 0xd0, 0x2f, 0x7c,
	0x09, 0xed, 0xe2, 0x9e, 0x30, 0xe6, 0x2b, 0x2a, 0x84, 0x5a, 0x05, 0xe6, 0xf3, 0x15, 0xe1, 0x19,
	0x34, 0xa5, 0xe2, 0x42, 0x85, 0xf3, 0xd5, 0x82, 0xd5, 0x1c, 0xd3, 0x6b, 0x06, 0x96, 0x06, 0x2e,
	0x56, 0x0b, 0x3c, 0x05, 0x4b, 0xac, 0xe3, 0x30, 0xe5, 0xea, 0x8e, 0x99, 0xba, 0xb7, 0x21, 0xd6,
	0xf1, 0x84, 0xab, 0x3b, 0x64, 0xd0, 0xa0, 0x98, 0xdf, 0x2c, 0x69, 0xc1, 0x9e, 0x39, 0x86, 0x67,
	0x05, 0x65, 0x89, 0xaf, 0xa0, 0xbb, 0x8c, 0xa4, 0xa2, 0x38, 0x8a, 0x6f, 0xc3, 0x34, 0x11, 0x8a,
	0xd5, 0x1d, 0xc3, 0xeb, 0x04, 0x9d, 0x2d, 0x3a, 0x49, 0x84, 0xc2, 0x21, 0x58, 0x14, 0x67, 0x61,
	0xc6, 0x85, 0x64, 0xc7, 0x8e, 0xe9, 0xb5, 0xce, 0xfb, 0x9b, 0x30, 0x86, 0xbb, 0x11, 0xe4, 0xb2,
	0xd9, 0x35, 0x17, 0xd2, 0x7d, 0x01, 0x83, 0xea, 0x13, 0x65, 0xea, 0xfe, 0x36, 0xe0, 0x79, 0x40,
	0xda, 0xf2, 0x13, 0xbe, 0x7c, 0xd7, 0x9d, 0xf9, 0x08, 0x77, 0x03, 0xc0, 0x7d, 0x13, 0x32, 0x75,
	0xdf, 0x01, 0x8e, 0xa8, 0x44, 0xa6, 0x8a, 0x2b, 0x7a, 0x9c, 0x37, 0xf7, 0x0a, 0xfa, 0x95, 0x46,
	0x99, 0xe2, 0xfb, 0xed, 0xde, 0x84, 0x32, 0x07, 0x75, 0x6b, 0x77, 0x6b, 0xed, 0xc1, 0xf9, 0xf2,
	0x0e, 0x5d, 0xbd, 0xbe, 0x84, 0xf6, 0x2e, 0x8b, 0x27, 0xd0, 0xf2, 0xaf, 0x66, 0x61, 0xf0, 0xdd,
	0xf7, 0xc7, 0xfe, 0xa8, 0x77, 0x84, 0x6d, 0xb0, 0xa6, 0xb3, 0x0f, 0xc1, 0x2c, 0xaf, 0x0c, 0x6c,
	0x41, 0xa3, 0xa4, 0x6a, 0x39, 0xf5, 0x79, 0xec, 0x8f, 0xa7, 0x97, 0x9f, 0x3e, 0xf6, 0xcc, 0xf3,
	0xbf, 0x06, 0x74, 0x4b, 0x29, 0x12, 0x59, 0x34, 0x27, 0xfc, 0x06, 0xbd, 0xfd, 0xd1, 0xa0, 0x5d,
	0x78, 0xfa, 0xcf, 0x5a, 0xda, 0x67, 0x07, 0x39, 0x99, 0xba, 0x47, 0x38, 0x82, 0xee, 0xc3, 0x2c,
	0x91, 0x15, 0x0d, 0x95, 0x39, 0xdb, 0xa7, 0x07, 0x18, 0x2d, 0xf4, 0x05, 0x4e, 0xf6, 0x52, 0xc4,
	0xf2, 0x7c, 0x75, 0x2c, 0xb6, 0x7d, 0x88, 0xca, 0xb5, 0x6e, 0x8e, 0x35, 0xf9, 0xe6, 0x5f, 0x00,
	0x00, 0x00, 0xff, 0xff, 0x07, 0x3f, 0x46, 0x5b, 0xb9, 0x03, 0x00, 0x00,
}
