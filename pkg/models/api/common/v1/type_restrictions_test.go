package v1

import (
	"testing"

	"github.com/DanLavine/willow/pkg/models/datatypes"

	. "github.com/onsi/gomega"
)

func Test_TypeRestritions_Validate(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It returns an error if the MinDataType is not valid", func(t *testing.T) {
		valueTypeRestrictions := TypeRestrictions{}

		err := valueTypeRestrictions.Validate()
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("MinDataType: unknown value received '0'"))
	})

	t.Run("It returns an error if the MaxDataType is not valid", func(t *testing.T) {
		valueTypeRestrictions := TypeRestrictions{MinDataType: datatypes.T_int}

		err := valueTypeRestrictions.Validate()
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("MaxDataType: unknown value received '0'"))
	})

	t.Run("It returns an error if the MinDataType is greater than MaxDataType", func(t *testing.T) {
		valueTypeRestrictions := TypeRestrictions{MinDataType: datatypes.T_any, MaxDataType: datatypes.T_int}

		err := valueTypeRestrictions.Validate()
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("MinDataType is greater than MaxDataType"))
	})
}
