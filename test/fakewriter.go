package test

import (
	"net/http"
)

// FakeWriter can be used as a fake in tests where an HTTP writer is required
type FakeWriter struct {
	header http.Header
	Status int
	Bytes  []byte
}

// Header returns a header that has been set on a fake writer
func (t *FakeWriter) Header() http.Header {
	return t.header
}

// Write writes bytes to a fake writer
func (t *FakeWriter) Write(b []byte) (int, error) {
	t.Bytes = b
	return len(b), nil
}

// WriteHeader sets the HTTP status code on a fake writer
func (t *FakeWriter) WriteHeader(status int) {
	t.Status = status
}

// CreateFakeWriter constructs and returns a new fake writer
func CreateFakeWriter() *FakeWriter {
	return &FakeWriter{
		header: make(http.Header),
	}
}
