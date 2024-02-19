package datatypes

import (
	"testing"

	. "github.com/onsi/gomega"
)

func Test_Value_validate(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It returns an error if Exists and Value are nil", func(t *testing.T) {
		value := &Value{}

		err := value.validate()
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(Equal(": Requires an Exists or Value check"))
	})

	t.Run("It returns an error if both Exists and Value are not nil", func(t *testing.T) {
		True := true
		dInt := Int(5)
		value := &Value{Exists: &True, Value: &dInt}

		err := value.validate()
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(Equal(": Can only contain a single Exists or Value check, not both"))
	})

	t.Run("Context when setting Exists", func(t *testing.T) {
		t.Run("It accepts only an Exists", func(t *testing.T) {
			True := true
			value := &Value{Exists: &True}

			err := value.validate()
			g.Expect(err).ToNot(HaveOccurred())
		})

		t.Run("It accepts setting an ExistsType", func(t *testing.T) {
			True := true
			intType := T_int
			value := &Value{Exists: &True, ExistsType: &intType}

			err := value.validate()
			g.Expect(err).ToNot(HaveOccurred())
		})

		t.Run("It returns an error if the ExistsType is unkown", func(t *testing.T) {
			True := true
			badType := DataType(8000)
			value := &Value{Exists: &True, ExistsType: &badType}

			err := value.validate()
			g.Expect(err).To(HaveOccurred())
			g.Expect(err.Error()).To(Equal(": Unexpected ExistsType. Must be an int from 1-13 inclusive"))
		})

		t.Run("It returns an error if a Value is set", func(t *testing.T) {
			True := true
			valueString := String("wops")
			value := &Value{Exists: &True, Value: &valueString}

			err := value.validate()
			g.Expect(err).To(HaveOccurred())
			g.Expect(err.Error()).To(Equal(": Can only contain a single Exists or Value check, not both"))
		})
	})

	t.Run("Context when setting a Value", func(t *testing.T) {
		t.Run("It accepts a Value and ValueComparison", func(t *testing.T) {
			dInt := Int(5)
			value := &Value{Value: &dInt, ValueComparison: EqualsPtr()}

			err := value.validate()
			g.Expect(err).ToNot(HaveOccurred())
		})

		t.Run("It returns an error if the ValueComparison is not set", func(t *testing.T) {
			dInt := Int(5)
			value := &Value{Value: &dInt}

			err := value.validate()
			g.Expect(err).To(HaveOccurred())
			g.Expect(err.Error()).To(Equal(".ValueComparison: is required for a Value"))
		})

		t.Run("It returns an error if the ValueComparison is unknown", func(t *testing.T) {
			dInt := Int(5)
			badComparison := Comparison("foo")

			value := &Value{Value: &dInt, ValueComparison: &badComparison}

			err := value.validate()
			g.Expect(err).To(HaveOccurred())
			g.Expect(err.Error()).To(Equal(".ValueComparison: received an unexpected key"))
		})
	})
}

func Test_Value_validateReservedKey(t *testing.T) {
	g := NewGomegaWithT(t)

	isTrue := true

	t.Run("It returns an error if Exists is set", func(t *testing.T) {
		value := &Value{Exists: &isTrue}

		err := value.validateReservedKey()
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(Equal(": cannot be an existence check. It can only mach an exact string"))
	})

	t.Run("It returns an error if ExistsType is set", func(t *testing.T) {
		value := &Value{ExistsType: &T_int64}

		err := value.validateReservedKey()
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(Equal(": cannot check an existence type. It can only mach an exact string"))
	})

	t.Run("It returns an error if Value is not set", func(t *testing.T) {
		value := &Value{}

		err := value.validateReservedKey()
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(Equal(": requires a string Value to match against"))
	})

	t.Run("It returns an error if Value is not a string", func(t *testing.T) {
		intValue := Int(32)
		value := &Value{Value: &intValue}

		err := value.validateReservedKey()
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(Equal(": requires a string Value to match against"))
	})

	t.Run("It returns an error if Value is not a string", func(t *testing.T) {
		strValue := String("32")
		value := &Value{Value: &strValue, ValueComparison: LessThanPtr()}

		err := value.validateReservedKey()
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(Equal(": requires an Equals ValueComparison"))
	})

	t.Run("It accepts a string to equal againsy", func(t *testing.T) {
		strValue := String("32")
		value := &Value{Value: &strValue, ValueComparison: EqualsPtr()}

		err := value.validateReservedKey()
		g.Expect(err).ToNot(HaveOccurred())
	})
}
