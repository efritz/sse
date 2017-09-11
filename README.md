# Nacelle

[![GoDoc](https://godoc.org/github.com/efritz/sse?status.svg)](https://godoc.org/github.com/efritz/sse)
[![Build Status](https://secure.travis-ci.org/efritz/sse.png)](http://travis-ci.org/efritz/sse)
[![codecov.io](http://codecov.io/github/efritz/sse/coverage.svg?branch=master)](http://codecov.io/github/efritz/sse?branch=master)

Go server for announcing a stream of [Server-Sent Events](https://en.wikipedia.org/wiki/Server-sent_events).

## Example

An SSE announce server implements `http.Handler` so it can be registered to
an existing router in your application. It takes a read-only channel of
interface objects to send to all connected clients. Once a client connects
they will begin receiving all events that occur *at that point* until they
disconnect. Closing the event channel will shutdown the server's background
routines.

```go
events := make(chan interface{})
server := NewServer(events, WithBufferSize(50))

go func() {
    if err := server.Start(); err != nil {
        panic(err.Error())
    }
}

go func() {
    defer close(events)

    events <- map[string]int{"foo": 1}
    events <- map[string]int{"bar": 2}
    events <- map[string]int{"baz": 3}
}

http.Handle("/events", server)
http.ListenAndServe("0.0.0.0:8080", nil)
```

## License

Copyright (c) 2017 Eric Fritz

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
