package helpers_test

import (
	"fmt"
	"reflect"
	"runtime"
	"testing"

	"github.com/DanLavine/willow/internal/helpers"
	v1 "github.com/DanLavine/willow/pkg/models/v1"
	. "github.com/onsi/gomega"
)

func MatchOnce(g *GomegaWithT, group []v1.Strings, val v1.Strings) {
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

	group := v1.Strings{"a", "b", "c", "d", "e"}

	t.Run("it returns all individual elements", func(t *testing.T) {
		generatedGroup := helpers.GenerateGroupPairs(group)
		MatchOnce(g, generatedGroup, v1.Strings{"a"})
		MatchOnce(g, generatedGroup, v1.Strings{"b"})
		MatchOnce(g, generatedGroup, v1.Strings{"c"})
		MatchOnce(g, generatedGroup, v1.Strings{"d"})
		MatchOnce(g, generatedGroup, v1.Strings{"e"})
	})

	t.Run("it returns all 2 pair elements", func(t *testing.T) {
		generatedGroup := helpers.GenerateGroupPairs(group)

		// a group
		MatchOnce(g, generatedGroup, v1.Strings{"a", "b"})
		MatchOnce(g, generatedGroup, v1.Strings{"a", "c"})
		MatchOnce(g, generatedGroup, v1.Strings{"a", "d"})
		MatchOnce(g, generatedGroup, v1.Strings{"a", "e"})

		// b group
		MatchOnce(g, generatedGroup, v1.Strings{"b", "c"})
		MatchOnce(g, generatedGroup, v1.Strings{"b", "d"})
		MatchOnce(g, generatedGroup, v1.Strings{"b", "e"})

		// c group
		MatchOnce(g, generatedGroup, v1.Strings{"c", "d"})
		MatchOnce(g, generatedGroup, v1.Strings{"c", "e"})

		// d group
		MatchOnce(g, generatedGroup, v1.Strings{"d", "e"})
	})

	t.Run("it returns all 3 pair elements", func(t *testing.T) {
		generatedGroup := helpers.GenerateGroupPairs(group)

		// a group
		MatchOnce(g, generatedGroup, v1.Strings{"a", "b", "c"})
		MatchOnce(g, generatedGroup, v1.Strings{"a", "b", "d"})
		MatchOnce(g, generatedGroup, v1.Strings{"a", "b", "e"})
		MatchOnce(g, generatedGroup, v1.Strings{"a", "c", "d"})
		MatchOnce(g, generatedGroup, v1.Strings{"a", "c", "e"})
		MatchOnce(g, generatedGroup, v1.Strings{"a", "d", "e"})

		// b group
		MatchOnce(g, generatedGroup, v1.Strings{"b", "c", "d"})
		MatchOnce(g, generatedGroup, v1.Strings{"b", "c", "e"})
		MatchOnce(g, generatedGroup, v1.Strings{"b", "d", "e"})

		// c group
		MatchOnce(g, generatedGroup, v1.Strings{"c", "d", "e"})
	})

	t.Run("it returns all 4 pair elements", func(t *testing.T) {
		generatedGroup := helpers.GenerateGroupPairs(group)

		// a group
		MatchOnce(g, generatedGroup, v1.Strings{"a", "b", "c", "d"})
		MatchOnce(g, generatedGroup, v1.Strings{"a", "b", "c", "e"})
		MatchOnce(g, generatedGroup, v1.Strings{"a", "b", "d", "e"})
		MatchOnce(g, generatedGroup, v1.Strings{"a", "c", "d", "e"})

		// b group
		MatchOnce(g, generatedGroup, v1.Strings{"b", "c", "d", "e"})
	})

	t.Run("it returns all 5 pair elements", func(t *testing.T) {
		generatedGroup := helpers.GenerateGroupPairs(group)

		// a group
		MatchOnce(g, generatedGroup, v1.Strings{"a", "b", "c", "d", "e"})
	})

	t.Run("it has the proper size", func(t *testing.T) {
		generatedGroup := helpers.GenerateGroupPairs(group)

		// also matches total number of tests above
		g.Expect(len(generatedGroup)).To(Equal(31))
	})
}
