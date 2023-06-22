package datatypes

func Nil() EncapsulatedData {
	return EncapsulatedData{
		DataType: T_nil,
		Value:    nil,
	}
}
