package api

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	v1locker "github.com/DanLavine/willow/pkg/models/api/locker/v1"
	"github.com/DanLavine/willow/pkg/models/datatypes"

	. "github.com/onsi/gomega"
)

type readError struct{}

func (readError readError) Read(b []byte) (int, error) {
	return 0, fmt.Errorf("test read error")
}

func (readError readError) Close() error { return nil }

type writeError struct {
	Code int
}

func (writeError *writeError) Header() http.Header        { return http.Header{} }
func (writeError *writeError) WriteHeader(statusCode int) { writeError.Code = statusCode }
func (writeError *writeError) Write(p []byte) (n int, err error) {
	return 0, fmt.Errorf("failed test write")
}

type jsonError struct{}

func (je *jsonError) Validate() error             { return nil }
func (je *jsonError) DecodeJSON(b []byte) error   { return nil }
func (je *jsonError) EncodeJSON() ([]byte, error) { return nil, fmt.Errorf("encode test error") }

func Test_DecodeAndValidateHttpRequest(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It returns an error if the requst body cannot be read", func(t *testing.T) {
		testRequest := httptest.NewRequest("GET", "/", readError{})

		// DecodeAndValidateHttpRequest
		lock := &v1locker.Lock{}
		serverErr := DecodeAndValidateHttpRequest(testRequest, lock)

		// check the server response
		g.Expect(serverErr).To(HaveOccurred())
		g.Expect(serverErr.Error()).To(Equal("failed to read http request body: test read error"))
	})

	t.Run("It returns an error if the model is nil", func(t *testing.T) {
		testRequest := httptest.NewRequest("GET", "/", nil)

		// DecodeAndValidateHttpRequest
		serverErr := DecodeAndValidateHttpRequest(testRequest, nil)

		// check the server response
		g.Expect(serverErr).To(HaveOccurred())
		g.Expect(serverErr.Error()).To(Equal("unable to decode api model"))
	})

	t.Run("Context Content-Type headers", func(t *testing.T) {
		t.Run("It returns an error if the type is unkown", func(t *testing.T) {
			testRequest := httptest.NewRequest("GET", "/", nil)
			testRequest.Header.Set("Content-Type", "something bad")

			// DecodeAndValidateHttpRequest
			lock := &v1locker.Lock{}
			serverErr := DecodeAndValidateHttpRequest(testRequest, lock)

			// check the server response
			g.Expect(serverErr).To(HaveOccurred())
			g.Expect(serverErr.Error()).To(Equal("server recieved unkown content type: 'something bad'"))
		})

		t.Run("Context JSON", func(t *testing.T) {
			t.Run("It returns an error if request is not json", func(t *testing.T) {
				testRequest := httptest.NewRequest("GET", "/", bytes.NewBuffer([]byte(`so not json`)))
				testRequest.Header.Set("Content-Type", "application/json")

				// DecodeAndValidateHttpRequest
				lock := &v1locker.Lock{}
				serverErr := DecodeAndValidateHttpRequest(testRequest, lock)

				// check the server response
				g.Expect(serverErr).To(HaveOccurred())
				g.Expect(serverErr.Error()).To(ContainSubstring("failed to decode request:"))
			})

			t.Run("It returns an error if requst data is not valid", func(t *testing.T) {
				testRequest := httptest.NewRequest("GET", "/", bytes.NewBuffer([]byte(`{}`)))
				testRequest.Header.Set("Content-Type", "application/json")

				// DecodeAndValidateHttpRequest
				lockCreateRequest := &v1locker.LockCreateRequest{}
				serverErr := DecodeAndValidateHttpRequest(testRequest, lockCreateRequest)

				// check the server response
				g.Expect(serverErr).To(HaveOccurred())
				g.Expect(serverErr.Error()).To(Equal("failed validation: 'KeValues' is empty, but requires a length of at least 1"))
			})

			t.Run("It can decode an item successfully", func(t *testing.T) {
				data, err := (&v1locker.LockCreateRequest{
					KeyValues: datatypes.KeyValues{
						"key1": datatypes.Int(1),
					},
					LockTimeout: time.Second,
				}).EncodeJSON()
				g.Expect(err).ToNot(HaveOccurred())

				testRequest := httptest.NewRequest("GET", "/", bytes.NewBuffer(data))
				testRequest.Header.Set("Content-Type", "application/json")

				// DecodeAndValidateHttpRequest
				lockCreateRequest := &v1locker.LockCreateRequest{}
				serverErr := DecodeAndValidateHttpRequest(testRequest, lockCreateRequest)

				// check the server response
				g.Expect(serverErr).ToNot(HaveOccurred())
				g.Expect(lockCreateRequest.KeyValues).To(Equal(datatypes.KeyValues{"key1": datatypes.Int(1)}))
				g.Expect(lockCreateRequest.LockTimeout).To(Equal(time.Second))
			})
		})
	})
}

func Test_EncodeAndSendHttpResponse(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("Context when the APIObject is nil", func(t *testing.T) {
		t.Run("It writes the status code if the api object is nil", func(t *testing.T) {
			testWriter := httptest.NewRecorder()

			// Encode and send the http request
			dataWritten, serverErr := EncodeAndSendHttpResponse(http.Header{}, testWriter, http.StatusOK, nil)
			g.Expect(dataWritten).To(Equal(0))
			g.Expect(serverErr).ToNot(HaveOccurred())

			g.Expect(testWriter.Code).To(Equal(http.StatusOK))
		})
	})

	t.Run("Context when the APIObject is set", func(t *testing.T) {
		t.Run("It returns an error if the model is invalid", func(t *testing.T) {
			testWriter := httptest.NewRecorder()

			lockCreateResp := &v1locker.LockCreateResponse{}

			// Encode and send the http request
			dataWritten, serverErr := EncodeAndSendHttpResponse(http.Header{}, testWriter, http.StatusOK, lockCreateResp)
			g.Expect(dataWritten).To(Equal(0))
			g.Expect(serverErr).To(HaveOccurred())
			g.Expect(serverErr.Error()).To(ContainSubstring("failed to validate response object"))

			g.Expect(testWriter.Code).To(Equal(http.StatusInternalServerError))
		})

		t.Run("It returns an error on a if write fails", func(t *testing.T) {
			testWriter := &writeError{}

			lockCreateResp := &v1locker.LockCreateResponse{
				SessionID:   "something",
				LockTimeout: 5 * time.Second,
			}

			headers := http.Header{}

			// Encode and send the http request
			dataWritten, serverErr := EncodeAndSendHttpResponse(headers, testWriter, http.StatusOK, lockCreateResp)
			g.Expect(dataWritten).To(Equal(0))
			g.Expect(serverErr).To(HaveOccurred())
			g.Expect(serverErr.Error()).To(Equal("failed to write the respoonse to the client. Closed unexpectedly: failed test write"))

			// ensure response recieves proper status code
			g.Expect(testWriter.Code).To(Equal(http.StatusOK))
		})

		t.Run("Context Content-Type headers", func(t *testing.T) {
			t.Run("It returns an error if encoding the object fails", func(t *testing.T) {
				testWriter := httptest.NewRecorder()

				// Encode and send the http request
				dataWritten, serverErr := EncodeAndSendHttpResponse(http.Header{}, testWriter, http.StatusOK, &jsonError{})
				g.Expect(dataWritten).To(Equal(0))
				g.Expect(serverErr).To(HaveOccurred())
				g.Expect(serverErr.Error()).To(ContainSubstring("failed to encode the response object"))

				// ensure response recieves proper status code
				g.Expect(testWriter.Code).To(Equal(http.StatusInternalServerError))
			})

			t.Run("It sets the content type header to application/json if non are provided", func(t *testing.T) {
				testWriter := httptest.NewRecorder()

				lockCreateResp := &v1locker.LockCreateResponse{
					SessionID:   "something",
					LockTimeout: 5 * time.Second,
				}

				// Encode and send the http request
				dataWritten, serverErr := EncodeAndSendHttpResponse(http.Header{}, testWriter, http.StatusOK, lockCreateResp)
				g.Expect(dataWritten).To(Equal(50))
				g.Expect(serverErr).ToNot(HaveOccurred())

				// ensure response recieves proper status code
				g.Expect(testWriter.Code).To(Equal(http.StatusOK))

				// ensure response can be decode properly
				data, err := io.ReadAll(testWriter.Body)
				g.Expect(err).ToNot(HaveOccurred())

				parseLockResp := &v1locker.LockCreateResponse{}
				err = parseLockResp.DecodeJSON(data)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(parseLockResp).To(Equal(lockCreateResp))
			})

			t.Run("It can use application/json", func(t *testing.T) {
				testWriter := httptest.NewRecorder()

				lockCreateResp := &v1locker.LockCreateResponse{
					SessionID:   "something",
					LockTimeout: 5 * time.Second,
				}

				headers := http.Header{}
				headers.Add("Content-Type", "application/json")

				// Encode and send the http request
				dataWritten, serverErr := EncodeAndSendHttpResponse(headers, testWriter, http.StatusOK, lockCreateResp)
				g.Expect(dataWritten).To(Equal(50))
				g.Expect(serverErr).ToNot(HaveOccurred())

				// ensure response recieves proper status code
				g.Expect(testWriter.Code).To(Equal(http.StatusOK))

				// ensure response can be decode properly
				data, err := io.ReadAll(testWriter.Body)
				g.Expect(err).ToNot(HaveOccurred())

				parseLockResp := &v1locker.LockCreateResponse{}
				err = parseLockResp.DecodeJSON(data)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(parseLockResp).To(Equal(lockCreateResp))
			})

			t.Run("It returns an error on a bad content type", func(t *testing.T) {
				testWriter := httptest.NewRecorder()

				lockCreateResp := &v1locker.LockCreateResponse{
					SessionID:   "something",
					LockTimeout: 5 * time.Second,
				}

				headers := http.Header{}
				headers.Add("Content-Type", "wops")

				// Encode and send the http request
				dataWritten, serverErr := EncodeAndSendHttpResponse(headers, testWriter, http.StatusOK, lockCreateResp)
				g.Expect(dataWritten).To(Equal(0))
				g.Expect(serverErr).To(HaveOccurred())
				g.Expect(serverErr.Error()).To(Equal("unknown content type to send back to the client: wops"))

				// ensure response recieves proper status code
				g.Expect(testWriter.Code).To(Equal(http.StatusInternalServerError))
			})
		})
	})
}
