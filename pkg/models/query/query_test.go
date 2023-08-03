package query

import (
	"testing"

	. "github.com/onsi/gomega"
)

func Test_Query_SortedKeys(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It returns a slice of all keys in a sorted order", func(t *testing.T) {
		True := true

		query := &Query{
			KeyValues: map[string]Value{
				"3": Value{Exists: &True},
				"2": Value{Exists: &True},
				"1": Value{Exists: &True},
				"6": Value{Exists: &True},
				"5": Value{Exists: &True},
				"7": Value{Exists: &True},
				"4": Value{Exists: &True},
			},
		}

		sortedKeys := query.SortedKeys()
		g.Expect(sortedKeys).To(Equal([]string{"1", "2", "3", "4", "5", "6", "7"}))
	})
}

func Test_Query_Validate(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It returns an error if there are no KeyValues and no Limits", func(t *testing.T) {
		query := &Query{}

		err := query.Validate()
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(Equal("Requires KeyValues or Limits parameters"))
	})

	t.Run("Context KeyValues", func(t *testing.T) {
		t.Run("It returns an error if any Values are not correct", func(t *testing.T) {
			query := &Query{KeyValues: map[string]Value{"one": Value{}}}

			err := query.Validate()
			g.Expect(err).To(HaveOccurred())
			g.Expect(err.Error()).To(Equal("KeyValues[one]: Requires an Exists or Value check"))
		})
	})

	t.Run("Context when Limts are provided", func(t *testing.T) {
		t.Run("It returns an error if NumberOfKeys are nil", func(t *testing.T) {
			query := &Query{Limits: &KeyLimits{}}

			err := query.Validate()
			g.Expect(err).To(HaveOccurred())
			g.Expect(err.Error()).To(Equal("Limits.NumberOfKeys: Requires an int to be provided"))
		})

		t.Run("It returns an error if NumberOfKeys is 0", func(t *testing.T) {
			zero := 0
			query := &Query{Limits: &KeyLimits{NumberOfKeys: &zero}}

			err := query.Validate()
			g.Expect(err).To(HaveOccurred())
			g.Expect(err.Error()).To(Equal("Limits.NumberOfKeys: Must be larger than the provided 0"))
		})

		t.Run("It returns an error if NumberOfKeys is less than the number of KeyValues to search for", func(t *testing.T) {
			one := 1
			True := true
			query := &Query{KeyValues: map[string]Value{"one": Value{Exists: &True}, "two": Value{Exists: &True}}, Limits: &KeyLimits{NumberOfKeys: &one}}

			err := query.Validate()
			g.Expect(err).To(HaveOccurred())
			g.Expect(err.Error()).To(Equal("Limits.NumberOfKeys: Is Less than the number of KeyValues to match. Will always result in 0 matches"))
		})
	})
}
