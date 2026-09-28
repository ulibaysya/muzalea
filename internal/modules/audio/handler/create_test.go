package handler

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	pkghttp "github.com/ulibaysya/muzalea/pkg/http"
)

// type errReader struct{}
//
// func (r errReader) Read(_ []byte) (int, error) {
// 	return 0, fmt.Errorf("some error")
// }

type mockService struct {
	idToReturn  int64
	errToReturn error
	calls       int
}

func (m *mockService) SaveAudio(ctx context.Context, name, author string, tags []string, audio io.Reader) (int64, error) {
	m.calls += 1
	return m.idToReturn, m.errToReturn
}

type partOpt struct {
	fieldName string
	body      io.Reader
}

func newMultipartRequest(t *testing.T, partsOpts ...partOpt) *http.Request {
	t.Helper()

	body := bytes.NewBuffer(nil)
	writer := multipart.NewWriter(body)

	for _, i := range partsOpts {
		fieldWriter, err := writer.CreateFormField(i.fieldName)
		if err != nil {
			t.Fatal(err)
		}
		if i.body != nil {
			if _, err = io.Copy(fieldWriter, i.body); err != nil {
				t.Fatal(err)
			}
		}
	}

	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodPost, "/audio", body)
	req.Header.Set(pkghttp.HeaderContentType, writer.FormDataContentType())

	return req
}

func TestCreate(t *testing.T) {
	tests := []struct {
		name        string
		request     func(t *testing.T) *http.Request
		serviceFunc func(service *mockService)

		code         int
		location     string
		serviceCalls int
	}{
		{
			name: "NotMultipart",
			code: http.StatusBadRequest,
			request: func(t *testing.T) *http.Request {
				return httptest.NewRequest(http.MethodPost, "/audio", http.NoBody)
			},
		},
		{
			name: "MissingMetaPart",
			code: http.StatusBadRequest,
			request: func(t *testing.T) *http.Request {
				return newMultipartRequest(t)
			},
		},
		{
			name: "InvalidMetaFormName",
			code: http.StatusBadRequest,
			request: func(t *testing.T) *http.Request {
				return newMultipartRequest(t, []partOpt{
					{
						fieldName: "not-meta",
					},
				}...)
			},
		},
		{
			name: "MetaUnmarshalError",
			code: http.StatusBadRequest,
			request: func(t *testing.T) *http.Request {
				return newMultipartRequest(t, []partOpt{
					{
						fieldName: "meta",
						body:      strings.NewReader("not json at all"),
					},
				}...)
			},
		},
		{
			name: "ZeroMeta",
			code: http.StatusBadRequest,
			request: func(t *testing.T) *http.Request {
				return newMultipartRequest(t, []partOpt{
					{
						fieldName: "meta",
						body:      strings.NewReader("{}"),
					},
				}...)
			},
		},
		{
			name: "MissingDataPart",
			code: http.StatusBadRequest,
			request: func(t *testing.T) *http.Request {
				return newMultipartRequest(t, []partOpt{
					{
						fieldName: "meta",
						body:      strings.NewReader(`{"author":"some-name","name":"some-name","tags":["tag1", "tag2"]}`),
					},
				}...)
			},
		},
		{
			name: "InvalidDataFormName",
			code: http.StatusBadRequest,
			request: func(t *testing.T) *http.Request {
				return newMultipartRequest(t, []partOpt{
					{
						fieldName: "meta",
						body:      strings.NewReader(`{"author":"some-name","name":"some-name","tags":["tag1", "tag2"]}`),
					},
					{
						fieldName: "not-data",
					},
				}...)
			},
		},
		// {
		// 	name: "CopyingDataError",
		// 	code: http.StatusBadRequest,
		// 	request: func(t *testing.T) *http.Request {
		// 		body := bytes.NewBuffer(nil)
		// 		multipart.NewWriter(body).WriteField
		// 		httptest.NewRequest(http.MethodPost, "/audio", body)
		// 		// return newMultipartRequest(t, []partOpt{
		// 		// 	{
		// 		// 		fieldName: "meta",
		// 		// 		body:      validMetaBody,
		// 		// 	},
		// 		// 	{
		// 		// 		fieldName: "data",
		// 		// 		body:      errReader{},
		// 		// 	},
		// 		// }...)
		// 		return nil
		// 	},
		// },
		{
			name: "DataExceedsLimit",
			code: http.StatusRequestEntityTooLarge,
			request: func(t *testing.T) *http.Request {
				return newMultipartRequest(t, []partOpt{
					{
						fieldName: "meta",
						body:      strings.NewReader(`{"author":"some-name","name":"some-name","tags":["tag1", "tag2"]}`),
					},
					{
						fieldName: "data",
						body:      bytes.NewBuffer(make([]byte, (20<<20)+1)),
					},
				}...)
			},
		},
		{
			name: "UnexpectedAdditionalPart",
			code: http.StatusBadRequest,
			request: func(t *testing.T) *http.Request {
				return newMultipartRequest(t, []partOpt{
					{
						fieldName: "meta",
						body:      strings.NewReader(`{"author":"some-name","name":"some-name","tags":["tag1", "tag2"]}`),
					},
					{
						fieldName: "data",
						body:      bytes.NewReader([]byte("some-audio")),
					},
					{
						fieldName: "unexpected",
					},
				}...)
			},
		},
		// TODO check that handler passes correct arguments (author, name, tags, audio)
		{
			name: "SaveAudioError",
			code: http.StatusBadRequest,
			request: func(t *testing.T) *http.Request {
				return newMultipartRequest(t, []partOpt{
					{
						fieldName: "meta",
						body:      strings.NewReader(`{"author":"some-name","name":"some-name","tags":["tag1", "tag2"]}`),
					},
					{
						fieldName: "data",
						body:      bytes.NewReader([]byte("some-audio")),
					},
				}...)
			},
			serviceCalls: 1,
			serviceFunc: func(service *mockService) {
				service.errToReturn = fmt.Errorf("some error")
			},
		},
		{
			name: "Success",
			code: http.StatusCreated,
			request: func(t *testing.T) *http.Request {
				return newMultipartRequest(t, []partOpt{
					{
						fieldName: "meta",
						body:      strings.NewReader(`{"author":"some-name","name":"some-name","tags":["tag1", "tag2"]}`),
					},
					{
						fieldName: "data",
						body:      bytes.NewReader([]byte("some-audio")),
					},
				}...)
			},
			serviceCalls: 1,
			serviceFunc: func(service *mockService) {
				service.idToReturn = 123
			},
			location: "/audio/123",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := &mockService{}
			rec := httptest.NewRecorder()
			req := test.request(t)
			if test.serviceFunc != nil {
				test.serviceFunc(service)
			}

			New(slog.New(slog.NewTextHandler(t.Output(), nil)), service).Create(rec, req)

			if rec.Code != test.code {
				t.Fatalf("unexpected code %d, want %d", rec.Code, test.code)
			}

			if location := rec.Header().Get(pkghttp.HeaderLocation); location != test.location {
				t.Errorf("unexpected Location %s, want %s", location, test.location)
			}

			if test.serviceCalls != service.calls {
				t.Errorf("service called %v times, want %v", service.calls, test.serviceCalls)
			}
		})
	}
}
