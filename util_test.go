package sse

import (
	"math"

	"github.com/aphistic/sweet"
	. "github.com/onsi/gomega"
)

type UtilSuite struct{}

func (s *UtilSuite) TestSerializeSSE(t sweet.T) {
	data, err := serializeSSE(map[string]interface{}{
		"foo": 3.141,
		"bar": "baz",
	})

	Expect(err).To(BeNil())
	Expect(string(data)).To(HavePrefix("data:"))
	Expect(string(data)).To(HaveSuffix("\n\n"))

	Expect(string(data[5:])).To(MatchJSON(`{
		"foo": 3.141,
		"bar": "baz"
	}`))
}

func (s *UtilSuite) TestSerializeSSEMarshalError(t sweet.T) {
	_, err := serializeSSE(math.Inf(1))
	Expect(err).NotTo(BeNil())
}
