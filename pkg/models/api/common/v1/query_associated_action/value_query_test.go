package queryassociatedaction

import (
	"testing"

	v1 "github.com/DanLavine/willow/pkg/models/api/common/v1"
	"github.com/DanLavine/willow/pkg/models/datatypes"

	. "github.com/onsi/gomega"
)

func Test_ValueQuery_Validate(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It returns an error if TypeRestriction is invalid", func(t *testing.T) {
		valueQuery := ValueQuery{
			Value:            datatypes.Any(),
			Comparison:       v1.Equals,
			TypeRestrictions: v1.TypeRestrictions{MinDataType: datatypes.MaxDataType, MaxDataType: datatypes.MinDataType},
		}

		err := valueQuery.Validate()
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("TypeRestrictions: MinDataType is greater than MaxDataType"))
	})

	t.Run("Describe Value with comparisons", func(t *testing.T) {
		t.Run("It returns an error if Comparison is invalid", func(t *testing.T) {
			valueQuery := ValueQuery{Value: datatypes.Any(), Comparison: "bad"}

			err := valueQuery.Validate()
			g.Expect(err).To(HaveOccurred())
			g.Expect(err.Error()).To(ContainSubstring("Comparison: unknown value 'bad'"))
		})

		t.Run("It allows T_any value when the Comparison is [Equals | NotEquals]", func(t *testing.T) {
			goodValueQuery := ValueQuery{
				Value:            datatypes.Any(),
				Comparison:       v1.Equals,
				TypeRestrictions: v1.TypeRestrictions{MinDataType: datatypes.MinDataType, MaxDataType: datatypes.MaxDataType},
			}

			err := goodValueQuery.Validate()
			g.Expect(err).ToNot(HaveOccurred())
		})

		t.Run("It does not allow T_any for all other comparisons", func(t *testing.T) {
			badValueQuery := ValueQuery{
				Value:            datatypes.Any(),
				Comparison:       v1.LessThan,
				TypeRestrictions: v1.TypeRestrictions{MinDataType: datatypes.MinDataType, MaxDataType: datatypes.MaxDataType},
			}

			err := badValueQuery.Validate()
			g.Expect(err).To(HaveOccurred())
			g.Expect(err.Error()).To(ContainSubstring("Value.Type: invalid value '1024'. The required value must be with the data types [1:13] inclusively"))
		})
	})
}
