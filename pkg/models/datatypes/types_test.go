package datatypes

import (
	"testing"

	. "github.com/onsi/gomega"
)

// test struct To satisfy the CheckLess interface
type customTest struct {
	value string
}

func (at customTest) Less(item any) bool {
	return at.value < item.(customTest).value
}

func TestComparable_Less(t *testing.T) {
	g := NewGomegaWithT(t)

	// custom
	tCustom := Custom(customTest{value: "custom"})

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

	// nil
	tNil := Nil()

	t.Run("keys of the same values are always equal", func(t *testing.T) {
		// custom
		g.Expect(tCustom.Less(Custom(customTest{value: "custom"}))).To(BeFalse())
		g.Expect(Custom(customTest{value: "custom"}).Less(tCustom)).To(BeFalse())

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

		// nil
		g.Expect(tNil.Less(Nil())).To(BeFalse())
		g.Expect(Nil().Less(tNil)).To(BeFalse())
	})

	t.Run("keys of the same type have proper Less than values", func(t *testing.T) {
		// custom
		g.Expect(tCustom.Less(Custom(customTest{value: "customMore"}))).To(BeTrue())

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

		// nil
		g.Expect(tNil.Less(Nil())).To(BeFalse())
	})

	t.Run("keys of the same type respect greater than values", func(t *testing.T) {
		// custom
		g.Expect(tCustom.Less(Custom(customTest{value: "custo"}))).To(BeFalse())

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

		// nil
		g.Expect(tNil.Less(Nil())).To(BeFalse())
	})
}

func TestComparable_LessType(t *testing.T) {
	g := NewGomegaWithT(t)

	// custom
	tCustom := Custom(customTest{value: "custom"})

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

	// nil
	tNil := Nil()

	t.Run("type of the same values are always equal", func(t *testing.T) {
		// custom
		g.Expect(tCustom.LessType(Custom(customTest{value: "custom"}))).To(BeFalse())
		g.Expect(Custom(customTest{value: "custom"}).LessType(tCustom)).To(BeFalse())

		// ints
		//// int
		g.Expect(tInt.LessType(Int(1))).To(BeFalse())
		g.Expect(Int(1).LessType(tInt)).To(BeFalse())
		//// int8
		g.Expect(tInt8.LessType(Int8(1))).To(BeFalse())
		g.Expect(Int8(1).LessType(tInt8)).To(BeFalse())
		//// int16
		g.Expect(tInt16.LessType(Int16(1))).To(BeFalse())
		g.Expect(Int16(1).LessType(tInt16)).To(BeFalse())
		//// int32
		g.Expect(tInt32.LessType(Int32(1))).To(BeFalse())
		g.Expect(Int32(1).LessType(tInt32)).To(BeFalse())
		//// int64
		g.Expect(tInt64.LessType(Int64(1))).To(BeFalse())
		g.Expect(Int64(1).LessType(tInt64)).To(BeFalse())

		// uints
		//// uint
		g.Expect(tUint.LessType(Uint(1))).To(BeFalse())
		g.Expect(Uint(1).LessType(tUint)).To(BeFalse())
		//// int8
		g.Expect(tUint8.LessType(Uint8(1))).To(BeFalse())
		g.Expect(Uint8(1).LessType(tUint8)).To(BeFalse())
		//// int16
		g.Expect(tUint16.LessType(Uint16(1))).To(BeFalse())
		g.Expect(Uint16(1).LessType(tUint16)).To(BeFalse())
		//// int32
		g.Expect(tUint32.LessType(Uint32(1))).To(BeFalse())
		g.Expect(Uint32(1).LessType(tUint32)).To(BeFalse())
		//// int64
		g.Expect(tUint64.LessType(Uint64(1))).To(BeFalse())
		g.Expect(Uint64(1).LessType(tUint64)).To(BeFalse())

		// floats
		//// float32
		g.Expect(tFloat32.LessType(Float32(1))).To(BeFalse())
		g.Expect(Float32(1).LessType(tFloat32)).To(BeFalse())
		//// floatt64
		g.Expect(tFloat64.LessType(Float64(1))).To(BeFalse())
		g.Expect(Float64(1).LessType(tFloat64)).To(BeFalse())

		// string
		g.Expect(tString.LessType(String("a"))).To(BeFalse())
		g.Expect(String("a").LessType(tString)).To(BeFalse())

		// nil
		g.Expect(tNil.LessType(Nil())).To(BeFalse())
		g.Expect(Nil().LessType(tNil)).To(BeFalse())
	})
}

func TestComparable_LessValue(t *testing.T) {
	g := NewGomegaWithT(t)

	// custom
	tCustom := Custom(customTest{value: "custom"})

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

	// nil
	tNil := Nil()

	t.Run("panics if they types are not the same", func(t *testing.T) {
		// custom
		g.Expect(func() { tInt.LessValue(Custom(customTest{value: "custom"})) }).To(Panic())

		// ints
		g.Expect(func() { tInt.LessValue(Int64(1)) }).To(Panic())
		//// int16
		g.Expect(func() { tInt16.LessValue(Int(1)) }).To(Panic())
		//// int32
		g.Expect(func() { tInt32.LessValue(Int(1)) }).To(Panic())
		//// int64
		g.Expect(func() { tInt64.LessValue(Int(1)) }).To(Panic())

		// uints
		//// uint
		g.Expect(func() { tUint.LessValue(Uint64(1)) }).To(Panic())
		//// int8
		g.Expect(func() { tUint8.LessValue(Uint(1)) }).To(Panic())
		//// int16
		g.Expect(func() { tUint16.LessValue(Uint(1)) }).To(Panic())
		//// int32
		g.Expect(func() { tUint32.LessValue(Uint(1)) }).To(Panic())
		//// int64
		g.Expect(func() { tUint64.LessValue(Uint(1)) }).To(Panic())

		// floats
		//// float32
		g.Expect(func() { tFloat32.LessValue(Float64(1)) }).To(Panic())
		//// floatt64
		g.Expect(func() { tFloat64.LessValue(Float32(1)) }).To(Panic())

		// string
		g.Expect(func() { tString.LessValue(Int(4)) }).To(Panic())

		// nil
		g.Expect(func() { tNil.LessValue(Int(1)) }).To(Panic())
	})

	t.Run("compares values properly when they are the same type", func(t *testing.T) {
		// custom
		g.Expect(tCustom.LessValue(Custom(customTest{value: "custom"}))).To(BeFalse())
		g.Expect(Custom(customTest{value: "custom"}).LessValue(tCustom)).To(BeFalse())

		// ints
		//// int
		g.Expect(tInt.LessValue(Int(1))).To(BeFalse())
		g.Expect(Int(1).LessValue(tInt)).To(BeFalse())
		//// int8
		g.Expect(tInt8.LessValue(Int8(1))).To(BeFalse())
		g.Expect(Int8(1).LessValue(tInt8)).To(BeFalse())
		//// int16
		g.Expect(tInt16.LessValue(Int16(1))).To(BeFalse())
		g.Expect(Int16(1).LessValue(tInt16)).To(BeFalse())
		//// int32
		g.Expect(tInt32.LessValue(Int32(1))).To(BeFalse())
		g.Expect(Int32(1).LessValue(tInt32)).To(BeFalse())
		//// int64
		g.Expect(tInt64.LessValue(Int64(1))).To(BeFalse())
		g.Expect(Int64(1).LessValue(tInt64)).To(BeFalse())

		// uints
		//// uint
		g.Expect(tUint.LessValue(Uint(1))).To(BeFalse())
		g.Expect(Uint(1).LessValue(tUint)).To(BeFalse())
		//// int8
		g.Expect(tUint8.LessValue(Uint8(1))).To(BeFalse())
		g.Expect(Uint8(1).LessValue(tUint8)).To(BeFalse())
		//// int16
		g.Expect(tUint16.LessValue(Uint16(1))).To(BeFalse())
		g.Expect(Uint16(1).LessValue(tUint16)).To(BeFalse())
		//// int32
		g.Expect(tUint32.LessValue(Uint32(1))).To(BeFalse())
		g.Expect(Uint32(1).LessValue(tUint32)).To(BeFalse())
		//// int64
		g.Expect(tUint64.LessValue(Uint64(1))).To(BeFalse())
		g.Expect(Uint64(1).LessValue(tUint64)).To(BeFalse())

		// floats
		//// float32
		g.Expect(tFloat32.LessValue(Float32(1))).To(BeFalse())
		g.Expect(Float32(1).LessValue(tFloat32)).To(BeFalse())
		//// floatt64
		g.Expect(tFloat64.LessValue(Float64(1))).To(BeFalse())
		g.Expect(Float64(1).LessValue(tFloat64)).To(BeFalse())

		// string
		g.Expect(tString.LessValue(String("a"))).To(BeFalse())
		g.Expect(String("a").LessValue(tString)).To(BeFalse())

		// nil
		g.Expect(tNil.LessValue(Nil())).To(BeFalse())
		g.Expect(Nil().LessValue(tNil)).To(BeFalse())
	})
}

func TestComparable_Validate(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("it returns an error if the Value is nil", func(t *testing.T) {
		encapsulatedData := EncapsulatedData{DataType: T_int, Value: nil}

		err := encapsulatedData.Validate()
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(Equal("EncapsulatedData has a nil data Value"))
	})

	t.Run("it returns an error if the DataTypes is unknown", func(t *testing.T) {
		encapsulatedData := EncapsulatedData{DataType: DataType(1_000_000), Value: "something"}

		err := encapsulatedData.Validate()
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(Equal("EncapsulatedData has an unkown data type"))
	})

	t.Run("it returns an error if the DataType and value don't match", func(t *testing.T) {
		// custom
		err := EncapsulatedData{DataType: T_custom, Value: "bad"}.Validate()
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(Equal("EncapsulatedData has a custom data type which dos not implement: CheckLess"))

		// ints
		//// int
		err = EncapsulatedData{DataType: T_int, Value: "nope"}.Validate()
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(Equal("EncapsulatedData has an int data type, but the Value is a: string"))
		//// int8
		err = EncapsulatedData{DataType: T_int8, Value: "nope"}.Validate()
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(Equal("EncapsulatedData has an int8 data type, but the Value is a: string"))
		//// int16
		err = EncapsulatedData{DataType: T_int16, Value: "nope"}.Validate()
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(Equal("EncapsulatedData has an int16 data type, but the Value is a: string"))
		//// int32
		err = EncapsulatedData{DataType: T_int32, Value: "nope"}.Validate()
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(Equal("EncapsulatedData has an int32 data type, but the Value is a: string"))
		//// int64
		err = EncapsulatedData{DataType: T_int64, Value: "nope"}.Validate()
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(Equal("EncapsulatedData has an int64 data type, but the Value is a: string"))

		// uints
		//// uint
		err = EncapsulatedData{DataType: T_uint, Value: "nope"}.Validate()
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(Equal("EncapsulatedData has a uint data type, but the Value is a: string"))
		//// uint8
		err = EncapsulatedData{DataType: T_uint8, Value: "nope"}.Validate()
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(Equal("EncapsulatedData has a uint8 data type, but the Value is a: string"))
		//// uint16
		err = EncapsulatedData{DataType: T_uint16, Value: "nope"}.Validate()
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(Equal("EncapsulatedData has a uint16 data type, but the Value is a: string"))
		//// uint32
		err = EncapsulatedData{DataType: T_uint32, Value: "nope"}.Validate()
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(Equal("EncapsulatedData has a uint32 data type, but the Value is a: string"))
		//// uint64
		err = EncapsulatedData{DataType: T_uint64, Value: "nope"}.Validate()
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(Equal("EncapsulatedData has a uint64 data type, but the Value is a: string"))

		// floats
		//// float32
		err = EncapsulatedData{DataType: T_float32, Value: "nope"}.Validate()
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(Equal("EncapsulatedData has a float32 data type, but the Value is a: string"))
		//// float64
		err = EncapsulatedData{DataType: T_float64, Value: "nope"}.Validate()
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(Equal("EncapsulatedData has a float64 data type, but the Value is a: string"))

		// string
		err = EncapsulatedData{DataType: T_string, Value: int(42)}.Validate()
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(Equal("EncapsulatedData has a string data type, but the Value is a: int"))

		// nil
		err = EncapsulatedData{DataType: T_nil, Value: int(42)}.Validate()
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(Equal("EncapsulatedData has a 'nil' data type and requires the Value to be nil"))
	})
}
