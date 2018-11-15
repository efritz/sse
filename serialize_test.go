package sse

import (
	"math"

	"github.com/aphistic/sweet"
	. "github.com/onsi/gomega"
)

type SerializeSuite struct{}

func (s *SerializeSuite) TestSerializeEvent(t sweet.T) {
	data, err := serializeEvent(map[string]interface{}{
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

func (s *SerializeSuite) TestSerializeEventMarshalError(t sweet.T) {
	_, err := serializeEvent(math.Inf(1))
	Expect(err).NotTo(BeNil())
}
