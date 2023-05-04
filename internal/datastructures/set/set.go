package set

type Set interface {
	Clear()

	Add(values []uint64)

	Keep(values []uint64)

	Values() []uint64

	Len() int
}

type set struct {
	values map[uint64]struct{}
}

func New() *set {
	return &set{
		values: map[uint64]struct{}{},
	}
}

func (s *set) Add(values []uint64) {
	for _, value := range values {
		s.values[value] = struct{}{}
	}
}

func (s *set) Clear() {
	s.values = map[uint64]struct{}{}
}

func (s *set) Keep(valuesToKeep []uint64) {
	newValues := map[uint64]struct{}{}

	for _, valueToKeep := range valuesToKeep {
		if _, ok := s.values[valueToKeep]; ok {
			newValues[valueToKeep] = struct{}{}
		}
	}

	s.values = newValues
}

func (s *set) Values() []uint64 {
	values := []uint64{}
	for key, _ := range s.values {
		values = append(values, key)
	}

	return values
}

func (s *set) Len() int {
	return len(s.values)
}
