package v1

import (
	"bytes"
	"io"
	"testing"

	"github.com/DanLavine/willow/pkg/models/api"
	"github.com/DanLavine/willow/pkg/models/datatypes"
	. "github.com/onsi/gomega"
)

func TestV1LimiterModels_Overrides(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("Describe Decode", func(t *testing.T) {
		t.Run("It properly decodes overrides", func(t *testing.T) {
			overrides := &Overrides{}

			err := overrides.Decode(api.ContentTypeJSON, io.NopCloser(bytes.NewBuffer([]byte(`{"Overrides": [{"Name":"test name","KeyValues":{"key1":{"Type":9,"Data":"1"}},"Limit":3}] }`))))
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(len(overrides.Overrides)).To(Equal(1))
			g.Expect(overrides.Overrides[0].Name).To(Equal("test name"))
			g.Expect(overrides.Overrides[0].KeyValues).To(Equal(datatypes.KeyValues{"key1": datatypes.Int(1)}))
			g.Expect(overrides.Overrides[0].Limit).To(Equal(uint64(3)))
		})
	})
}
