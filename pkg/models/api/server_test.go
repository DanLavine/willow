package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DanLavine/willow/internal/helpers"
	"github.com/DanLavine/willow/pkg/models/api/common/errors"
	dbdefinition "github.com/DanLavine/willow/pkg/models/api/common/v1/db_definition"
	v1locker "github.com/DanLavine/willow/pkg/models/api/locker/v1"
	"github.com/DanLavine/willow/pkg/models/datatypes"

	. "github.com/onsi/gomega"
)

type readError struct{}

func (readError readError) Read(b []byte) (int, error) { return 0, fmt.Errorf("test read error") }
func (readError readError) Close() error               { return nil }

type writeError struct {
	Headers http.Header
	Code    int
}

func (writeError *writeError) Header() http.Header        { return writeError.Headers }
func (writeError *writeError) WriteHeader(statusCode int) { writeError.Code = statusCode }
func (writeError *writeError) Write(p []byte) (n int, err error) {
	return 0, fmt.Errorf("failed test write")
}

type jsonError struct{}

func (je *jsonError) Validate() *errors.ModelError { return nil }
func (je *jsonError) MarshalJSON() ([]byte, error) { return nil, fmt.Errorf("failed to marshal json") }

func Test_DecodeRequest(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It returns an error if the requst body cannot be read", func(t *testing.T) {
		testRequest := httptest.NewRequest("GET", "/", readError{})

		// DecodeHttpRequest
		lock := &v1locker.Lock{}
		serverErr := DecodeRequest(testRequest, lock)

		// check the server response
		g.Expect(serverErr).To(HaveOccurred())
		g.Expect(serverErr.Error()).To(Equal("failed to decode request: test read error"))
	})

	t.Run("It can decode an item successfully", func(t *testing.T) {
		data, err := json.Marshal(&v1locker.Lock{
			Spec: &v1locker.LockSpec{
				DBDefinition: &v1locker.LockDBDefinition{
					KeyValues: dbdefinition.TypedKeyValues{
						"key1": datatypes.Int(1),
					},
				},
				Properties: &v1locker.LockProperties{
					Timeout: helpers.PointerOf(time.Second),
				},
			},
		})
		g.Expect(err).ToNot(HaveOccurred())

		testRequest := httptest.NewRequest("GET", "/", bytes.NewBuffer(data))
		testRequest.Header.Set("Content-Type", "application/json")

		// DecodeHttpRequest
		lockReq := &v1locker.Lock{}
		err = DecodeRequest(testRequest, lockReq)

		// check the server response
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(lockReq.Spec.DBDefinition.KeyValues).To(Equal(dbdefinition.TypedKeyValues{"key1": datatypes.Int(1)}))
		g.Expect(*lockReq.Spec.Properties.Timeout).To(Equal(time.Second))
	})
}

func Test_ModelDecodeRequest(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It returns an error if the requst body cannot be read", func(t *testing.T) {
		testRequest := httptest.NewRequest("GET", "/", readError{})

		// DecodeHttpRequest
		lock := &v1locker.Lock{}
		serverErr := ModelDecodeRequest(testRequest, lock)

		// check the server response
		g.Expect(serverErr).To(HaveOccurred())
		g.Expect(serverErr.Error()).To(Equal("failed to decode request: test read error"))
	})

	t.Run("It returns an error if the object is invalid", func(t *testing.T) {
		data, err := json.Marshal(v1locker.Lock{

			Spec: &v1locker.LockSpec{
				DBDefinition: &v1locker.LockDBDefinition{},
				Properties: &v1locker.LockProperties{
					Timeout: helpers.PointerOf(time.Second),
				},
			},
		})
		g.Expect(err).ToNot(HaveOccurred())

		testRequest := httptest.NewRequest("GET", "/", bytes.NewBuffer(data))

		// DecodeHttpRequest
		lockReq := &v1locker.Lock{}
		err = ModelDecodeRequest(testRequest, lockReq)

		// check the server response
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("failed validation:"))
	})

	t.Run("It can decode an item successfully", func(t *testing.T) {
		data, err := json.Marshal(&v1locker.LockClaim{
			SessionID: "something",
		})
		g.Expect(err).ToNot(HaveOccurred())

		testRequest := httptest.NewRequest("GET", "/", bytes.NewBuffer(data))

		// DecodeHttpRequest
		lockReq := &v1locker.LockClaim{}
		err = ModelDecodeRequest(testRequest, lockReq)

		// check the server response
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(lockReq.SessionID).To(Equal("something"))
	})
}

func Test_ObjectDecodeRequest(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It returns an error if the requst body cannot be read", func(t *testing.T) {
		testRequest := httptest.NewRequest("GET", "/", readError{})

		// DecodeHttpRequest
		lock := &v1locker.Lock{}
		serverErr := ObjectDecodeRequest(testRequest, lock)

		// check the server response
		g.Expect(serverErr).To(HaveOccurred())
		g.Expect(serverErr.Error()).To(Equal("failed to decode request: test read error"))
	})

	t.Run("It returns an error if the object is invalid", func(t *testing.T) {
		data, err := json.Marshal(&v1locker.Lock{
			Spec: &v1locker.LockSpec{
				DBDefinition: &v1locker.LockDBDefinition{},
				Properties: &v1locker.LockProperties{
					Timeout: helpers.PointerOf(time.Second),
				},
			},
		})
		g.Expect(err).ToNot(HaveOccurred())

		testRequest := httptest.NewRequest("GET", "/", bytes.NewBuffer(data))
		testRequest.Header.Set("Content-Type", "application/json")

		// DecodeHttpRequest
		lockReq := &v1locker.Lock{}
		err = ObjectDecodeRequest(testRequest, lockReq)

		// check the server response
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("failed validation:"))
	})

	t.Run("It can decode an item successfully", func(t *testing.T) {
		data, err := json.Marshal(v1locker.Lock{
			Spec: &v1locker.LockSpec{
				DBDefinition: &v1locker.LockDBDefinition{
					KeyValues: dbdefinition.TypedKeyValues{
						"key1": datatypes.Int(1),
					},
				},
				Properties: &v1locker.LockProperties{
					Timeout: helpers.PointerOf(time.Second),
				},
			},
		})
		g.Expect(err).ToNot(HaveOccurred())

		testRequest := httptest.NewRequest("GET", "/", bytes.NewBuffer(data))
		testRequest.Header.Set("Content-Type", "application/json")

		// DecodeHttpRequest
		lockReq := &v1locker.Lock{}
		err = ObjectDecodeRequest(testRequest, lockReq)

		// check the server response
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(lockReq.Spec.DBDefinition.KeyValues).To(Equal(dbdefinition.TypedKeyValues{"key1": datatypes.Int(1)}))
		g.Expect(*lockReq.Spec.Properties.Timeout).To(Equal(time.Second))
	})
}

func Test_ModelEncodeResponse(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("Context when the APIObject is nil", func(t *testing.T) {
		t.Run("It writes the status code if the api object is nil", func(t *testing.T) {
			testWriter := httptest.NewRecorder()

			// Encode and send the http request
			dataWritten, serverErr := ModelEncodeResponse(testWriter, http.StatusOK, nil)
			g.Expect(dataWritten).To(Equal(0))
			g.Expect(serverErr).ToNot(HaveOccurred())

			g.Expect(testWriter.Code).To(Equal(http.StatusOK))
		})
	})

	t.Run("Context when the APIObject is set", func(t *testing.T) {
		t.Run("It returns an error if the model is invalid", func(t *testing.T) {
			testWriter := httptest.NewRecorder()
			lockCreateResp := &v1locker.Lock{}

			// Encode and send the http request
			dataWritten, serverErr := ModelEncodeResponse(testWriter, http.StatusOK, lockCreateResp)
			g.Expect(dataWritten).To(Equal(0))
			g.Expect(serverErr).To(HaveOccurred())
			g.Expect(serverErr.Error()).To(ContainSubstring("failed to validate response object"))

			g.Expect(testWriter.Code).To(Equal(http.StatusInternalServerError))
		})

		t.Run("It returns an error if the encoder fails to encode the object", func(t *testing.T) {
			testWriter := httptest.NewRecorder()

			// Encode and send the http request
			dataWritten, serverErr := ModelEncodeResponse(testWriter, http.StatusOK, &jsonError{})
			g.Expect(dataWritten).To(Equal(0))
			g.Expect(serverErr).To(HaveOccurred())
			g.Expect(serverErr.Error()).To(ContainSubstring("failed to encode the response object"))

			g.Expect(testWriter.Code).To(Equal(http.StatusInternalServerError))
		})

		t.Run("It returns an error if writing to the client fails", func(t *testing.T) {
			testWriter := &writeError{Headers: http.Header{}}

			lockCreateResp := &v1locker.Lock{
				Spec: &v1locker.LockSpec{
					DBDefinition: &v1locker.LockDBDefinition{
						KeyValues: dbdefinition.TypedKeyValues{
							"key1": datatypes.Int(1),
						},
					},
					Properties: &v1locker.LockProperties{
						Timeout: helpers.PointerOf(time.Second),
					},
				},
				State: &v1locker.LockState{
					LockID:             "something",
					SessionID:          "other",
					TimeTillExipre:     time.Second,
					LocksHeldOrWaiting: 1,
				},
			}

			// Encode and send the http request
			dataWritten, serverErr := ModelEncodeResponse(testWriter, http.StatusOK, lockCreateResp)
			g.Expect(dataWritten).To(Equal(0))
			g.Expect(serverErr).To(HaveOccurred())
			g.Expect(serverErr.Error()).To(ContainSubstring("failed to write the response to the client: failed test write"))

			// ensure response recieves proper status code
			g.Expect(testWriter.Code).To(Equal(http.StatusOK))
		})

		t.Run("It write a valid model", func(t *testing.T) {
			testWriter := httptest.NewRecorder()

			lockCreateResp := &v1locker.Lock{
				Spec: &v1locker.LockSpec{
					DBDefinition: &v1locker.LockDBDefinition{
						KeyValues: dbdefinition.TypedKeyValues{
							"key1": datatypes.Int(1),
						},
					},
				},
				State: &v1locker.LockState{
					LockID:             "something",
					SessionID:          "other",
					TimeTillExipre:     time.Second,
					LocksHeldOrWaiting: 1,
				},
			}

			// Encode and send the http request
			dataWritten, serverErr := ModelEncodeResponse(testWriter, http.StatusOK, lockCreateResp)
			g.Expect(dataWritten).To(Equal(173))
			g.Expect(serverErr).ToNot(HaveOccurred())

			// ensure response recieves proper status code
			g.Expect(testWriter.Code).To(Equal(http.StatusOK))

			// ensure response can be decode properly
			writeData, err := io.ReadAll(testWriter.Result().Body)
			g.Expect(err).ToNot(HaveOccurred())

			parseLockResp := &v1locker.Lock{}
			err = json.Unmarshal(writeData, parseLockResp)

			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(parseLockResp).To(Equal(lockCreateResp))
		})
	})
}
