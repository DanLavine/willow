package datatypes

func String(value string) EncapsulatedData {
	return EncapsulatedData{
		DataType: T_string,
		Value:    value,
	}
}
