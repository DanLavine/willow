package query

import (
	"testing"

	"github.com/DanLavine/willow/pkg/models/datatypes"
	. "github.com/onsi/gomega"
)

func Test_Value_Validate(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It accepts a Value and ValueComparison", func(t *testing.T) {
		dInt := datatypes.Int(5)
		value := &Value{Value: &dInt, ValueComparison: Equals()}

		err := value.Validate()
		g.Expect(err).ToNot(HaveOccurred())
	})

	t.Run("It accepts an Exists", func(t *testing.T) {
		True := true
		value := &Value{Exists: &True}

		err := value.Validate()
		g.Expect(err).ToNot(HaveOccurred())
	})

	t.Run("It returns an error if Exists and Value are nil", func(t *testing.T) {
		value := &Value{}

		err := value.Validate()
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(Equal(": Requires an Exists or Value check"))
	})

	t.Run("It returns an error if both Exists and Value are not nil", func(t *testing.T) {
		True := true
		dInt := datatypes.Int(5)
		value := &Value{Exists: &True, Value: &dInt}

		err := value.Validate()
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(Equal(": Can only contain a single Exists or Value check, not both"))
	})

	t.Run("Context when setting a Value", func(t *testing.T) {
		t.Run("It returns an error if the ValueComparison is not set", func(t *testing.T) {
			dInt := datatypes.Int(5)
			value := &Value{Value: &dInt}

			err := value.Validate()
			g.Expect(err).To(HaveOccurred())
			g.Expect(err.Error()).To(Equal(".ValueComparison: is required for a Value"))
		})

		t.Run("It returns an error if the ValueComparison is unkown", func(t *testing.T) {
			dInt := datatypes.Int(5)
			badComparison := Comparison("foo")

			value := &Value{Value: &dInt, ValueComparison: &badComparison}

			err := value.Validate()
			g.Expect(err).To(HaveOccurred())
			g.Expect(err.Error()).To(Equal(".ValueComparison: received an unexpected key"))
		})
	})
}
