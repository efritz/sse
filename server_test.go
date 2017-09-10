package sse

import (
	"github.com/aphistic/sweet"
	. "github.com/onsi/gomega"
)

type ServerSuite struct{}

func (s *ServerSuite) TestSometing(t sweet.T) {
	Expect(true).To(BeTrue())
}
