package mapdecoder_test

import (
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/solo-io/go-utils/mapstructureutils"
	"google.golang.org/protobuf/types/known/structpb"
)

var _ = Describe("Normalize Map Decode", func() {
	DescribeTable("should correctly decode JSON numbers into int64 or float64",
		func(input map[string]interface{}, expectedResult map[string]interface{}) {
			result := make(map[string]interface{})
			err := NormalizeMapDecode(input, &result)
			Expect(err).NotTo(HaveOccurred())

			Expect(result).To(Equal(expectedResult))
			value, err := structpb.NewValue(result)
			Expect(err).NotTo(HaveOccurred())
			Expect(value).ToNot(BeNil())
		},
		Entry("deeply nested JSON number to int",
			map[string]interface{}{
				"number": json.Number("10"),
				"float":  float64(10.5),
				"nested": map[string]interface{}{
					"number": json.Number("100"),
					"float":  float64(100.5),
					"deeplyNested": map[string]interface{}{
						"number": json.Number("1000"),
						"float":  float64(100.5),
					},
				},
			},
			map[string]interface{}{
				"number": int64(10),
				"float":  float64(10.5),
				"nested": map[string]interface{}{
					"number": int64(100),
					"float":  float64(100.5),
					"deeplyNested": map[string]interface{}{
						"number": int64(1000),
						"float":  float64(100.5),
					},
				},
			},
		),
	)
})
