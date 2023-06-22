package datatypes

// Create a Comparable data type with internals hidden
func Any(value ComparableDataType) EncapsulatedData {
	return EncapsulatedData{
		DataType: T_any,
		Value:    value,
	}
}
