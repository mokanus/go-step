package util

import "github.com/golang/protobuf/proto"

func PbMarshal(msg proto.Message) ([]byte, error) {
	if msg == nil {
		return nil, nil
	}
	return proto.Marshal(msg)
}
