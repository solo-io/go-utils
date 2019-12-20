package gogoutils

import (
	gogojson "github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/types"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"
	structpb "github.com/golang/protobuf/ptypes/struct"
	"github.com/solo-io/go-utils/errors"
)

var (
	jsonpbMarshaler               = &jsonpb.Marshaler{OrigName: false}
	jsonpbMarshalerEmitZeroValues = &jsonpb.Marshaler{OrigName: false, EmitDefaults: true}

	gogoJsonpbMarshaler = &gogojson.Marshaler{OrigName: false}

	NilStructError = errors.New("cannot unmarshal nil struct")
)

func StructPbToGogo(structuredData *structpb.Struct) (*types.Struct, error) {
	if structuredData == nil {
		return nil, NilStructError
	}
	byt, err := proto.Marshal(structuredData)
	if err != nil {
		return nil, err
	}
	var st types.Struct
	if err := proto.Unmarshal(byt, &st); err != nil {
		return nil, err
	}
	return &st, nil
}

func StructGogoToPb(structuredData *types.Struct) (*structpb.Struct, error) {
	if structuredData == nil {
		return nil, NilStructError
	}
	byt, err := proto.Marshal(structuredData)
	if err != nil {
		return nil, err
	}
	var st structpb.Struct
	if err := proto.Unmarshal(byt, &st); err != nil {
		return nil, err
	}
	return &st, nil
}

func AnyPbToGogo(structuredData *any.Any) (*types.Any, error) {
	if structuredData == nil {
		return nil, NilStructError
	}
	byt, err := proto.Marshal(structuredData)
	if err != nil {
		return nil, err
	}
	var st types.Any
	if err := proto.Unmarshal(byt, &st); err != nil {
		return nil, err
	}
	return &st, nil
}

func AnyGogoToPb(structuredData *types.Any) (*any.Any, error) {
	if structuredData == nil {
		return nil, NilStructError
	}
	byt, err := proto.Marshal(structuredData)
	if err != nil {
		return nil, err
	}
	var st any.Any
	if err := proto.Unmarshal(byt, &st); err != nil {
		return nil, err
	}
	return &st, nil
}
