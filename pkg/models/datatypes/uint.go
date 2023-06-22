package datatypes

func Uint8(value uint8) EncapsulatedData {
	return EncapsulatedData{
		DataType: T_uint8,
		Value:    value,
	}
}

func Uint16(value uint16) EncapsulatedData {
	return EncapsulatedData{
		DataType: T_uint16,
		Value:    value,
	}
}

func Uint32(value uint32) EncapsulatedData {
	return EncapsulatedData{
		DataType: T_uint32,
		Value:    value,
	}
}

func Uint64(value uint64) EncapsulatedData {
	return EncapsulatedData{
		DataType: T_uint64,
		Value:    value,
	}
}

func Uint(value uint) EncapsulatedData {
	return EncapsulatedData{
		DataType: T_uint,
		Value:    value,
	}
}
