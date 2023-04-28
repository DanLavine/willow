package helpers_test

import (
	"fmt"
	"reflect"
	"runtime"
	"testing"

	"github.com/DanLavine/willow/internal/helpers"
	"github.com/DanLavine/willow/pkg/models/datatypes"
	. "github.com/onsi/gomega"
)

func MatchOnce(g *GomegaWithT, group []datatypes.Strings, val datatypes.Strings) {
	count := 0
	for _, groupVal := range group {
		if reflect.DeepEqual(groupVal, val) {
			count++
		}
	}

	_, _, line, _ := runtime.Caller(1)
	g.Expect(count).To(Equal(1), fmt.Sprintf("line: %d, vale: %s", line, val))
}

func TestGenerateGroupPairs(t *testing.T) {
	g := NewGomegaWithT(t)

	group := datatypes.Strings{"a", "b", "c", "d", "e"}

	t.Run("it returns all individual elements", func(t *testing.T) {
		generatedGroup := helpers.GenerateGroupPairs(group)
		MatchOnce(g, generatedGroup, datatypes.Strings{"a"})
		MatchOnce(g, generatedGroup, datatypes.Strings{"b"})
		MatchOnce(g, generatedGroup, datatypes.Strings{"c"})
		MatchOnce(g, generatedGroup, datatypes.Strings{"d"})
		MatchOnce(g, generatedGroup, datatypes.Strings{"e"})
	})

	t.Run("it returns all 2 pair elements", func(t *testing.T) {
		generatedGroup := helpers.GenerateGroupPairs(group)

		// a group
		MatchOnce(g, generatedGroup, datatypes.Strings{"a", "b"})
		MatchOnce(g, generatedGroup, datatypes.Strings{"a", "c"})
		MatchOnce(g, generatedGroup, datatypes.Strings{"a", "d"})
		MatchOnce(g, generatedGroup, datatypes.Strings{"a", "e"})

		// b group
		MatchOnce(g, generatedGroup, datatypes.Strings{"b", "c"})
		MatchOnce(g, generatedGroup, datatypes.Strings{"b", "d"})
		MatchOnce(g, generatedGroup, datatypes.Strings{"b", "e"})

		// c group
		MatchOnce(g, generatedGroup, datatypes.Strings{"c", "d"})
		MatchOnce(g, generatedGroup, datatypes.Strings{"c", "e"})

		// d group
		MatchOnce(g, generatedGroup, datatypes.Strings{"d", "e"})
	})

	t.Run("it returns all 3 pair elements", func(t *testing.T) {
		generatedGroup := helpers.GenerateGroupPairs(group)

		// a group
		MatchOnce(g, generatedGroup, datatypes.Strings{"a", "b", "c"})
		MatchOnce(g, generatedGroup, datatypes.Strings{"a", "b", "d"})
		MatchOnce(g, generatedGroup, datatypes.Strings{"a", "b", "e"})
		MatchOnce(g, generatedGroup, datatypes.Strings{"a", "c", "d"})
		MatchOnce(g, generatedGroup, datatypes.Strings{"a", "c", "e"})
		MatchOnce(g, generatedGroup, datatypes.Strings{"a", "d", "e"})

		// b group
		MatchOnce(g, generatedGroup, datatypes.Strings{"b", "c", "d"})
		MatchOnce(g, generatedGroup, datatypes.Strings{"b", "c", "e"})
		MatchOnce(g, generatedGroup, datatypes.Strings{"b", "d", "e"})

		// c group
		MatchOnce(g, generatedGroup, datatypes.Strings{"c", "d", "e"})
	})

	t.Run("it returns all 4 pair elements", func(t *testing.T) {
		generatedGroup := helpers.GenerateGroupPairs(group)

		// a group
		MatchOnce(g, generatedGroup, datatypes.Strings{"a", "b", "c", "d"})
		MatchOnce(g, generatedGroup, datatypes.Strings{"a", "b", "c", "e"})
		MatchOnce(g, generatedGroup, datatypes.Strings{"a", "b", "d", "e"})
		MatchOnce(g, generatedGroup, datatypes.Strings{"a", "c", "d", "e"})

		// b group
		MatchOnce(g, generatedGroup, datatypes.Strings{"b", "c", "d", "e"})
	})

	t.Run("it returns all 5 pair elements", func(t *testing.T) {
		generatedGroup := helpers.GenerateGroupPairs(group)

		// a group
		MatchOnce(g, generatedGroup, datatypes.Strings{"a", "b", "c", "d", "e"})
	})

	t.Run("it has the proper size", func(t *testing.T) {
		generatedGroup := helpers.GenerateGroupPairs(group)

		// also matches total number of tests above
		g.Expect(len(generatedGroup)).To(Equal(31))
	})
}
