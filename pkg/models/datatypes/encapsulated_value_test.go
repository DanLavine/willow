package datatypes

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestEncapsulatedValue_Less(t *testing.T) {
	g := NewGomegaWithT(t)

	// ints
	tInt := Int(1)
	tInt8 := Int8(1)
	tInt16 := Int16(1)
	tInt32 := Int32(1)
	tInt64 := Int64(1)

	// uints
	tUint := Uint(1)
	tUint8 := Uint8(1)
	tUint16 := Uint16(1)
	tUint32 := Uint32(1)
	tUint64 := Uint64(1)

	// floats
	tFloat32 := Float32(1)
	tFloat64 := Float64(1)

	// string
	tString := String("a")

	// any
	tAny := Any()

	t.Run("keys of the same values are always equal", func(t *testing.T) {
		// ints
		//// int
		g.Expect(tInt.Less(Int(1))).To(BeFalse())
		g.Expect(Int(1).Less(tInt)).To(BeFalse())
		//// int8
		g.Expect(tInt8.Less(Int8(1))).To(BeFalse())
		g.Expect(Int8(1).Less(tInt8)).To(BeFalse())
		//// int16
		g.Expect(tInt16.Less(Int16(1))).To(BeFalse())
		g.Expect(Int16(1).Less(tInt16)).To(BeFalse())
		//// int32
		g.Expect(tInt32.Less(Int32(1))).To(BeFalse())
		g.Expect(Int32(1).Less(tInt32)).To(BeFalse())
		//// int64
		g.Expect(tInt64.Less(Int64(1))).To(BeFalse())
		g.Expect(Int64(1).Less(tInt64)).To(BeFalse())

		// uints
		//// uint
		g.Expect(tUint.Less(Uint(1))).To(BeFalse())
		g.Expect(Uint(1).Less(tUint)).To(BeFalse())
		//// int8
		g.Expect(tUint8.Less(Uint8(1))).To(BeFalse())
		g.Expect(Uint8(1).Less(tUint8)).To(BeFalse())
		//// int16
		g.Expect(tUint16.Less(Uint16(1))).To(BeFalse())
		g.Expect(Uint16(1).Less(tUint16)).To(BeFalse())
		//// int32
		g.Expect(tUint32.Less(Uint32(1))).To(BeFalse())
		g.Expect(Uint32(1).Less(tUint32)).To(BeFalse())
		//// int64
		g.Expect(tUint64.Less(Uint64(1))).To(BeFalse())
		g.Expect(Uint64(1).Less(tUint64)).To(BeFalse())

		// floats
		//// float32
		g.Expect(tFloat32.Less(Float32(1))).To(BeFalse())
		g.Expect(Float32(1).Less(tFloat32)).To(BeFalse())
		//// floatt64
		g.Expect(tFloat64.Less(Float64(1))).To(BeFalse())
		g.Expect(Float64(1).Less(tFloat64)).To(BeFalse())

		// string
		g.Expect(tString.Less(String("a"))).To(BeFalse())
		g.Expect(String("a").Less(tString)).To(BeFalse())

		// any
		g.Expect(tAny.Less(Any())).To(BeFalse())
		g.Expect(Any().Less(tAny)).To(BeFalse())
	})

	t.Run("keys of the same type have proper Less than values", func(t *testing.T) {
		// ints
		//// int
		g.Expect(tInt.Less(Int(2))).To(BeTrue())
		//// int8
		g.Expect(tInt8.Less(Int8(2))).To(BeTrue())
		//// int16
		g.Expect(tInt16.Less(Int16(2))).To(BeTrue())
		//// int32
		g.Expect(tInt32.Less(Int32(2))).To(BeTrue())
		//// int64
		g.Expect(tInt64.Less(Int64(2))).To(BeTrue())

		// uints
		//// uint
		g.Expect(tUint.Less(Uint(2))).To(BeTrue())
		//// int8
		g.Expect(tUint8.Less(Uint8(2))).To(BeTrue())
		//// int16
		g.Expect(tUint16.Less(Uint16(2))).To(BeTrue())
		//// int32
		g.Expect(tUint32.Less(Uint32(2))).To(BeTrue())
		//// int64
		g.Expect(tUint64.Less(Uint64(2))).To(BeTrue())

		// floats
		//// float32
		g.Expect(tFloat32.Less(Float32(2))).To(BeTrue())
		//// floatt64
		g.Expect(tFloat64.Less(Float64(2))).To(BeTrue())

		// string
		g.Expect(tString.Less(String("b"))).To(BeTrue())

		// any
		g.Expect(tAny.Less(Any())).To(BeFalse())
	})

	t.Run("keys of the same type respect greater than values", func(t *testing.T) {
		// ints
		//// int
		g.Expect(tInt.Less(Int(0))).To(BeFalse())
		//// int8
		g.Expect(tInt8.Less(Int8(0))).To(BeFalse())
		//// int16
		g.Expect(tInt16.Less(Int16(0))).To(BeFalse())
		//// int32
		g.Expect(tInt32.Less(Int32(0))).To(BeFalse())
		//// int64
		g.Expect(tInt64.Less(Int64(0))).To(BeFalse())

		// uints
		//// uint
		g.Expect(tUint.Less(Uint(0))).To(BeFalse())
		//// int8
		g.Expect(tUint8.Less(Uint8(0))).To(BeFalse())
		//// int16
		g.Expect(tUint16.Less(Uint16(0))).To(BeFalse())
		//// int32
		g.Expect(tUint32.Less(Uint32(0))).To(BeFalse())
		//// int64
		g.Expect(tUint64.Less(Uint64(0))).To(BeFalse())

		// floats
		//// float32
		g.Expect(tFloat32.Less(Float32(0))).To(BeFalse())
		//// floatt64
		g.Expect(tFloat64.Less(Float64(0))).To(BeFalse())

		// string
		g.Expect(tString.Less(String("0"))).To(BeFalse())

		// any
		g.Expect(Any().Less(tAny)).To(BeFalse())
	})

	t.Run("keys of different types report less properly", func(t *testing.T) {
		// ints
		//// int
		g.Expect(tInt.Less(tString)).To(BeTrue())
		g.Expect(tString.Less(tInt)).To(BeFalse())
		//// int8
		g.Expect(tInt8.Less(tString)).To(BeTrue())
		g.Expect(tString.Less(tInt8)).To(BeFalse())
		//// int16
		g.Expect(tInt16.Less(tString)).To(BeTrue())
		g.Expect(tString.Less(tInt16)).To(BeFalse())
		//// int32
		g.Expect(tInt32.Less(tString)).To(BeTrue())
		g.Expect(tString.Less(tInt32)).To(BeFalse())
		//// int64
		g.Expect(tInt64.Less(tString)).To(BeTrue())
		g.Expect(tString.Less(tInt64)).To(BeFalse())

		// uints
		//// uint
		g.Expect(tUint.Less(tString)).To(BeTrue())
		g.Expect(tString.Less(tUint)).To(BeFalse())
		//// int8
		g.Expect(tUint8.Less(tString)).To(BeTrue())
		g.Expect(tString.Less(tUint8)).To(BeFalse())
		//// int16
		g.Expect(tUint16.Less(tString)).To(BeTrue())
		g.Expect(tString.Less(tUint16)).To(BeFalse())
		//// int32
		g.Expect(tUint32.Less(tString)).To(BeTrue())
		g.Expect(tString.Less(tUint32)).To(BeFalse())
		//// int64
		g.Expect(tUint64.Less(tString)).To(BeTrue())
		g.Expect(tString.Less(tUint64)).To(BeFalse())

		// floats
		//// float32
		g.Expect(tFloat32.Less(tString)).To(BeTrue())
		g.Expect(tString.Less(tFloat32)).To(BeFalse())
		//// floatt64
		g.Expect(tFloat64.Less(tString)).To(BeTrue())
		g.Expect(tString.Less(tFloat64)).To(BeFalse())

		// string with any
		g.Expect(tString.Less(tAny)).To(BeFalse())
		g.Expect(tAny.Less(tString)).To(BeFalse())
	})

	t.Run("keys of type T_any always returns false", func(t *testing.T) {
		// ints
		//// int
		g.Expect(tInt.Less(tAny)).To(BeFalse())
		g.Expect(tAny.Less(tInt)).To(BeFalse())
		//// int8
		g.Expect(tInt8.Less(tAny)).To(BeFalse())
		g.Expect(tAny.Less(tInt8)).To(BeFalse())
		//// int16
		g.Expect(tInt16.Less(tAny)).To(BeFalse())
		g.Expect(tAny.Less(tInt16)).To(BeFalse())
		//// int32
		g.Expect(tInt32.Less(tAny)).To(BeFalse())
		g.Expect(tAny.Less(tInt32)).To(BeFalse())
		//// int64
		g.Expect(tInt64.Less(tAny)).To(BeFalse())
		g.Expect(tAny.Less(tInt64)).To(BeFalse())

		// uints
		//// uint
		g.Expect(tUint.Less(tAny)).To(BeFalse())
		g.Expect(tAny.Less(tUint)).To(BeFalse())
		//// int8
		g.Expect(tUint8.Less(tAny)).To(BeFalse())
		g.Expect(tAny.Less(tUint8)).To(BeFalse())
		//// int16
		g.Expect(tUint16.Less(tAny)).To(BeFalse())
		g.Expect(tAny.Less(tUint16)).To(BeFalse())
		//// int32
		g.Expect(tUint32.Less(tAny)).To(BeFalse())
		g.Expect(tAny.Less(tUint32)).To(BeFalse())
		//// int64
		g.Expect(tUint64.Less(tAny)).To(BeFalse())
		g.Expect(tAny.Less(tUint64)).To(BeFalse())

		// floats
		//// float32
		g.Expect(tFloat32.Less(tAny)).To(BeFalse())
		g.Expect(tAny.Less(tFloat32)).To(BeFalse())
		//// floatt64
		g.Expect(tFloat64.Less(tAny)).To(BeFalse())
		g.Expect(tAny.Less(tFloat64)).To(BeFalse())

		// string
		g.Expect(tString.Less(tAny)).To(BeFalse())
		g.Expect(tAny.Less(tString)).To(BeFalse())

		// any
		g.Expect(tAny.Less(Any())).To(BeFalse())
		g.Expect(Any().Less(tAny)).To(BeFalse())
	})
}

func TestEncapsulatedValue_LessMatchType(t *testing.T) {
	g := NewGomegaWithT(t)

	// ints
	tInt := Int(1)
	tInt8 := Int8(1)
	tInt16 := Int16(1)
	tInt32 := Int32(1)
	tInt64 := Int64(1)

	// uints
	tUint := Uint(1)
	tUint8 := Uint8(1)
	tUint16 := Uint16(1)
	tUint32 := Uint32(1)
	tUint64 := Uint64(1)

	// floats
	tFloat32 := Float32(1)
	tFloat64 := Float64(1)

	// string
	tString := String("a")

	// any
	tAny := Any()

	t.Run("type of the same values are always equal", func(t *testing.T) {
		// ints
		//// int
		g.Expect(tInt.LessMatchType(Int(1))).To(BeFalse())
		g.Expect(Int(1).LessMatchType(tInt)).To(BeFalse())
		//// int8
		g.Expect(tInt8.LessMatchType(Int8(1))).To(BeFalse())
		g.Expect(Int8(1).LessMatchType(tInt8)).To(BeFalse())
		//// int16
		g.Expect(tInt16.LessMatchType(Int16(1))).To(BeFalse())
		g.Expect(Int16(1).LessMatchType(tInt16)).To(BeFalse())
		//// int32
		g.Expect(tInt32.LessMatchType(Int32(1))).To(BeFalse())
		g.Expect(Int32(1).LessMatchType(tInt32)).To(BeFalse())
		//// int64
		g.Expect(tInt64.LessMatchType(Int64(1))).To(BeFalse())
		g.Expect(Int64(1).LessMatchType(tInt64)).To(BeFalse())

		// uints
		//// uint
		g.Expect(tUint.LessMatchType(Uint(1))).To(BeFalse())
		g.Expect(Uint(1).LessMatchType(tUint)).To(BeFalse())
		//// int8
		g.Expect(tUint8.LessMatchType(Uint8(1))).To(BeFalse())
		g.Expect(Uint8(1).LessMatchType(tUint8)).To(BeFalse())
		//// int16
		g.Expect(tUint16.LessMatchType(Uint16(1))).To(BeFalse())
		g.Expect(Uint16(1).LessMatchType(tUint16)).To(BeFalse())
		//// int32
		g.Expect(tUint32.LessMatchType(Uint32(1))).To(BeFalse())
		g.Expect(Uint32(1).LessMatchType(tUint32)).To(BeFalse())
		//// int64
		g.Expect(tUint64.LessMatchType(Uint64(1))).To(BeFalse())
		g.Expect(Uint64(1).LessMatchType(tUint64)).To(BeFalse())

		// floats
		//// float32
		g.Expect(tFloat32.LessMatchType(Float32(1))).To(BeFalse())
		g.Expect(Float32(1).LessMatchType(tFloat32)).To(BeFalse())
		//// floatt64
		g.Expect(tFloat64.LessMatchType(Float64(1))).To(BeFalse())
		g.Expect(Float64(1).LessMatchType(tFloat64)).To(BeFalse())

		// string
		g.Expect(tString.LessMatchType(String("a"))).To(BeFalse())
		g.Expect(String("a").LessMatchType(tString)).To(BeFalse())

		// any
		g.Expect(tAny.LessMatchType(Any())).To(BeFalse())
		g.Expect(Any().LessMatchType(tAny)).To(BeFalse())
	})

	t.Run("different types are compared properly", func(t *testing.T) {
		// ints
		//// int
		g.Expect(tInt.LessMatchType(String("1"))).To(BeTrue())
		g.Expect(String("1").LessMatchType(tInt)).To(BeFalse())
		//// int8
		g.Expect(tInt8.LessMatchType(String("1"))).To(BeTrue())
		g.Expect(String("1").LessMatchType(tInt8)).To(BeFalse())
		//// int16
		g.Expect(tInt16.LessMatchType(String("1"))).To(BeTrue())
		g.Expect(String("1").LessMatchType(tInt16)).To(BeFalse())
		//// int32
		g.Expect(tInt32.LessMatchType(String("1"))).To(BeTrue())
		g.Expect(String("1").LessMatchType(tInt32)).To(BeFalse())
		//// int64
		g.Expect(tInt64.LessMatchType(String("1"))).To(BeTrue())
		g.Expect(String("1").LessMatchType(tInt64)).To(BeFalse())

		// uints
		//// uint
		g.Expect(tUint.LessMatchType(String("1"))).To(BeTrue())
		g.Expect(String("1").LessMatchType(tUint)).To(BeFalse())
		//// int8
		g.Expect(tUint8.LessMatchType(String("1"))).To(BeTrue())
		g.Expect(String("1").LessMatchType(tUint8)).To(BeFalse())
		//// int16
		g.Expect(tUint16.LessMatchType(String("1"))).To(BeTrue())
		g.Expect(String("1").LessMatchType(tUint16)).To(BeFalse())
		//// int32
		g.Expect(tUint32.LessMatchType(String("1"))).To(BeTrue())
		g.Expect(String("1").LessMatchType(tUint32)).To(BeFalse())
		//// int64
		g.Expect(tUint64.LessMatchType(String("1"))).To(BeTrue())
		g.Expect(String("1").LessMatchType(tUint64)).To(BeFalse())

		// floats
		//// float32
		g.Expect(tFloat32.LessMatchType(String("1"))).To(BeTrue())
		g.Expect(String("1").LessMatchType(tFloat32)).To(BeFalse())
		//// floatt64
		g.Expect(tFloat64.LessMatchType(String("1"))).To(BeTrue())
		g.Expect(String("1").LessMatchType(tFloat64)).To(BeFalse())

		// string
		g.Expect(tString.LessMatchType(Int(3))).To(BeFalse())
	})

	t.Run("type compared agains T_any reprot properly", func(t *testing.T) {
		// ints
		//// int
		g.Expect(tInt.LessMatchType(tAny)).To(BeTrue())
		g.Expect(tAny.LessMatchType(tInt)).To(BeFalse())
		//// int8
		g.Expect(tInt8.LessMatchType(tAny)).To(BeTrue())
		g.Expect(tAny.LessMatchType(tInt8)).To(BeFalse())
		//// int16
		g.Expect(tInt16.LessMatchType(tAny)).To(BeTrue())
		g.Expect(tAny.LessMatchType(tInt16)).To(BeFalse())
		//// int32
		g.Expect(tInt32.LessMatchType(tAny)).To(BeTrue())
		g.Expect(tAny.LessMatchType(tInt32)).To(BeFalse())
		//// int64
		g.Expect(tInt64.LessMatchType(tAny)).To(BeTrue())
		g.Expect(tAny.LessMatchType(tInt64)).To(BeFalse())

		// uints
		//// uint
		g.Expect(tUint.LessMatchType(tAny)).To(BeTrue())
		g.Expect(tAny.LessMatchType(tUint)).To(BeFalse())
		//// int8
		g.Expect(tUint8.LessMatchType(tAny)).To(BeTrue())
		g.Expect(tAny.LessMatchType(tUint8)).To(BeFalse())
		//// int16
		g.Expect(tUint16.LessMatchType(tAny)).To(BeTrue())
		g.Expect(tAny.LessMatchType(tUint16)).To(BeFalse())
		//// int32
		g.Expect(tUint32.LessMatchType(tAny)).To(BeTrue())
		g.Expect(tAny.LessMatchType(tUint32)).To(BeFalse())
		//// int64
		g.Expect(tUint64.LessMatchType(tAny)).To(BeTrue())
		g.Expect(tAny.LessMatchType(tUint64)).To(BeFalse())

		// floats
		//// float32
		g.Expect(tFloat32.LessMatchType(tAny)).To(BeTrue())
		g.Expect(tAny.LessMatchType(tFloat32)).To(BeFalse())
		//// floatt64
		g.Expect(tFloat64.LessMatchType(tAny)).To(BeTrue())
		g.Expect(tAny.LessMatchType(tFloat64)).To(BeFalse())

		// string
		g.Expect(tString.LessMatchType(tAny)).To(BeTrue())
		g.Expect(tAny.LessMatchType(tString)).To(BeFalse())
	})
}

func TestEncapsulatedValue_Validate(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("it returns an error if the Value is nil", func(t *testing.T) {
		encapsulatedData := EncapsulatedValue{Type: T_int, Data: nil}

		err := encapsulatedData.Validate(MinDataType, MaxDataType)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(Equal("Type: 'int' has nil Data, but requires a castable string"))
	})

	t.Run("it returns an error if the DataTypes is unknown", func(t *testing.T) {
		encapsulatedData := EncapsulatedValue{Type: DataType(1_000_000), Data: "something"}

		err := encapsulatedData.Validate(MinDataType, MaxDataType)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(Equal("Type: 'unknown' has Data 'something'"))
	})

	t.Run("it returns an error if the DataType and value don't match", func(t *testing.T) {
		// ints
		//// int
		err := EncapsulatedValue{Type: T_int, Data: "nope"}.Validate(MinDataType, MaxDataType)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(Equal("Type: 'int' has Data of kind: string"))
		//// int8
		err = EncapsulatedValue{Type: T_int8, Data: "nope"}.Validate(MinDataType, MaxDataType)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(Equal("Type: 'int8' has Data of kind: string"))
		//// int16
		err = EncapsulatedValue{Type: T_int16, Data: "nope"}.Validate(MinDataType, MaxDataType)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(Equal("Type: 'int16' has Data of kind: string"))
		//// int32
		err = EncapsulatedValue{Type: T_int32, Data: "nope"}.Validate(MinDataType, MaxDataType)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(Equal("Type: 'int32' has Data of kind: string"))
		//// int64
		err = EncapsulatedValue{Type: T_int64, Data: "nope"}.Validate(MinDataType, MaxDataType)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(Equal("Type: 'int64' has Data of kind: string"))

		// uints
		//// uint
		err = EncapsulatedValue{Type: T_uint, Data: "nope"}.Validate(MinDataType, MaxDataType)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(Equal("Type: 'uint' has Data of kind: string"))
		//// uint8
		err = EncapsulatedValue{Type: T_uint8, Data: "nope"}.Validate(MinDataType, MaxDataType)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(Equal("Type: 'uint8' has Data of kind: string"))
		//// uint16
		err = EncapsulatedValue{Type: T_uint16, Data: "nope"}.Validate(MinDataType, MaxDataType)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(Equal("Type: 'uint16' has Data of kind: string"))
		//// uint32
		err = EncapsulatedValue{Type: T_uint32, Data: "nope"}.Validate(MinDataType, MaxDataType)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(Equal("Type: 'uint32' has Data of kind: string"))
		//// uint64
		err = EncapsulatedValue{Type: T_uint64, Data: "nope"}.Validate(MinDataType, MaxDataType)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(Equal("Type: 'uint64' has Data of kind: string"))

		// floats
		//// float32
		err = EncapsulatedValue{Type: T_float32, Data: "nope"}.Validate(MinDataType, MaxDataType)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(Equal("Type: 'float32' has Data of kind: string"))
		//// float64
		err = EncapsulatedValue{Type: T_float64, Data: "nope"}.Validate(MinDataType, MaxDataType)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(Equal("Type: 'float64' has Data of kind: string"))

		// string
		err = EncapsulatedValue{Type: T_string, Data: int(42)}.Validate(MinDataType, MaxDataType)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(Equal("Type: 'string' has Data of kind: int"))
	})

	t.Run("It returns an error if the type is outside the min range", func(t *testing.T) {
		err := EncapsulatedValue{Type: T_int, Data: int(42)}.Validate(T_string, MaxDataType)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("Type: invalid value '10'. The required value must be with the data types [13:1024] inclusively"))
	})

	t.Run("It returns an error if the type is outside the min range", func(t *testing.T) {
		err := EncapsulatedValue{Type: T_int, Data: int(42)}.Validate(T_uint8, T_uint16)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("Type: invalid value '10'. The required value must be with the data types [1:2] inclusively"))
	})
}
