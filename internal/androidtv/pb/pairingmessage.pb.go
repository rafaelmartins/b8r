// pairingmessage.proto

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.6
// 	protoc        v5.29.3
// source: pairingmessage.proto

package pb

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

type RoleType int32

const (
	RoleType_ROLE_TYPE_UNKNOWN RoleType = 0
	RoleType_ROLE_TYPE_INPUT   RoleType = 1
	RoleType_ROLE_TYPE_OUTPUT  RoleType = 2
	RoleType_UNRECOGNIZED      RoleType = -1
)

// Enum value maps for RoleType.
var (
	RoleType_name = map[int32]string{
		0:  "ROLE_TYPE_UNKNOWN",
		1:  "ROLE_TYPE_INPUT",
		2:  "ROLE_TYPE_OUTPUT",
		-1: "UNRECOGNIZED",
	}
	RoleType_value = map[string]int32{
		"ROLE_TYPE_UNKNOWN": 0,
		"ROLE_TYPE_INPUT":   1,
		"ROLE_TYPE_OUTPUT":  2,
		"UNRECOGNIZED":      -1,
	}
)

func (x RoleType) Enum() *RoleType {
	p := new(RoleType)
	*p = x
	return p
}

func (x RoleType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (RoleType) Descriptor() protoreflect.EnumDescriptor {
	return file_pairingmessage_proto_enumTypes[0].Descriptor()
}

func (RoleType) Type() protoreflect.EnumType {
	return &file_pairingmessage_proto_enumTypes[0]
}

func (x RoleType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use RoleType.Descriptor instead.
func (RoleType) EnumDescriptor() ([]byte, []int) {
	return file_pairingmessage_proto_rawDescGZIP(), []int{0}
}

type PairingEncoding_EncodingType int32

const (
	PairingEncoding_ENCODING_TYPE_UNKNOWN      PairingEncoding_EncodingType = 0
	PairingEncoding_ENCODING_TYPE_ALPHANUMERIC PairingEncoding_EncodingType = 1
	PairingEncoding_ENCODING_TYPE_NUMERIC      PairingEncoding_EncodingType = 2
	PairingEncoding_ENCODING_TYPE_HEXADECIMAL  PairingEncoding_EncodingType = 3
	PairingEncoding_ENCODING_TYPE_QRCODE       PairingEncoding_EncodingType = 4
	PairingEncoding_UNRECOGNIZED               PairingEncoding_EncodingType = -1
)

// Enum value maps for PairingEncoding_EncodingType.
var (
	PairingEncoding_EncodingType_name = map[int32]string{
		0:  "ENCODING_TYPE_UNKNOWN",
		1:  "ENCODING_TYPE_ALPHANUMERIC",
		2:  "ENCODING_TYPE_NUMERIC",
		3:  "ENCODING_TYPE_HEXADECIMAL",
		4:  "ENCODING_TYPE_QRCODE",
		-1: "UNRECOGNIZED",
	}
	PairingEncoding_EncodingType_value = map[string]int32{
		"ENCODING_TYPE_UNKNOWN":      0,
		"ENCODING_TYPE_ALPHANUMERIC": 1,
		"ENCODING_TYPE_NUMERIC":      2,
		"ENCODING_TYPE_HEXADECIMAL":  3,
		"ENCODING_TYPE_QRCODE":       4,
		"UNRECOGNIZED":               -1,
	}
)

func (x PairingEncoding_EncodingType) Enum() *PairingEncoding_EncodingType {
	p := new(PairingEncoding_EncodingType)
	*p = x
	return p
}

func (x PairingEncoding_EncodingType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (PairingEncoding_EncodingType) Descriptor() protoreflect.EnumDescriptor {
	return file_pairingmessage_proto_enumTypes[1].Descriptor()
}

func (PairingEncoding_EncodingType) Type() protoreflect.EnumType {
	return &file_pairingmessage_proto_enumTypes[1]
}

func (x PairingEncoding_EncodingType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use PairingEncoding_EncodingType.Descriptor instead.
func (PairingEncoding_EncodingType) EnumDescriptor() ([]byte, []int) {
	return file_pairingmessage_proto_rawDescGZIP(), []int{2, 0}
}

type PairingMessage_Status int32

const (
	PairingMessage_UNKNOWN                  PairingMessage_Status = 0
	PairingMessage_STATUS_OK                PairingMessage_Status = 200
	PairingMessage_STATUS_ERROR             PairingMessage_Status = 400
	PairingMessage_STATUS_BAD_CONFIGURATION PairingMessage_Status = 401
	PairingMessage_STATUS_BAD_SECRET        PairingMessage_Status = 402
	PairingMessage_UNRECOGNIZED             PairingMessage_Status = -1
)

// Enum value maps for PairingMessage_Status.
var (
	PairingMessage_Status_name = map[int32]string{
		0:   "UNKNOWN",
		200: "STATUS_OK",
		400: "STATUS_ERROR",
		401: "STATUS_BAD_CONFIGURATION",
		402: "STATUS_BAD_SECRET",
		-1:  "UNRECOGNIZED",
	}
	PairingMessage_Status_value = map[string]int32{
		"UNKNOWN":                  0,
		"STATUS_OK":                200,
		"STATUS_ERROR":             400,
		"STATUS_BAD_CONFIGURATION": 401,
		"STATUS_BAD_SECRET":        402,
		"UNRECOGNIZED":             -1,
	}
)

func (x PairingMessage_Status) Enum() *PairingMessage_Status {
	p := new(PairingMessage_Status)
	*p = x
	return p
}

func (x PairingMessage_Status) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (PairingMessage_Status) Descriptor() protoreflect.EnumDescriptor {
	return file_pairingmessage_proto_enumTypes[2].Descriptor()
}

func (PairingMessage_Status) Type() protoreflect.EnumType {
	return &file_pairingmessage_proto_enumTypes[2]
}

func (x PairingMessage_Status) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use PairingMessage_Status.Descriptor instead.
func (PairingMessage_Status) EnumDescriptor() ([]byte, []int) {
	return file_pairingmessage_proto_rawDescGZIP(), []int{8, 0}
}

type PairingRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	ClientName    string                 `protobuf:"bytes,2,opt,name=client_name,json=clientName,proto3" json:"client_name,omitempty"`
	ServiceName   string                 `protobuf:"bytes,1,opt,name=service_name,json=serviceName,proto3" json:"service_name,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *PairingRequest) Reset() {
	*x = PairingRequest{}
	mi := &file_pairingmessage_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *PairingRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PairingRequest) ProtoMessage() {}

func (x *PairingRequest) ProtoReflect() protoreflect.Message {
	mi := &file_pairingmessage_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PairingRequest.ProtoReflect.Descriptor instead.
func (*PairingRequest) Descriptor() ([]byte, []int) {
	return file_pairingmessage_proto_rawDescGZIP(), []int{0}
}

func (x *PairingRequest) GetClientName() string {
	if x != nil {
		return x.ClientName
	}
	return ""
}

func (x *PairingRequest) GetServiceName() string {
	if x != nil {
		return x.ServiceName
	}
	return ""
}

type PairingRequestAck struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	ServerName    string                 `protobuf:"bytes,1,opt,name=server_name,json=serverName,proto3" json:"server_name,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *PairingRequestAck) Reset() {
	*x = PairingRequestAck{}
	mi := &file_pairingmessage_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *PairingRequestAck) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PairingRequestAck) ProtoMessage() {}

func (x *PairingRequestAck) ProtoReflect() protoreflect.Message {
	mi := &file_pairingmessage_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PairingRequestAck.ProtoReflect.Descriptor instead.
func (*PairingRequestAck) Descriptor() ([]byte, []int) {
	return file_pairingmessage_proto_rawDescGZIP(), []int{1}
}

func (x *PairingRequestAck) GetServerName() string {
	if x != nil {
		return x.ServerName
	}
	return ""
}

type PairingEncoding struct {
	state         protoimpl.MessageState       `protogen:"open.v1"`
	Type          PairingEncoding_EncodingType `protobuf:"varint,1,opt,name=type,proto3,enum=pairing.PairingEncoding_EncodingType" json:"type,omitempty"`
	SymbolLength  uint32                       `protobuf:"varint,2,opt,name=symbol_length,json=symbolLength,proto3" json:"symbol_length,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *PairingEncoding) Reset() {
	*x = PairingEncoding{}
	mi := &file_pairingmessage_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *PairingEncoding) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PairingEncoding) ProtoMessage() {}

func (x *PairingEncoding) ProtoReflect() protoreflect.Message {
	mi := &file_pairingmessage_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PairingEncoding.ProtoReflect.Descriptor instead.
func (*PairingEncoding) Descriptor() ([]byte, []int) {
	return file_pairingmessage_proto_rawDescGZIP(), []int{2}
}

func (x *PairingEncoding) GetType() PairingEncoding_EncodingType {
	if x != nil {
		return x.Type
	}
	return PairingEncoding_ENCODING_TYPE_UNKNOWN
}

func (x *PairingEncoding) GetSymbolLength() uint32 {
	if x != nil {
		return x.SymbolLength
	}
	return 0
}

type PairingOption struct {
	state           protoimpl.MessageState `protogen:"open.v1"`
	InputEncodings  []*PairingEncoding     `protobuf:"bytes,1,rep,name=input_encodings,json=inputEncodings,proto3" json:"input_encodings,omitempty"`
	OutputEncodings []*PairingEncoding     `protobuf:"bytes,2,rep,name=output_encodings,json=outputEncodings,proto3" json:"output_encodings,omitempty"`
	PreferredRole   RoleType               `protobuf:"varint,3,opt,name=preferred_role,json=preferredRole,proto3,enum=pairing.RoleType" json:"preferred_role,omitempty"`
	unknownFields   protoimpl.UnknownFields
	sizeCache       protoimpl.SizeCache
}

func (x *PairingOption) Reset() {
	*x = PairingOption{}
	mi := &file_pairingmessage_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *PairingOption) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PairingOption) ProtoMessage() {}

func (x *PairingOption) ProtoReflect() protoreflect.Message {
	mi := &file_pairingmessage_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PairingOption.ProtoReflect.Descriptor instead.
func (*PairingOption) Descriptor() ([]byte, []int) {
	return file_pairingmessage_proto_rawDescGZIP(), []int{3}
}

func (x *PairingOption) GetInputEncodings() []*PairingEncoding {
	if x != nil {
		return x.InputEncodings
	}
	return nil
}

func (x *PairingOption) GetOutputEncodings() []*PairingEncoding {
	if x != nil {
		return x.OutputEncodings
	}
	return nil
}

func (x *PairingOption) GetPreferredRole() RoleType {
	if x != nil {
		return x.PreferredRole
	}
	return RoleType_ROLE_TYPE_UNKNOWN
}

type PairingConfiguration struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Encoding      *PairingEncoding       `protobuf:"bytes,1,opt,name=encoding,proto3" json:"encoding,omitempty"`
	ClientRole    RoleType               `protobuf:"varint,2,opt,name=client_role,json=clientRole,proto3,enum=pairing.RoleType" json:"client_role,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *PairingConfiguration) Reset() {
	*x = PairingConfiguration{}
	mi := &file_pairingmessage_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *PairingConfiguration) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PairingConfiguration) ProtoMessage() {}

func (x *PairingConfiguration) ProtoReflect() protoreflect.Message {
	mi := &file_pairingmessage_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PairingConfiguration.ProtoReflect.Descriptor instead.
func (*PairingConfiguration) Descriptor() ([]byte, []int) {
	return file_pairingmessage_proto_rawDescGZIP(), []int{4}
}

func (x *PairingConfiguration) GetEncoding() *PairingEncoding {
	if x != nil {
		return x.Encoding
	}
	return nil
}

func (x *PairingConfiguration) GetClientRole() RoleType {
	if x != nil {
		return x.ClientRole
	}
	return RoleType_ROLE_TYPE_UNKNOWN
}

type PairingConfigurationAck struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *PairingConfigurationAck) Reset() {
	*x = PairingConfigurationAck{}
	mi := &file_pairingmessage_proto_msgTypes[5]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *PairingConfigurationAck) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PairingConfigurationAck) ProtoMessage() {}

func (x *PairingConfigurationAck) ProtoReflect() protoreflect.Message {
	mi := &file_pairingmessage_proto_msgTypes[5]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PairingConfigurationAck.ProtoReflect.Descriptor instead.
func (*PairingConfigurationAck) Descriptor() ([]byte, []int) {
	return file_pairingmessage_proto_rawDescGZIP(), []int{5}
}

type PairingSecret struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Secret        []byte                 `protobuf:"bytes,1,opt,name=secret,proto3" json:"secret,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *PairingSecret) Reset() {
	*x = PairingSecret{}
	mi := &file_pairingmessage_proto_msgTypes[6]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *PairingSecret) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PairingSecret) ProtoMessage() {}

func (x *PairingSecret) ProtoReflect() protoreflect.Message {
	mi := &file_pairingmessage_proto_msgTypes[6]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PairingSecret.ProtoReflect.Descriptor instead.
func (*PairingSecret) Descriptor() ([]byte, []int) {
	return file_pairingmessage_proto_rawDescGZIP(), []int{6}
}

func (x *PairingSecret) GetSecret() []byte {
	if x != nil {
		return x.Secret
	}
	return nil
}

type PairingSecretAck struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Secret        []byte                 `protobuf:"bytes,1,opt,name=secret,proto3" json:"secret,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *PairingSecretAck) Reset() {
	*x = PairingSecretAck{}
	mi := &file_pairingmessage_proto_msgTypes[7]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *PairingSecretAck) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PairingSecretAck) ProtoMessage() {}

func (x *PairingSecretAck) ProtoReflect() protoreflect.Message {
	mi := &file_pairingmessage_proto_msgTypes[7]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PairingSecretAck.ProtoReflect.Descriptor instead.
func (*PairingSecretAck) Descriptor() ([]byte, []int) {
	return file_pairingmessage_proto_rawDescGZIP(), []int{7}
}

func (x *PairingSecretAck) GetSecret() []byte {
	if x != nil {
		return x.Secret
	}
	return nil
}

type PairingMessage struct {
	state                   protoimpl.MessageState   `protogen:"open.v1"`
	ProtocolVersion         int32                    `protobuf:"varint,1,opt,name=protocol_version,json=protocolVersion,proto3" json:"protocol_version,omitempty"`
	Status                  PairingMessage_Status    `protobuf:"varint,2,opt,name=status,proto3,enum=pairing.PairingMessage_Status" json:"status,omitempty"`
	RequestCase             int32                    `protobuf:"varint,3,opt,name=request_case,json=requestCase,proto3" json:"request_case,omitempty"`
	PairingRequest          *PairingRequest          `protobuf:"bytes,10,opt,name=pairing_request,json=pairingRequest,proto3" json:"pairing_request,omitempty"`
	PairingRequestAck       *PairingRequestAck       `protobuf:"bytes,11,opt,name=pairing_request_ack,json=pairingRequestAck,proto3" json:"pairing_request_ack,omitempty"`
	PairingOption           *PairingOption           `protobuf:"bytes,20,opt,name=pairing_option,json=pairingOption,proto3" json:"pairing_option,omitempty"`
	PairingConfiguration    *PairingConfiguration    `protobuf:"bytes,30,opt,name=pairing_configuration,json=pairingConfiguration,proto3" json:"pairing_configuration,omitempty"`
	PairingConfigurationAck *PairingConfigurationAck `protobuf:"bytes,31,opt,name=pairing_configuration_ack,json=pairingConfigurationAck,proto3" json:"pairing_configuration_ack,omitempty"`
	PairingSecret           *PairingSecret           `protobuf:"bytes,40,opt,name=pairing_secret,json=pairingSecret,proto3" json:"pairing_secret,omitempty"`
	PairingSecretAck        *PairingSecretAck        `protobuf:"bytes,41,opt,name=pairing_secret_ack,json=pairingSecretAck,proto3" json:"pairing_secret_ack,omitempty"`
	unknownFields           protoimpl.UnknownFields
	sizeCache               protoimpl.SizeCache
}

func (x *PairingMessage) Reset() {
	*x = PairingMessage{}
	mi := &file_pairingmessage_proto_msgTypes[8]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *PairingMessage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PairingMessage) ProtoMessage() {}

func (x *PairingMessage) ProtoReflect() protoreflect.Message {
	mi := &file_pairingmessage_proto_msgTypes[8]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PairingMessage.ProtoReflect.Descriptor instead.
func (*PairingMessage) Descriptor() ([]byte, []int) {
	return file_pairingmessage_proto_rawDescGZIP(), []int{8}
}

func (x *PairingMessage) GetProtocolVersion() int32 {
	if x != nil {
		return x.ProtocolVersion
	}
	return 0
}

func (x *PairingMessage) GetStatus() PairingMessage_Status {
	if x != nil {
		return x.Status
	}
	return PairingMessage_UNKNOWN
}

func (x *PairingMessage) GetRequestCase() int32 {
	if x != nil {
		return x.RequestCase
	}
	return 0
}

func (x *PairingMessage) GetPairingRequest() *PairingRequest {
	if x != nil {
		return x.PairingRequest
	}
	return nil
}

func (x *PairingMessage) GetPairingRequestAck() *PairingRequestAck {
	if x != nil {
		return x.PairingRequestAck
	}
	return nil
}

func (x *PairingMessage) GetPairingOption() *PairingOption {
	if x != nil {
		return x.PairingOption
	}
	return nil
}

func (x *PairingMessage) GetPairingConfiguration() *PairingConfiguration {
	if x != nil {
		return x.PairingConfiguration
	}
	return nil
}

func (x *PairingMessage) GetPairingConfigurationAck() *PairingConfigurationAck {
	if x != nil {
		return x.PairingConfigurationAck
	}
	return nil
}

func (x *PairingMessage) GetPairingSecret() *PairingSecret {
	if x != nil {
		return x.PairingSecret
	}
	return nil
}

func (x *PairingMessage) GetPairingSecretAck() *PairingSecretAck {
	if x != nil {
		return x.PairingSecretAck
	}
	return nil
}

var File_pairingmessage_proto protoreflect.FileDescriptor

const file_pairingmessage_proto_rawDesc = "" +
	"\n" +
	"\x14pairingmessage.proto\x12\apairing\"T\n" +
	"\x0ePairingRequest\x12\x1f\n" +
	"\vclient_name\x18\x02 \x01(\tR\n" +
	"clientName\x12!\n" +
	"\fservice_name\x18\x01 \x01(\tR\vserviceName\"4\n" +
	"\x11PairingRequestAck\x12\x1f\n" +
	"\vserver_name\x18\x01 \x01(\tR\n" +
	"serverName\"\xac\x02\n" +
	"\x0fPairingEncoding\x129\n" +
	"\x04type\x18\x01 \x01(\x0e2%.pairing.PairingEncoding.EncodingTypeR\x04type\x12#\n" +
	"\rsymbol_length\x18\x02 \x01(\rR\fsymbolLength\"\xb8\x01\n" +
	"\fEncodingType\x12\x19\n" +
	"\x15ENCODING_TYPE_UNKNOWN\x10\x00\x12\x1e\n" +
	"\x1aENCODING_TYPE_ALPHANUMERIC\x10\x01\x12\x19\n" +
	"\x15ENCODING_TYPE_NUMERIC\x10\x02\x12\x1d\n" +
	"\x19ENCODING_TYPE_HEXADECIMAL\x10\x03\x12\x18\n" +
	"\x14ENCODING_TYPE_QRCODE\x10\x04\x12\x19\n" +
	"\fUNRECOGNIZED\x10\xff\xff\xff\xff\xff\xff\xff\xff\xff\x01\"\xd1\x01\n" +
	"\rPairingOption\x12A\n" +
	"\x0finput_encodings\x18\x01 \x03(\v2\x18.pairing.PairingEncodingR\x0einputEncodings\x12C\n" +
	"\x10output_encodings\x18\x02 \x03(\v2\x18.pairing.PairingEncodingR\x0foutputEncodings\x128\n" +
	"\x0epreferred_role\x18\x03 \x01(\x0e2\x11.pairing.RoleTypeR\rpreferredRole\"\x80\x01\n" +
	"\x14PairingConfiguration\x124\n" +
	"\bencoding\x18\x01 \x01(\v2\x18.pairing.PairingEncodingR\bencoding\x122\n" +
	"\vclient_role\x18\x02 \x01(\x0e2\x11.pairing.RoleTypeR\n" +
	"clientRole\"\x19\n" +
	"\x17PairingConfigurationAck\"'\n" +
	"\rPairingSecret\x12\x16\n" +
	"\x06secret\x18\x01 \x01(\fR\x06secret\"*\n" +
	"\x10PairingSecretAck\x12\x16\n" +
	"\x06secret\x18\x01 \x01(\fR\x06secret\"\xaa\x06\n" +
	"\x0ePairingMessage\x12)\n" +
	"\x10protocol_version\x18\x01 \x01(\x05R\x0fprotocolVersion\x126\n" +
	"\x06status\x18\x02 \x01(\x0e2\x1e.pairing.PairingMessage.StatusR\x06status\x12!\n" +
	"\frequest_case\x18\x03 \x01(\x05R\vrequestCase\x12@\n" +
	"\x0fpairing_request\x18\n" +
	" \x01(\v2\x17.pairing.PairingRequestR\x0epairingRequest\x12J\n" +
	"\x13pairing_request_ack\x18\v \x01(\v2\x1a.pairing.PairingRequestAckR\x11pairingRequestAck\x12=\n" +
	"\x0epairing_option\x18\x14 \x01(\v2\x16.pairing.PairingOptionR\rpairingOption\x12R\n" +
	"\x15pairing_configuration\x18\x1e \x01(\v2\x1d.pairing.PairingConfigurationR\x14pairingConfiguration\x12\\\n" +
	"\x19pairing_configuration_ack\x18\x1f \x01(\v2 .pairing.PairingConfigurationAckR\x17pairingConfigurationAck\x12=\n" +
	"\x0epairing_secret\x18( \x01(\v2\x16.pairing.PairingSecretR\rpairingSecret\x12G\n" +
	"\x12pairing_secret_ack\x18) \x01(\v2\x19.pairing.PairingSecretAckR\x10pairingSecretAck\"\x8a\x01\n" +
	"\x06Status\x12\v\n" +
	"\aUNKNOWN\x10\x00\x12\x0e\n" +
	"\tSTATUS_OK\x10\xc8\x01\x12\x11\n" +
	"\fSTATUS_ERROR\x10\x90\x03\x12\x1d\n" +
	"\x18STATUS_BAD_CONFIGURATION\x10\x91\x03\x12\x16\n" +
	"\x11STATUS_BAD_SECRET\x10\x92\x03\x12\x19\n" +
	"\fUNRECOGNIZED\x10\xff\xff\xff\xff\xff\xff\xff\xff\xff\x01*g\n" +
	"\bRoleType\x12\x15\n" +
	"\x11ROLE_TYPE_UNKNOWN\x10\x00\x12\x13\n" +
	"\x0fROLE_TYPE_INPUT\x10\x01\x12\x14\n" +
	"\x10ROLE_TYPE_OUTPUT\x10\x02\x12\x19\n" +
	"\fUNRECOGNIZED\x10\xff\xff\xff\xff\xff\xff\xff\xff\xff\x01B4Z2github.com/rafaelmartins/b8r/internal/androidtv/pbb\x06proto3"

var (
	file_pairingmessage_proto_rawDescOnce sync.Once
	file_pairingmessage_proto_rawDescData []byte
)

func file_pairingmessage_proto_rawDescGZIP() []byte {
	file_pairingmessage_proto_rawDescOnce.Do(func() {
		file_pairingmessage_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_pairingmessage_proto_rawDesc), len(file_pairingmessage_proto_rawDesc)))
	})
	return file_pairingmessage_proto_rawDescData
}

var file_pairingmessage_proto_enumTypes = make([]protoimpl.EnumInfo, 3)
var file_pairingmessage_proto_msgTypes = make([]protoimpl.MessageInfo, 9)
var file_pairingmessage_proto_goTypes = []any{
	(RoleType)(0),                     // 0: pairing.RoleType
	(PairingEncoding_EncodingType)(0), // 1: pairing.PairingEncoding.EncodingType
	(PairingMessage_Status)(0),        // 2: pairing.PairingMessage.Status
	(*PairingRequest)(nil),            // 3: pairing.PairingRequest
	(*PairingRequestAck)(nil),         // 4: pairing.PairingRequestAck
	(*PairingEncoding)(nil),           // 5: pairing.PairingEncoding
	(*PairingOption)(nil),             // 6: pairing.PairingOption
	(*PairingConfiguration)(nil),      // 7: pairing.PairingConfiguration
	(*PairingConfigurationAck)(nil),   // 8: pairing.PairingConfigurationAck
	(*PairingSecret)(nil),             // 9: pairing.PairingSecret
	(*PairingSecretAck)(nil),          // 10: pairing.PairingSecretAck
	(*PairingMessage)(nil),            // 11: pairing.PairingMessage
}
var file_pairingmessage_proto_depIdxs = []int32{
	1,  // 0: pairing.PairingEncoding.type:type_name -> pairing.PairingEncoding.EncodingType
	5,  // 1: pairing.PairingOption.input_encodings:type_name -> pairing.PairingEncoding
	5,  // 2: pairing.PairingOption.output_encodings:type_name -> pairing.PairingEncoding
	0,  // 3: pairing.PairingOption.preferred_role:type_name -> pairing.RoleType
	5,  // 4: pairing.PairingConfiguration.encoding:type_name -> pairing.PairingEncoding
	0,  // 5: pairing.PairingConfiguration.client_role:type_name -> pairing.RoleType
	2,  // 6: pairing.PairingMessage.status:type_name -> pairing.PairingMessage.Status
	3,  // 7: pairing.PairingMessage.pairing_request:type_name -> pairing.PairingRequest
	4,  // 8: pairing.PairingMessage.pairing_request_ack:type_name -> pairing.PairingRequestAck
	6,  // 9: pairing.PairingMessage.pairing_option:type_name -> pairing.PairingOption
	7,  // 10: pairing.PairingMessage.pairing_configuration:type_name -> pairing.PairingConfiguration
	8,  // 11: pairing.PairingMessage.pairing_configuration_ack:type_name -> pairing.PairingConfigurationAck
	9,  // 12: pairing.PairingMessage.pairing_secret:type_name -> pairing.PairingSecret
	10, // 13: pairing.PairingMessage.pairing_secret_ack:type_name -> pairing.PairingSecretAck
	14, // [14:14] is the sub-list for method output_type
	14, // [14:14] is the sub-list for method input_type
	14, // [14:14] is the sub-list for extension type_name
	14, // [14:14] is the sub-list for extension extendee
	0,  // [0:14] is the sub-list for field type_name
}

func init() { file_pairingmessage_proto_init() }
func file_pairingmessage_proto_init() {
	if File_pairingmessage_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_pairingmessage_proto_rawDesc), len(file_pairingmessage_proto_rawDesc)),
			NumEnums:      3,
			NumMessages:   9,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_pairingmessage_proto_goTypes,
		DependencyIndexes: file_pairingmessage_proto_depIdxs,
		EnumInfos:         file_pairingmessage_proto_enumTypes,
		MessageInfos:      file_pairingmessage_proto_msgTypes,
	}.Build()
	File_pairingmessage_proto = out.File
	file_pairingmessage_proto_goTypes = nil
	file_pairingmessage_proto_depIdxs = nil
}
