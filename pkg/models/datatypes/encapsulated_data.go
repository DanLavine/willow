package datatypes

type EncapsulatedData interface {
	DataType() DataType

	Value() any

	Validate() error

	Less(item EncapsulatedData) bool

	LessType(item EncapsulatedData) bool

	LessValue(item EncapsulatedData) bool
}
