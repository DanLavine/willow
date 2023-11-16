package datatypes

import (
	"testing"

	. "github.com/onsi/gomega"
)

func Test_KeyValueSelection_SortedKeys(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It returns a slice of all keys in a sorted order", func(t *testing.T) {
		True := true

		keyValueSelection := &KeyValueSelection{
			KeyValues: map[string]Value{
				"3": {Exists: &True},
				"2": {Exists: &True},
				"1": {Exists: &True},
				"6": {Exists: &True},
				"5": {Exists: &True},
				"7": {Exists: &True},
				"4": {Exists: &True},
			},
		}

		sortedKeys := keyValueSelection.SortedKeys()
		g.Expect(sortedKeys).To(Equal([]string{"1", "2", "3", "4", "5", "6", "7"}))
	})
}

func Test_KeyValueSelection_Validate(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It returns an error if there are no KeyValues and no Limits", func(t *testing.T) {
		keyValueSelection := &KeyValueSelection{}

		err := keyValueSelection.Validate()
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(Equal(": requires KeyValues or Limits parameters"))
	})

	t.Run("Context KeyValues", func(t *testing.T) {
		t.Run("It returns an error if any Values are not correct", func(t *testing.T) {
			keyValueSelection := &KeyValueSelection{KeyValues: map[string]Value{"one": {}}}

			err := keyValueSelection.Validate()
			g.Expect(err).To(HaveOccurred())
			g.Expect(err.Error()).To(Equal(".KeyValues[one]: Requires an Exists or Value check"))
		})
	})

	t.Run("Context when Limts are provided", func(t *testing.T) {
		t.Run("It returns an error if NumberOfKeys are nil", func(t *testing.T) {
			keyValueSelection := &KeyValueSelection{Limits: &KeyLimits{}}

			err := keyValueSelection.Validate()
			g.Expect(err).To(HaveOccurred())
			g.Expect(err.Error()).To(Equal(".Limits.NumberOfKeys: Requires an int to be provided"))
		})

		t.Run("It returns an error if NumberOfKeys is 0", func(t *testing.T) {
			zero := 0
			keyValueSelection := &KeyValueSelection{Limits: &KeyLimits{NumberOfKeys: &zero}}

			err := keyValueSelection.Validate()
			g.Expect(err).To(HaveOccurred())
			g.Expect(err.Error()).To(Equal(".Limits.NumberOfKeys: Must be larger than the provided 0"))
		})

		t.Run("It returns an error if NumberOfKeys is less than the number of KeyValues to search for", func(t *testing.T) {
			one := 1
			True := true
			keyValueSelection := &KeyValueSelection{KeyValues: map[string]Value{"one": {Exists: &True}, "two": {Exists: &True}}, Limits: &KeyLimits{NumberOfKeys: &one}}

			err := keyValueSelection.Validate()
			g.Expect(err).To(HaveOccurred())
			g.Expect(err.Error()).To(Equal(".Limits.NumberOfKeys: Is Less than the number of KeyValues to match. Will always result in 0 matches"))
		})
	})
}
