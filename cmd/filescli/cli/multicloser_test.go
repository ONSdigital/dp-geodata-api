package cli

import (
	"errors"
	"io"
	"strings"
	"testing"
)

// errorCloser is an io.ReadCloser that just reads from an io.Reader
// and does NOT permit multiple calls to Close.
type errorCloser struct {
	numCloses int       // number of times Close has been called
	Err       error     // error to return from Close
	R         io.Reader // underlying reader
}

func (e *errorCloser) Read(p []byte) (int, error) {
	return e.R.Read(p)
}

func (e *errorCloser) Close() error {
	e.numCloses++
	if e.numCloses > 1 {
		return errors.New("errorClose already closed")
	}
	return e.Err
}

func Test_multiCloser(t *testing.T) {
	var tests = map[string]error{
		"error from Close":    errors.New("error from close"),
		"no error from Close": nil,
	}

	contents := "this is the what should be read"
	for name, wantErr := range tests {
		t.Run(name, func(t *testing.T) {
			reader := strings.NewReader(contents)
			ec := &errorCloser{
				Err: wantErr,
				R:   reader,
			}
			mc := &multiCloser{
				RC: ec,
			}

			// test Read is sane
			buf := make([]byte, len(contents))
			n, err := mc.Read(buf)
			if string(buf) != contents {
				t.Errorf("mc.Read: %q, want %q", string(buf), contents)
				return
			}
			if n != len(contents) {
				t.Errorf("mc.Read returned length: %d, want %d", n, len(contents))
				return
			}
			if err != nil {
				t.Errorf("mc.Read %q, want nil", err)
				return
			}

			// first Close
			err = mc.Close()
			if err != wantErr {
				t.Errorf("first mc.Close %q, want %q", err, wantErr)
				return
			}

			// second Close should return original error
			err = mc.Close()
			if err != wantErr {
				t.Errorf("second mc.Close %q, want %q", err, wantErr)
				return
			}

			// underlying ReadCloser must receive a single Close
			if ec.numCloses != 1 {
				t.Errorf("underlying ReadCloser got %d closes, want 1", ec.numCloses)
			}
		})
	}
}
