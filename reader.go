package sse

import (
	"io"
	"io/ioutil"
)

type eventReader struct {
	events  <-chan []byte
	current []byte
}

func newEventReader(events <-chan []byte) io.ReadCloser {
	return ioutil.NopCloser(&eventReader{
		events: events,
	})
}

func (r *eventReader) Read(p []byte) (int, error) {
	if len(r.current) == 0 {
		current, ok := <-r.events
		if !ok {
			return 0, io.EOF
		}

		r.current = current
	}

	copied := copy(p, r.current)
	r.current = r.current[copied:]
	return copied, nil
}
