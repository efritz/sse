package sse

import (
	"io"

	"github.com/aphistic/sweet"
	. "github.com/onsi/gomega"
)

type ReaderSuite struct{}

func (s *ReaderSuite) TestRead(t sweet.T) {
	events := make(chan []byte, 2)
	reader := newEventReader(events)

	buffer := make([]byte, 4)
	events <- []byte("foobar")
	events <- []byte("bazbonk")
	close(events)

	n, err := reader.Read(buffer)
	Expect(err).To(BeNil())
	Expect(n).To(Equal(4))
	Expect(buffer[:n]).To(Equal([]byte("foob")))

	n, err = reader.Read(buffer)
	Expect(err).To(BeNil())
	Expect(n).To(Equal(2))
	Expect(buffer[:n]).To(Equal([]byte("ar")))

	n, err = reader.Read(buffer)
	Expect(err).To(BeNil())
	Expect(n).To(Equal(4))
	Expect(buffer[:n]).To(Equal([]byte("bazb")))

	n, err = reader.Read(buffer)
	Expect(err).To(BeNil())
	Expect(n).To(Equal(3))
	Expect(buffer[:n]).To(Equal([]byte("onk")))

	_, err = reader.Read(buffer)
	Expect(err).To(Equal(io.EOF))
}
