package datatypes

/*
import (
	"encoding/json"
	"testing"

	. "github.com/onsi/gomega"
)

func Test_TypeQuery_Validate(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It returns an error if the Type is invalid", func(t *testing.T) {
		typeExistenceQuery := TypeQuery{Type: 0}

		err := typeExistenceQuery.Validate()
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("Type: is outside the range for for valid types [1:1024], but received '0'"))
	})

	t.Run("It returns an error if the DataTypeRestrictions is invalid", func(t *testing.T) {
		typeExistenceQuery := TypeQuery{Type: T_int64, DataTypeRestrictions: DataTypeRestrictions{MinDataType: -1}}

		err := typeExistenceQuery.Validate()
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("DataTypeRestrictions.MinDataType: unknown value received '-1'"))
	})
}

func Test_DataQuery_Validate(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It returns an error if the EncapsulatedValue is invalid", func(t *testing.T) {
		typeExistenceQuery := DataQuery{EncapsulatedValue: EncapsulatedValue{Type: 0, Data: "bad"}, DataComparison: Equals}

		err := typeExistenceQuery.Validate()
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("EncapsulatedValue.Type: 'unkown' has Data: bad"))
	})

	t.Run("It returns an error if the DataComparison is not valid", func(t *testing.T) {
		typeExistenceQuery := DataQuery{EncapsulatedValue: String("ok"), DataComparison: DataComparison("wooo")}

		err := typeExistenceQuery.Validate()
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("DataComparison: unknown value received 'wooo'"))
	})

	t.Run("It returns an error if the DataTypeRestrictions is invalid", func(t *testing.T) {
		typeExistenceQuery := DataQuery{
			EncapsulatedValue: String("ok"),
			DataComparison:    Equals,
			DataTypeRestrictions: DataTypeRestrictions{
				MinDataType: T_float64,
				MaxDataType: T_int,
			},
		}

		err := typeExistenceQuery.Validate()
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("DataTypeRestrictions: MinDataType is greater than MaxDataType"))
	})
}

func Test_DataTypeRestrictions_Validate(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It returns an error if the MinDataType is invalid", func(t *testing.T) {
		dataTypeRestrictions := DataTypeRestrictions{MinDataType: -1}

		err := dataTypeRestrictions.Validate()
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("MinDataType: unknown value received '-1'"))
	})

	t.Run("It returns an error if the MaxDataType is invalid", func(t *testing.T) {
		dataTypeRestrictions := DataTypeRestrictions{MaxDataType: 10_000}

		err := dataTypeRestrictions.Validate()
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("MaxDataType: unknown value received '10000'"))
	})

	t.Run("It returns an error if the MinDataType is greater than the MaxDataType", func(t *testing.T) {
		dataTypeRestrictions := DataTypeRestrictions{MinDataType: 4, MaxDataType: 3}

		err := dataTypeRestrictions.Validate()
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("MinDataType is greater than MaxDataType"))
	})
}

func Test_DataTypeRestrictions_Encoding(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("Describe JSON", func(t *testing.T) {
		t.Run("It sets default values as part of the Unmarshal", func(t *testing.T) {
			dataTypeRestrictions := DataTypeRestrictions{}

			err := json.Unmarshal([]byte(`{}`), &dataTypeRestrictions)
			g.Expect(err).ToNot(HaveOccurred())

			g.Expect(dataTypeRestrictions.MinDataType).To(Equal(T_uint8))
			g.Expect(dataTypeRestrictions.MaxDataType).To(Equal(T_any))
		})
	})
}
*/
