package datatypes

type Int int

func (i Int) Less(compareKey any) bool {
	return i < compareKey.(Int)
}

type Ints []Int

func (ints Ints) Pop() (CompareType, EnumerableCompareType) {
	return ints[0], ints[1:]
}

func (ints Ints) Len() int {
	return len(ints)
}
