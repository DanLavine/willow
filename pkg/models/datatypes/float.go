package datatypes

func Float32(value float32) EncapsulatedData {
	return EncapsulatedData{
		DataType: T_float32,
		Value:    value,
	}
}

func Float64(value float64) EncapsulatedData {
	return EncapsulatedData{
		DataType: T_float64,
		Value:    value,
	}
}
