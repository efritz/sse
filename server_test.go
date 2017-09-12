package sse

import (
	"io/ioutil"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/aphistic/sweet"
	"github.com/aphistic/sweet-junit"
	. "github.com/onsi/gomega"
)

func TestMain(m *testing.M) {
	RegisterFailHandler(sweet.GomegaFail)

	sweet.Run(m, func(s *sweet.S) {
		s.RegisterPlugin(junit.NewPlugin())
		s.AddSuite(&ServerSuite{})
	})
}

type ServerSuite struct{}

func (s *ServerSuite) TestServe(t sweet.T) {
	var (
		events = make(chan interface{})
		server = NewServer(events)
		sync   = make(chan struct{})
		w      = &CloseNotifierResponseRecorder{httptest.NewRecorder(), nil}
	)

	go server.Start()

	go func() {
		r, _ := http.NewRequest("GET", "/", nil)
		server.ServeHTTP(w, r)
		close(sync)
	}()

	<-time.After(time.Millisecond * 50)
	events <- map[string]int{"foo": 1}
	events <- map[string]int{"bar": 2}
	events <- map[string]int{"baz": 3}
	close(events)
	<-sync

	resp := w.Result()
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	Expect(resp.StatusCode).To(Equal(http.StatusOK))
	Expect(resp.Header.Get("Content-Type")).To(Equal("text/event-stream"))
	Expect(string(body)).To(ContainSubstring(`data:{"foo":1}`))
	Expect(string(body)).To(ContainSubstring(`data:{"bar":2}`))
	Expect(string(body)).To(ContainSubstring(`data:{"baz":3}`))
}

func (s *ServerSuite) TestServeMultiple(t sweet.T) {
	var (
		events = make(chan interface{})
		server = NewServer(events)
		sync   = make(chan struct{})
		w1     = &CloseNotifierResponseRecorder{httptest.NewRecorder(), nil}
		w2     = &CloseNotifierResponseRecorder{httptest.NewRecorder(), nil}
	)

	defer close(sync)
	go server.Start()

	go func() {
		r, _ := http.NewRequest("GET", "/", nil)
		server.ServeHTTP(w1, r)
		sync <- struct{}{}
	}()

	<-time.After(time.Millisecond * 50)
	events <- map[string]int{"foo": 1}
	events <- map[string]int{"bar": 2}
	events <- map[string]int{"baz": 3}

	go func() {
		r, _ := http.NewRequest("GET", "/", nil)
		server.ServeHTTP(w2, r)
		sync <- struct{}{}
	}()

	<-time.After(time.Millisecond * 50)
	events <- map[string]int{"bonk": 4}
	events <- map[string]int{"quux": 5}
	events <- map[string]int{"honk": 6}
	close(events)
	<-sync
	<-sync

	resp1 := w1.Result()
	defer resp1.Body.Close()
	body1, _ := ioutil.ReadAll(resp1.Body)

	resp2 := w2.Result()
	defer resp2.Body.Close()
	body2, _ := ioutil.ReadAll(resp2.Body)

	Expect(resp1.StatusCode).To(Equal(http.StatusOK))
	Expect(resp1.Header.Get("Content-Type")).To(Equal("text/event-stream"))
	Expect(string(body1)).To(ContainSubstring(`data:{"foo":1}`))
	Expect(string(body1)).To(ContainSubstring(`data:{"bar":2}`))
	Expect(string(body1)).To(ContainSubstring(`data:{"baz":3}`))
	Expect(string(body1)).To(ContainSubstring(`data:{"bonk":4}`))
	Expect(string(body1)).To(ContainSubstring(`data:{"quux":5}`))
	Expect(string(body1)).To(ContainSubstring(`data:{"honk":6}`))

	Expect(resp2.StatusCode).To(Equal(http.StatusOK))
	Expect(resp2.Header.Get("Content-Type")).To(Equal("text/event-stream"))
	Expect(string(body2)).To(ContainSubstring(`data:{"bonk":4}`))
	Expect(string(body2)).To(ContainSubstring(`data:{"quux":5}`))
	Expect(string(body2)).To(ContainSubstring(`data:{"honk":6}`))
}

func (s *ServerSuite) TestNoClients(t sweet.T) {
	var (
		events  = make(chan interface{})
		server  = NewServer(events)
		errChan = make(chan error)
	)

	go func() {
		errChan <- server.Start()
	}()

	events <- map[string]int{"foo": 1}
	events <- map[string]int{"bar": 2}
	events <- map[string]int{"baz": 3}
	close(events)

	Eventually(errChan).Should(Receive(BeNil()))
}

func (s *ServerSuite) TestSerializationError(t sweet.T) {
	var (
		events  = make(chan interface{})
		server  = NewServer(events)
		errChan = make(chan error)
	)

	defer close(events)

	go func() {
		errChan <- server.Start()
	}()

	events <- math.Inf(1)
	Eventually(errChan).Should(Receive(Not(BeNil())))
}

func (s *ServerSuite) TestClientDisconnect(t sweet.T) {
	var (
		events    = make(chan interface{})
		server    = NewServer(events)
		closeChan = make(chan bool)
		w         = &CloseNotifierResponseRecorder{httptest.NewRecorder(), closeChan}
	)

	go server.Start()

	go func() {
		r, _ := http.NewRequest("GET", "/", nil)
		server.ServeHTTP(w, r)
	}()

	<-time.After(time.Millisecond * 50)
	events <- map[string]int{"foo": 1}
	events <- map[string]int{"bar": 2}
	events <- map[string]int{"baz": 3}

	<-time.After(time.Millisecond * 50)
	close(closeChan)

	<-time.After(time.Millisecond * 50)
	events <- map[string]int{"bonk": 4}
	events <- map[string]int{"quux": 5}
	events <- map[string]int{"honk": 6}
	close(events)

	resp := w.Result()
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	Expect(resp.StatusCode).To(Equal(http.StatusOK))
	Expect(resp.Header.Get("Content-Type")).To(Equal("text/event-stream"))
	Expect(string(body)).To(ContainSubstring(`data:{"foo":1}`))
	Expect(string(body)).To(ContainSubstring(`data:{"bar":2}`))
	Expect(string(body)).To(ContainSubstring(`data:{"baz":3}`))
}

func (s *ServerSuite) TestSerializeSSE(t sweet.T) {
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

func (s *ServerSuite) TestSerializeSSEMarshalError(t sweet.T) {
	_, err := serializeSSE(math.Inf(1))
	Expect(err).NotTo(BeNil())
}

//
// Helpers

type CloseNotifierResponseRecorder struct {
	*httptest.ResponseRecorder
	ch chan bool
}

func (r *CloseNotifierResponseRecorder) CloseNotify() <-chan bool {
	if r.ch != nil {
		return r.ch
	}

	return make(chan bool)
}
