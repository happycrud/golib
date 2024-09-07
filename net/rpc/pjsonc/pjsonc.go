package pjsonc

import (
	"encoding/json"

	"google.golang.org/grpc/encoding"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func init() {
	encoding.RegisterCodec(JSON{})
}

type JSON struct {
}

func (j JSON) Name() string {
	return "protojson"
}

func (j JSON) Marshal(v interface{}) (out []byte, err error) {

	if vs, ok2 := v.(string); ok2 {
		return []byte(vs), nil
	}
	if pm, ok := v.(proto.Message); ok {
		return protojson.Marshal(pm)
	}

	return json.Marshal(v)
}

func (j JSON) Unmarshal(data []byte, v interface{}) (err error) {
	if x, ok := v.(*Response); ok {
		x.Data = string(data)
		return nil
	}
	if pm, ok := v.(proto.Message); ok {
		return protojson.Unmarshal(data, pm)
	}
	return json.Unmarshal(data, v)

}

type Response struct {
	Data string
}
