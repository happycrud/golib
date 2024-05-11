// Name is the name registered for the proto compressor.
package pjson

import (
	"fmt"

	"google.golang.org/grpc/encoding"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

const Name = "protojson"

func init() {
	encoding.RegisterCodec(codec{})
}

// RegisterProtoJSONCodec Customization of encoding and decoding can be achieved through methods
func RegisterProtoJSONCodec(maOption protojson.MarshalOptions, unMaOption protojson.UnmarshalOptions) {
	encoding.RegisterCodec(codec{mar: maOption, unMar: unMaOption})
}

// codec is a Codec implementation with protojson. It is a option choice for gRPC.
type codec struct {
	mar   protojson.MarshalOptions
	unMar protojson.UnmarshalOptions
}

func (c codec) Name() string {
	return Name
}
func (c codec) Marshal(v any) ([]byte, error) {
	vv, ok := v.(proto.Message)
	if !ok {
		return nil, fmt.Errorf("failed to marshal, message is %T, want proto.Message", v)
	}
	return c.mar.Marshal(vv)

}

func (c codec) Unmarshal(data []byte, v any) error {
	vv, ok := v.(proto.Message)
	if !ok {
		return fmt.Errorf("failed to unmarshal, message is %T, want proto.Message", v)
	}
	return c.unMar.Unmarshal(data, vv)
}
