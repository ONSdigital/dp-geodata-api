// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package mock

import (
	"context"
	"github.com/ONSdigital/dp-api-clients-go/v2/upload"
	"github.com/ONSdigital/dp-geodata-api/cmd/filescli/app"
	"io"
	"sync"
)

// Ensure, that UploaderMock does implement app.Uploader.
// If this is not the case, regenerate this file with moq.
var _ app.Uploader = &UploaderMock{}

// UploaderMock is a mock implementation of app.Uploader.
//
// 	func TestSomethingThatUsesUploader(t *testing.T) {
//
// 		// make and configure a mocked app.Uploader
// 		mockedUploader := &UploaderMock{
// 			UploadFunc: func(ctx context.Context, f io.ReadCloser, meta upload.Metadata) error {
// 				panic("mock out the Upload method")
// 			},
// 		}
//
// 		// use mockedUploader in code that requires app.Uploader
// 		// and then make assertions.
//
// 	}
type UploaderMock struct {
	// UploadFunc mocks the Upload method.
	UploadFunc func(ctx context.Context, f io.ReadCloser, meta upload.Metadata) error

	// calls tracks calls to the methods.
	calls struct {
		// Upload holds details about calls to the Upload method.
		Upload []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// F is the f argument value.
			F io.ReadCloser
			// Meta is the meta argument value.
			Meta upload.Metadata
		}
	}
	lockUpload sync.RWMutex
}

// Upload calls UploadFunc.
func (mock *UploaderMock) Upload(ctx context.Context, f io.ReadCloser, meta upload.Metadata) error {
	if mock.UploadFunc == nil {
		panic("UploaderMock.UploadFunc: method is nil but Uploader.Upload was just called")
	}
	callInfo := struct {
		Ctx  context.Context
		F    io.ReadCloser
		Meta upload.Metadata
	}{
		Ctx:  ctx,
		F:    f,
		Meta: meta,
	}
	mock.lockUpload.Lock()
	mock.calls.Upload = append(mock.calls.Upload, callInfo)
	mock.lockUpload.Unlock()
	return mock.UploadFunc(ctx, f, meta)
}

// UploadCalls gets all the calls that were made to Upload.
// Check the length with:
//     len(mockedUploader.UploadCalls())
func (mock *UploaderMock) UploadCalls() []struct {
	Ctx  context.Context
	F    io.ReadCloser
	Meta upload.Metadata
} {
	var calls []struct {
		Ctx  context.Context
		F    io.ReadCloser
		Meta upload.Metadata
	}
	mock.lockUpload.RLock()
	calls = mock.calls.Upload
	mock.lockUpload.RUnlock()
	return calls
}
