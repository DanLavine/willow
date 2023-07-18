package query

import (
	"testing"

	. "github.com/onsi/gomega"
)

func Test_Select_Validate(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It accepts all empty values. This is a Select ALL", func(t *testing.T) {
		selection := &Select{}

		err := selection.Validate()
		g.Expect(err).ToNot(HaveOccurred())
	})

	t.Run("Context when the WHERE clause is bad", func(t *testing.T) {
		t.Run("It propigates the error", func(t *testing.T) {
			selection := &Select{Where: &Query{Limits: &KeyLimits{}}}

			err := selection.Validate()
			g.Expect(err).To(HaveOccurred())
			g.Expect(err.Error()).To(Equal("Where.Limits.NumberOfKeys: Requires an int to be provided"))
		})
	})

	t.Run("Context when the OR clause is bad", func(t *testing.T) {
		t.Run("It propigates the error", func(t *testing.T) {
			selection := &Select{Or: []Select{{Where: &Query{Limits: &KeyLimits{}}}}}

			err := selection.Validate()
			g.Expect(err).To(HaveOccurred())
			g.Expect(err.Error()).To(Equal("Or[0].Where.Limits.NumberOfKeys: Requires an int to be provided"))
		})
	})

	t.Run("Context when the AND clause is bad", func(t *testing.T) {
		t.Run("It propigates the error", func(t *testing.T) {
			selection := &Select{And: []Select{{Where: &Query{Limits: &KeyLimits{}}}}}

			err := selection.Validate()
			g.Expect(err).To(HaveOccurred())
			g.Expect(err.Error()).To(Equal("And[0].Where.Limits.NumberOfKeys: Requires an int to be provided"))
		})
	})

	t.Run("Describe multi level joins", func(t *testing.T) {
		t.Run("It reports errors nicely", func(t *testing.T) {
			numberOfKeys := 5

			selection := &Select{
				And: []Select{
					{Or: []Select{
						{Where: &Query{Limits: &KeyLimits{NumberOfKeys: &numberOfKeys}}},
						{Where: &Query{Limits: &KeyLimits{}}},
					},
					},
				},
			}

			err := selection.Validate()
			g.Expect(err).To(HaveOccurred())
			g.Expect(err.Error()).To(Equal("And[0].Or[1].Where.Limits.NumberOfKeys: Requires an int to be provided"))
		})
	})
}

func Test_Select_Parse(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It can parse a JSON select", func(t *testing.T) {
		selection, err := ParseSelect([]byte(`{"Where": {"KeyValues": {"value1":{"Exists":true}}}}`))
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(*(selection.Where.KeyValues["value1"].Exists)).To(BeTrue())
	})
}
