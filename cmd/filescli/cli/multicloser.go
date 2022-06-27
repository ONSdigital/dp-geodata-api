package cli

import "io"

// multiCloser implements io.ReadCloser but allows multiple calls to Close.
//
// The Upload method in the upload client takes an io.ReadCloser as the source
// for the file to upload, but I don't see that it actually calls .Close
// anywhere.
//
// So the ReadCloser is wrapped in multiCloser before calling Upload, and
// an explicit Close is called once Upload returns, and it is no error to
// call Close twice, just in case something does a close within downstream
// libraries.
type multiCloser struct {
	isClosed bool
	err      error
	RC       io.ReadCloser
}

func (m *multiCloser) Read(p []byte) (int, error) {
	return m.RC.Read(p)
}

func (m *multiCloser) Close() error {
	if m.isClosed {
		return m.err
	}
	m.err = m.RC.Close()
	m.isClosed = true
	return m.err
}
