package v1

import (
	"bytes"
	"io/ioutil"
	"testing"

	. "github.com/onsi/gomega"
)

func TestQuery_ParseQuery(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("it returns an error on an invalid json syntax", func(t *testing.T) {
		buffer := ioutil.NopCloser(bytes.NewBufferString(`{"BrokerName":"test"`))

		query, err := ParseQuery(buffer)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("Failed to parse request body"))
		g.Expect(query).To(BeNil())
	})
}

func TestQuery_validate(t *testing.T) {
	g := NewGomegaWithT(t)

	setup := func() *Query {
		return &Query{
			BrokerName: "test",
		}
	}

	t.Run("it returns an error when the BrokerName is not set", func(t *testing.T) {
		query := setup()
		query.BrokerName = ""

		err := query.validate()
		g.Expect(err).ToNot(BeNil())
		g.Expect(err.Error()).To(ContainSubstring("BrokerName cannot be empty"))
	})
}
