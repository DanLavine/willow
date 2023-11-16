package datatypes

type EncapsulatedData interface {
	Validate() error

	DataType() DataType

	Value() any

	Less(item EncapsulatedData) bool

	LessType(item EncapsulatedData) bool

	LessValue(item EncapsulatedData) bool
}
