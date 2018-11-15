package sse

import (
	"net/http"
	"sync"

	"github.com/efritz/response"
)

// Server is an http.Handler that will fanout Server-Sent Events
// to all connected clients.
type Server struct {
	events     <-chan interface{}
	requests   map[*http.Request]chan []byte
	mutex      sync.RWMutex
	bufferSize int

	ServeHTTP http.HandlerFunc
}

// NewServer creates a new server with the given event channel.
// The server returned by this function has not yet started.
func NewServer(events <-chan interface{}, configs ...ConfigFunc) *Server {
	s := &Server{
		events:     events,
		requests:   map[*http.Request]chan []byte{},
		bufferSize: 100,
	}

	for _, f := range configs {
		f(s)
	}

	s.ServeHTTP = response.Convert(s.Handler)
	return s
}

// Start will begin serializing events that come in on the
// event channel and sending the payload to each registered
// client. This method will block until the event channel
// has closed.
func (s *Server) Start() error {
	defer s.deregisterAll()

	for event := range s.events {
		data, err := serializeEvent(event)
		if err != nil {
			return err
		}

		s.publish(data)
	}

	return nil
}

// Handler converts an HTTP request into a streaming response.
// This can be used with libraries that utilize efritz/response.
// Alternatively, the ServeHTTP member on the Server struct is
// a http.HandlerFunc that can be served directly.
func (s *Server) Handler(r *http.Request) response.Response {
	events := s.register(r)
	progress := make(chan int)

	go func() {
		defer s.deregister(r)

		for range progress {
		}
	}()

	resp := response.Stream(
		newEventReader(events),
		response.WithFlush(),
		response.WithProgressChan(progress),
	)

	resp.AddHeader("Cache-Control", "no-cache")
	resp.AddHeader("Connection", "keep-alive")
	resp.AddHeader("Content-Type", "text/event-stream")

	return resp
}

//
// Helpers

func (s *Server) register(r *http.Request) <-chan []byte {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	ch := make(chan []byte, s.bufferSize)
	s.requests[r] = ch
	return ch
}

func (s *Server) deregister(r *http.Request) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if ch, ok := s.requests[r]; ok {
		close(ch)
		delete(s.requests, r)
	}
}

func (s *Server) deregisterAll() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for r, ch := range s.requests {
		close(ch)
		delete(s.requests, r)
	}
}

func (s *Server) publish(data []byte) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	for _, ch := range s.requests {
		select {
		case ch <- data:

		default:
		}
	}
}
