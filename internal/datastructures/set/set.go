package set

type Set[T comparable] interface {
	// Clear out the entire set by removing all elements
	Clear()

	// Add a value to the set
	Add(values T)

	// add a number of values to a set
	AddBulk(values []T)

	// Remove a value from the set
	Remove(value T)

	// Perform an intersection on the provided values
	Intersection(values []T)

	// Get all the values in the Set
	Values() []T

	// Get the number of values in the Set
	Size() int
}

type set[T comparable] struct {
	values map[T]struct{}
}

func New[T comparable](initValues ...T) *set[T] {
	initMap := map[T]struct{}{}
	for _, value := range initValues {
		initMap[value] = struct{}{}
	}

	return &set[T]{values: initMap}
}

func (s *set[T]) Add(value T) {
	s.values[value] = struct{}{}
}

func (s *set[T]) AddBulk(values []T) {
	for _, value := range values {
		s.values[value] = struct{}{}
	}
}

func (s *set[T]) Clear() {
	s.values = map[T]struct{}{}
}

func (s *set[T]) Intersection(values []T) {
	newValues := map[T]struct{}{}

	for _, value := range values {
		if _, ok := s.values[value]; ok {
			newValues[value] = struct{}{}
		}
	}

	s.values = newValues
}

func (s *set[T]) Remove(value T) {
	delete(s.values, value)
}

func (s *set[T]) Values() []T {
	values := []T{}
	for key, _ := range s.values {
		values = append(values, key)
	}

	return values
}

func (s *set[T]) Size() int {
	return len(s.values)
}
