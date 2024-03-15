package helpers

func PointerOf[T any](value T) *T {
	return &value
}
