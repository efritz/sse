package sse

// ConfigFunc is a function used to initialize a new server.
type ConfigFunc func(*Server)

// WithBufferSize sets the internal buffer for each connected client.
// This buffer counts distinct events. The default is 100.
func WithBufferSize(bufferSize int) ConfigFunc {
	return func(s *Server) { s.bufferSize = bufferSize }
}
