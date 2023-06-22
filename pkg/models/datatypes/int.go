package datatypes

//type Int int
//
//func (i Int) Less(compareKey any) bool {
//	return i < compareKey.(Int)
//}
//
//type Ints []Int
//
//func (ints Ints) Pop() (CompareType, EnumerableCompareType) {
//	return ints[0], ints[1:]
//}
//
//func (ints Ints) Len() int {
//	return len(ints)
//}

func Int8(value int8) EncapsulatedData {
	return EncapsulatedData{
		DataType: T_int8,
		Value:    value,
	}
}

func Int16(value int16) EncapsulatedData {
	return EncapsulatedData{
		DataType: T_int16,
		Value:    value,
	}
}

func Int32(value int32) EncapsulatedData {
	return EncapsulatedData{
		DataType: T_int32,
		Value:    value,
	}
}

func Int64(value int64) EncapsulatedData {
	return EncapsulatedData{
		DataType: T_int64,
		Value:    value,
	}
}

func Int(value int) EncapsulatedData {
	return EncapsulatedData{
		DataType: T_int,
		Value:    value,
	}
}
