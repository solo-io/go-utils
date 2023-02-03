package protoutils_test

import (
    "encoding/json"
    "fmt"

    "github.com/gogo/protobuf/proto"
    "github.com/gogo/protobuf/types"
    . "github.com/onsi/ginkgo/v2"
    . "github.com/onsi/gomega"

    . "github.com/solo-io/go-utils/protoutils"
)

type testType struct {
    A string
    B b
}

func (t testType) Reset() {}

func (t testType) String() string {
    byt, _ := json.Marshal(t)
    return string(byt)
}

func (t testType) ProtoMessage() {}

type b struct {
    C string
    D string
}

var tests = []struct {
    in       interface{}
    expected proto.Message
}{
    {
        in: testType{
            A: "a",
            B: b{
                C: "c",
                D: "d",
            },
        },
        expected: &types.Struct{
            Fields: map[string]*types.Value{
                "A": {
                    Kind: &types.Value_StringValue{StringValue: "a"},
                },
                "B": {
                    Kind: &types.Value_StructValue{
                        StructValue: &types.Struct{
                            Fields: map[string]*types.Value{
                                "C": {
                                    Kind: &types.Value_StringValue{StringValue: "c"},
                                },
                                "D": {
                                    Kind: &types.Value_StringValue{StringValue: "d"},
                                },
                            },
                        },
                    },
                },
            },
        },
    },
    {
        in: map[string]interface{}{
            "A": "a",
            "B": b{
                C: "c",
                D: "d",
            },
        },
        expected: &types.Struct{
            Fields: map[string]*types.Value{
                "a": {
                    Kind: &types.Value_StringValue{StringValue: "b"},
                },
                "c": {
                    Kind: &types.Value_StringValue{StringValue: "d"},
                },
            },
        },
    },
}

// TODO: Fix this text
// This test is temporarily out of order because it assumed an old version of MarshalStruct which took random data
// However, it now takes a proto.message to begin with. So the setup is no longer valid
var _ = PDescribe("Protoutil Funcs", func() {
    Describe("MarshalStruct", func() {
        for _, test := range tests {
            It("returns a pb struct for object of the given type", func() {
                var tt testType
                byt, err := json.Marshal(test.in)
                fmt.Printf("byte: %s", byt)
                Expect(err).NotTo(HaveOccurred())
                err = UnmarshalBytes(byt, &tt)
                Expect(err).NotTo(HaveOccurred())
                pb, err := MarshalStruct(&tt)
                Expect(err).NotTo(HaveOccurred())
                Expect(pb).To(Equal(test.expected))
            })
        }
    })
})
