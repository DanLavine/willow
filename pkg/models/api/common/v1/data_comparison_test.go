package v1

import (
	"testing"

	. "github.com/onsi/gomega"
)

func Test_DataComparison_Validate(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It returns an error if comparison is unkown", func(t *testing.T) {
		dataComparison := DataComparison("nope")

		err := dataComparison.Validate()
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("unknown value 'nope'"))
	})
}
