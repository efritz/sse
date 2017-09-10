package sse

import (
	"net/http"
	"sync"
)

// Server is an http.Handler that will fanout Server-Sent Events
// to all connected clients.
type Server struct {
	events     <-chan interface{}
	requests   map[*http.Request]chan []byte
	mutex      *sync.RWMutex
	bufferSize int
}

// NewServer creates a new server with the given event channel.
// The server returned by this function has not yet started.
func NewServer(events <-chan interface{}) *Server {
	return &Server{
		events:     events,
		requests:   map[*http.Request]chan []byte{},
		mutex:      &sync.RWMutex{},
		bufferSize: 100,
	}
}

// ServeHTTP will begin sending events to a new client. This handler
// exits if either the client disconnects or if the server's event
// channel is closed. This endpoint will return an error if streaming
// is not supported on the client.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	notify := w.(http.CloseNotifier).CloseNotify()
	w.Header().Add("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ch := s.register(r)
	defer s.deregister(r)

	for {
		select {
		case <-notify:
			return

		case data := <-ch:
			if _, err := w.Write(data); err != nil {
				return
			}

			f.Flush()
		}
	}
}

// Start will begin serializing events that come in on the
// event channel and sending the payload to each registered
// client. This routine stops once the server's event channel
// is closed.
func (s *Server) Start() {
	go func() {
		for event := range s.events {
			data, err := serializeSSE(event)
			if err != nil {
				return // TODO
			}

			s.publish(data)
		}

		s.deregisterAll()
	}()
}

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
