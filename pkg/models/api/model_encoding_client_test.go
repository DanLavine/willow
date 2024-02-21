package api

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	v1locker "github.com/DanLavine/willow/pkg/models/api/locker/v1"

	. "github.com/onsi/gomega"
)

func setupBody(g *GomegaWithT) io.ReadCloser {
	lockCreateResponse := &v1locker.LockCreateResponse{
		SessionID:   "some id",
		LockTimeout: time.Second,
	}

	data, err := lockCreateResponse.EncodeJSON()
	g.Expect(err).ToNot(HaveOccurred())

	return io.NopCloser(bytes.NewBuffer(data))
}

func Test_DecodeAndValidateHttpResponse(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It panic if the object is nil", func(t *testing.T) {
		testResponse := &http.Response{}

		// DecodeAndValidateHttpRequest
		g.Expect(func() { DecodeAndValidateHttpResponse(testResponse, nil) }).To(Panic())
	})

	t.Run("It returns an error if the response body cannot be read", func(t *testing.T) {
		testResponse := &http.Response{}
		testResponse.Body = &readError{}

		// DecodeAndValidateHttpRequest
		lockCreateResponse := &v1locker.LockCreateResponse{}
		clientErr := DecodeAndValidateHttpResponse(testResponse, lockCreateResponse)

		// check the server response
		g.Expect(clientErr).To(HaveOccurred())
		g.Expect(clientErr.Error()).To(Equal("failed to read http response body: test read error"))
	})

	t.Run("Context Content-Type headers", func(t *testing.T) {
		t.Run("It returns an error if the type is unkown", func(t *testing.T) {
			testResponse := &http.Response{}
			testResponse.Body = setupBody(g)
			testResponse.Header = http.Header{}
			testResponse.Header.Add("Content-Type", "bad type")

			// DecodeAndValidateHttpRequest
			lockCreateResponse := &v1locker.LockCreateResponse{}
			clientErr := DecodeAndValidateHttpResponse(testResponse, lockCreateResponse)

			// check the server response
			g.Expect(clientErr).To(HaveOccurred())
			g.Expect(clientErr.Error()).To(Equal("unkown content type recieved from service: bad type"))
		})

		t.Run("Context JSON", func(t *testing.T) {
			t.Run("It defaults to json if no content type header is provided", func(t *testing.T) {
				testResponse := &http.Response{}
				testResponse.Body = setupBody(g)
				testResponse.Header = http.Header{}

				// DecodeAndValidateHttpRequest
				lockCreateResponse := &v1locker.LockCreateResponse{}
				clientErr := DecodeAndValidateHttpResponse(testResponse, lockCreateResponse)

				// check the server response
				g.Expect(clientErr).ToNot(HaveOccurred())
				g.Expect(lockCreateResponse.SessionID).To(Equal("some id"))
				g.Expect(lockCreateResponse.LockTimeout).To(Equal(time.Second))
			})

			t.Run("It can decode an item successfully", func(t *testing.T) {
				testResponse := &http.Response{}
				testResponse.Body = setupBody(g)
				testResponse.Header = http.Header{}
				testResponse.Header.Add("content-type", "application/json")

				// DecodeAndValidateHttpRequest
				lockCreateResponse := &v1locker.LockCreateResponse{}
				clientErr := DecodeAndValidateHttpResponse(testResponse, lockCreateResponse)

				// check the server response
				g.Expect(clientErr).ToNot(HaveOccurred())
				g.Expect(lockCreateResponse.SessionID).To(Equal("some id"))
				g.Expect(lockCreateResponse.LockTimeout).To(Equal(time.Second))
			})

			t.Run("It returns an error if data cannot be decoded", func(t *testing.T) {
				testResponse := &http.Response{}
				testResponse.Body = io.NopCloser(bytes.NewBuffer([]byte(`not json`)))
				testResponse.Header = http.Header{}
				testResponse.Header.Add("content-type", "application/json")

				// DecodeAndValidateHttpRequest
				lockCreateResponse := &v1locker.LockCreateResponse{}
				clientErr := DecodeAndValidateHttpResponse(testResponse, lockCreateResponse)

				// check the server response
				g.Expect(clientErr).To(HaveOccurred())
				g.Expect(clientErr.Error()).To(ContainSubstring("failed to decode response"))
			})
		})

		t.Run("It returns an error if the data is not valid", func(t *testing.T) {
			testResponse := &http.Response{}
			testResponse.Body = io.NopCloser(bytes.NewBuffer([]byte(`{"SessionID":""}`)))
			testResponse.Header = http.Header{}
			testResponse.Header.Add("content-type", "application/json")

			// DecodeAndValidateHttpRequest
			lockCreateResponse := &v1locker.LockCreateResponse{}
			clientErr := DecodeAndValidateHttpResponse(testResponse, lockCreateResponse)

			// check the server response
			g.Expect(clientErr).To(HaveOccurred())
			fmt.Println(clientErr.Error())
			g.Expect(clientErr.Error()).To(Equal("failed validation for api response: 'SessionID' is the empty string"))
		})
	})
}
