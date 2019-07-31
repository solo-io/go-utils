package helmchart_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/solo-io/go-utils/installutils/helmchart"
)

type Config struct {
	Namespace *Namespace `json:"namespace,omitempty" desc:"create namespace"`
	Array     []Array    `json:"array,omitempty"`
	Bool      bool       `json:"booleanValue,omitempty"`
	Complex   Complex    `json:"complex,omitempty"`
}

type Complex struct {
	SomeMap map[string]string `json:"items"`
}

type Namespace struct {
	Create bool `json:"create" desc:"create the installation namespace"`
}

type Array struct {
	Something string `json:"something" desc:"create something"`
}

var _ = Describe("Docs", func() {
	It("should document helm values", func() {
		c := Config{
			Namespace: &Namespace{Create: true},
			Bool:      true,
			Complex:   Complex{SomeMap: map[string]string{"default": "yes"}},
		}
		docDesc := Doc(c)

		expectedDocs := HelmValues{
			{
				Key:          "namespace.create",
				Type:         "bool",
				DefaultValue: "true",
				Description:  "create the installation namespace",
			},
			{
				Key:          "array[].something",
				Type:         "string",
				DefaultValue: "",
				Description:  "create something",
			},
			{Key: "booleanValue", Type: "bool", DefaultValue: "true", Description: ""},
			{
				Key:          "complex.items.NAME",
				Type:         "string",
				DefaultValue: "",
				Description:  "",
			},
			{
				Key:          "complex.items.default",
				Type:         "string",
				DefaultValue: "yes",
				Description:  "",
			},
		}

		Expect(expectedDocs).To(Equal(docDesc))
	})

	It("should print markdown", func() {
		values := HelmValues{{Key: "key", Type: "type", DefaultValue: "default", Description: "desc"}}
		expected := "|Option|Type|Default Value|Description|\n|------|----|-----------|-------------|\n|key|type|default|desc|\n"

		Expect(expected).To(Equal(values.ToMarkdown()))
	})
})
