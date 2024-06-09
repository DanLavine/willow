package api

import (
	"bytes"
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/DanLavine/willow/pkg/models/api/common/errors"
	dbdefinition "github.com/DanLavine/willow/pkg/models/api/common/v1/db_definition"
	v1locker "github.com/DanLavine/willow/pkg/models/api/locker/v1"
	"github.com/DanLavine/willow/pkg/models/datatypes"

	. "github.com/onsi/gomega"
)

type marshalError struct{}

func (me *marshalError) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("marshal error")
}
func (me *marshalError) Validate() *errors.ModelError         { return nil }
func (me *marshalError) ValidateSpecOnly() *errors.ModelError { return nil }

func Test_ObjectEncodeRequest(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It panics if the object is nil", func(t *testing.T) {
		g.Expect(func() { ObjectEncodeRequest(nil) }).To(Panic())
	})

	t.Run("It returns an error if the response body cannot be read", func(t *testing.T) {
		data, clientErr := ObjectEncodeRequest(&marshalError{})

		g.Expect(data).To(BeNil())
		g.Expect(clientErr).To(HaveOccurred())
		g.Expect(clientErr.Error()).To(Equal("json: error calling MarshalJSON for type *api.marshalError: marshal error"))
	})

	t.Run("It can encode an item successfully", func(t *testing.T) {
		// DecodeAndValidateHttpRequest
		lock := &v1locker.Lock{
			Spec: &v1locker.LockSpec{
				DBDefinition: &v1locker.LockDBDefinition{
					KeyValues: dbdefinition.TypedKeyValues{
						"key1": datatypes.String("1"),
					},
				},
			},
		}

		data, err := ObjectEncodeRequest(lock)

		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(string(data)).To(Equal(`{"Spec":{"DBDefinition":{"KeyValues":{"key1":{"Type":13,"Data":"1"}}}}}`))
	})
}

func Test_ModelEncodeRequest(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It panics if the object is nil", func(t *testing.T) {
		g.Expect(func() { ModelEncodeRequest(nil) }).To(Panic())
	})

	t.Run("It returns an error if the response body cannot be read", func(t *testing.T) {
		data, clientErr := ModelEncodeRequest(&marshalError{})

		g.Expect(data).To(BeNil())
		g.Expect(clientErr).To(HaveOccurred())
		g.Expect(clientErr.Error()).To(Equal("json: error calling MarshalJSON for type *api.marshalError: marshal error"))
	})

	t.Run("It can encode an item successfully", func(t *testing.T) {
		// DecodeAndValidateHttpRequest
		lock := &v1locker.LockClaim{
			SessionID: "something",
		}

		data, err := ModelEncodeRequest(lock)

		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(string(data)).To(Equal(`{"SessionID":"something"}`))
	})
}

func Test_ModelDecodeResponse(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It panic if the resp is nil", func(t *testing.T) {
		g.Expect(func() { ModelDecodeResponse(nil, &jsonError{}) }).To(Panic())
	})

	t.Run("It returns an error if the response body cannot be read", func(t *testing.T) {
		testResponse := httptest.NewRecorder()
		testResponse.Body = bytes.NewBuffer([]byte(`{bad json`))

		err := ModelDecodeResponse(testResponse.Result(), &v1locker.LockClaim{})

		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("failed to decode response: invalid character 'b' looking for beginning of object key string"))
	})

	t.Run("It returns an error if response is invalid", func(t *testing.T) {
		testResponse := httptest.NewRecorder()
		testResponse.Body = bytes.NewBuffer([]byte(`{"SessionID":""}`))

		err := ModelDecodeResponse(testResponse.Result(), &v1locker.LockClaim{})

		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("SessionID: received an empty string"))
	})

	t.Run("It can process a valid response", func(t *testing.T) {
		testResponse := httptest.NewRecorder()
		testResponse.Body = bytes.NewBuffer([]byte(`{"SessionID":"id"}`))

		lockClaim := &v1locker.LockClaim{}
		err := ModelDecodeResponse(testResponse.Result(), lockClaim)

		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(lockClaim.SessionID).To(Equal("id"))
	})
}
