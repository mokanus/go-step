// Code generated by protoc-gen-go. DO NOT EDIT.
// source: google/protobuf/any.proto

/*
Package any is a generated protocol buffer package.

It is generated from these files:

	google/protobuf/any.prot
	google/protobuf/any.proto

sages:

	Any
*/
package any

imp
It has these top-level messages:
	Any
*/
package any

import proto "github.com/golang/protobuf/proto"
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
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

// `Any` contains an arbitrary serialized protocol buffer message along with a
//	Foo foo = ...;
//	Any any;
//	any.PackFrom(foo);
//	...
//	if (any.UnpackTo(&foo)) {
//	  ...
//	}
//     Foo foo = ...;
//     Any any;
//     any.PackFrom(foo);
//	   Foo foo = ...;
//	   Any any = Any.pack(foo);
//	   ...
//	   if (any.is(Foo.class)) {
//	     foo = any.unpack(Foo.class);
//	   }
//
//	Example 3: Pack and unpack a message in Python.
//     Any any = Any.pack(foo);
//	   foo = Foo(...)
//	   any = Any()
//	   any.Pack(foo)
//	   ...
//	   if any.Is(Foo.DESCRIPTOR):
//	     any.Unpack(foo)
//	     ...
//     foo = Foo(...)
//	Example 4: Pack and unpack a message in Go
//     any.Pack(foo)
//	    foo := &pb.Foo{...}
//	    any, err := ptypes.MarshalAny(foo)
//	    ...
//	    foo := &pb.Foo{}
//	    if err := ptypes.UnmarshalAny(any, foo); err != nil {
//	      ...
//	    }
//      foo := &pb.Foo{...}
//      any, err := ptypes.MarshalAny(foo)
//      ...
//      foo := &pb.Foo{}
//      if err := ptypes.UnmarshalAny(any, foo); err != nil {
//        ...
//      }
// The pack methods provided by protobuf library will by default use
// 'type.googleapis.com/full.type.name' as the type URL and the unpack
// methods only use the fully qualified type name after the last '/'
// in the type URL, for example "foo.bar.com/x/y.z" will yield type
// name "y.z".
//
//	package google.profile;
//	message Person {
//	  string first_name = 1;
//	  string last_name = 2;
//	}
// additional field `@type` which contains the type URL. Example:
//	{
//	  "@type": "type.googleapis.com/google.profile.Person",
//	  "firstName": <string>,
//	  "lastName": <string>
//	}
//     }
//
//     {
//       "@type": "type.googleapis.com/google.profile.Person",
//       "firstName": <string>,
//       "lastName": <string>
//	{
//	  "@type": "type.googleapis.com/google.protobuf.Duration",
//	  "value": "1.212s"
//	}
// field. Example (for message [google.protobuf.Duration][]):
//
//     {
//       "@type": "type.googleapis.com/google.protobuf.Duration",
//       "value": "1.212s"
//     }
//
type Any struct {
	// A URL/resource name whose content describes the type of the
	// serialized protocol buffer message.
	//
	// For URLs which use the scheme `http`, `https`, or no scheme, the
	// following restrictions and interpretations apply:
	//
	// * If no scheme is provided, `https` is assumed.
	// * The last segment of the URL's path must represent the fully
	//   qualified name of the type (as in `path/google.protobuf.Duration`).
	//   The name should be in a canonical form (e.g., leading "." is
	//   not accepted).
	// * An HTTP GET on the URL must yield a [google.protobuf.Type][]
	//   value in binary format, or produce an error.
	// * Applications are allowed to cache lookup results based on the
	//   URL, or have them precompiled into a binary to avoid any
	//   lookup. Therefore, binary compatibility needs to be preserved
	//   on changes to types. (Use versioned type names to manage
	//   breaking changes.)
	//
	// Schemes other than `http`, `https` (or the empty scheme) might be
	// used with implementation specific semantics.
	//
	TypeUrl string `protobuf:"bytes,1,opt,name=type_url,json=typeUrl" json:"type_url,omitempty"`
	// Must be a valid serialized protocol buffer of the above specified type.
	Value []byte `protobuf:"bytes,2,opt,name=value,proto3" json:"value,omitempty"`
}

func (m *Any) Reset()                    { *m = Any{} }
func (m *Any) String() string            { return proto.CompactTextString(m) }
func (*Any) ProtoMessage()               {}
func (*Any) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }
func (*Any) XXX_WellKnownType() string   { return "Any" }

func (m *Any) GetTypeUrl() string {
	if m != nil {
		return m.TypeUrl
	}
	return ""
}

func (m *Any) GetValue() []byte {
	if m != nil {
		return m.Value
	}
	return nil
}

func init() {
	proto.RegisterType((*Any)(nil), "google.protobuf.Any")
}

func init() { proto.RegisterFile("google/protobuf/any.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 185 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x92, 0x4c, 0xcf, 0xcf, 0x4f,
	0xcf, 0x49, 0xd5, 0x2f, 0x28, 0xca, 0x2f, 0xc9, 0x4f, 0x2a, 0x4d, 0xd3, 0x4f, 0xcc, 0xab, 0xd4,
	0x03, 0x73, 0x84, 0xf8, 0x21, 0x52, 0x7a, 0x30, 0x29, 0x25, 0x33, 0x2e, 0x66, 0xc7, 0xbc, 0x4a,
	0x21, 0x49, 0x2e, 0x8e, 0x92, 0xca, 0x82, 0xd4, 0xf8, 0xd2, 0xa2, 0x1c, 0x09, 0x46, 0x05, 0x46,
	0x0d, 0xce, 0x20, 0x76, 0x10, 0x3f, 0xb4, 0x28, 0x47, 0x48, 0x84, 0x8b, 0xb5, 0x2c, 0x31, 0xa7,
	0x34, 0x55, 0x82, 0x49, 0x81, 0x51, 0x83, 0x27, 0x08, 0xc2, 0x71, 0xca, 0xe7, 0x12, 0x4e, 0xce,
	0xcf, 0xd5, 0x43, 0x33, 0xce, 0x89, 0xc3, 0x31, 0xaf, 0x32, 0x00, 0xc4, 0x09, 0x60, 0x8c, 0x52,
	0x4d, 0xcf, 0x2c, 0xc9, 0x28, 0x4d, 0xd2, 0x4b, 0xce, 0xcf, 0xd5, 0x4f, 0xcf, 0xcf, 0x49, 0xcc,
	0x4b, 0x47, 0xb8, 0xa8, 0x00, 0x64, 0x7a, 0x31, 0xc8, 0x61, 0x8b, 0x98, 0x98, 0xdd, 0x03, 0x9c,
	0x56, 0x31, 0xc9, 0xb9, 0x43, 0x8c, 0x0a, 0x80, 0x2a, 0xd1, 0x0b, 0x4f, 0xcd, 0xc9, 0xf1, 0xce,
	0xcb, 0x2f, 0xcf, 0x0b, 0x01, 0x29, 0x4d, 0x62, 0x03, 0xeb, 0x35, 0x06, 0x04, 0x00, 0x00, 0xff,
	0xff, 0x13, 0xf8, 0xe8, 0x42, 0xdd, 0x00, 0x00, 0x00,
}
