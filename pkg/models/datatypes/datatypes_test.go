package datatypes

import (
	"testing"

	. "github.com/onsi/gomega"
)

func Test_DataTypes_ToString(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("types are convertied to proper strings", func(t *testing.T) {
		g.Expect(T_uint8.ToString()).To(Equal("uint8"))
		g.Expect(T_uint16.ToString()).To(Equal("uint16"))
		g.Expect(T_uint32.ToString()).To(Equal("uint32"))
		g.Expect(T_uint64.ToString()).To(Equal("uint64"))
		g.Expect(T_uint.ToString()).To(Equal("uint"))
		g.Expect(T_int8.ToString()).To(Equal("int8"))
		g.Expect(T_int16.ToString()).To(Equal("int16"))
		g.Expect(T_int32.ToString()).To(Equal("int32"))
		g.Expect(T_int64.ToString()).To(Equal("int64"))
		g.Expect(T_int.ToString()).To(Equal("int"))
		g.Expect(T_float32.ToString()).To(Equal("float32"))
		g.Expect(T_float64.ToString()).To(Equal("float64"))
		g.Expect(T_string.ToString()).To(Equal("string"))
		g.Expect(T_any.ToString()).To(Equal("any"))
		g.Expect(DataType(17).ToString()).To(Equal("unknown"))
	})

}

func Test_DataTypes_Less(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("same types are always false", func(t *testing.T) {
		for _, dataType := range []DataType{T_uint8, T_uint16, T_uint32, T_uint64, T_uint, T_int8, T_int16, T_int32, T_int64, T_int, T_float32, T_float64, T_string, T_any} {
			g.Expect(dataType.Less(dataType)).To(BeFalse())
		}
	})

	t.Run("types compared against T_any are always false", func(t *testing.T) {
		for _, dataType := range []DataType{T_uint8, T_uint16, T_uint32, T_uint64, T_uint, T_int8, T_int16, T_int32, T_int64, T_int, T_float32, T_float64, T_string} {
			g.Expect(dataType.Less(T_any)).To(BeFalse())
			g.Expect(T_any.Less(dataType)).To(BeFalse())
		}
	})

	t.Run("types compared against each other report properly", func(t *testing.T) {
		typeRange := []DataType{T_uint8, T_uint16, T_uint32, T_uint64, T_uint, T_int8, T_int16, T_int32, T_int64, T_int, T_float32, T_float64, T_string}
		for index, dataType := range typeRange {
			for i := 0; i < len(typeRange); i++ {
				if i == index {
					continue
				}

				if i < index {
					g.Expect(typeRange[i].Less(dataType)).To(BeTrue())
				} else {
					g.Expect(typeRange[i].Less(dataType)).To(BeFalse())
				}
			}
		}
	})
}

func Test_DataTypes_LessMatchType(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("same types are always false", func(t *testing.T) {
		for _, dataType := range []DataType{T_uint8, T_uint16, T_uint32, T_uint64, T_uint, T_int8, T_int16, T_int32, T_int64, T_int, T_float32, T_float64, T_string, T_any} {
			g.Expect(dataType.LessMatchType(dataType)).To(BeFalse())
		}
	})

	t.Run("types compared against T_any are always compared against their value", func(t *testing.T) {
		for _, dataType := range []DataType{T_uint8, T_uint16, T_uint32, T_uint64, T_uint, T_int8, T_int16, T_int32, T_int64, T_int, T_float32, T_float64, T_string} {
			g.Expect(dataType.LessMatchType(T_any)).To(BeTrue())
			g.Expect(T_any.LessMatchType(dataType)).To(BeFalse())
		}
	})

	t.Run("types compared against each other report properly", func(t *testing.T) {
		typeRange := []DataType{T_uint8, T_uint16, T_uint32, T_uint64, T_uint, T_int8, T_int16, T_int32, T_int64, T_int, T_float32, T_float64, T_string}
		for index, dataType := range typeRange {
			for i := 0; i < len(typeRange); i++ {
				if i == index {
					continue
				}

				if i < index {
					g.Expect(typeRange[i].LessMatchType(dataType)).To(BeTrue())
				} else {
					g.Expect(typeRange[i].LessMatchType(dataType)).To(BeFalse())
				}
			}
		}
	})
}
