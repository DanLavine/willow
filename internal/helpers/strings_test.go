package helpers_test

import (
	"fmt"
	"reflect"
	"runtime"
	"testing"

	"github.com/DanLavine/willow/internal/datastructures"
	"github.com/DanLavine/willow/internal/helpers"
	. "github.com/onsi/gomega"
)

func MatchOnce(g *GomegaWithT, group [][]datastructures.TreeKey, val []datastructures.TreeKey) {
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

	group := []string{"a", "b", "c", "d", "e"}

	t.Run("it returns all individual elements", func(t *testing.T) {
		generatedGroup := helpers.GenerateGroupPairs(group)
		MatchOnce(g, generatedGroup, []datastructures.TreeKey{datastructures.NewStringTreeKey("a")})
		MatchOnce(g, generatedGroup, []datastructures.TreeKey{datastructures.NewStringTreeKey("b")})
		MatchOnce(g, generatedGroup, []datastructures.TreeKey{datastructures.NewStringTreeKey("c")})
		MatchOnce(g, generatedGroup, []datastructures.TreeKey{datastructures.NewStringTreeKey("d")})
		MatchOnce(g, generatedGroup, []datastructures.TreeKey{datastructures.NewStringTreeKey("e")})
	})

	t.Run("it returns all 2 pair elements", func(t *testing.T) {
		generatedGroup := helpers.GenerateGroupPairs(group)

		// a group
		MatchOnce(g, generatedGroup, []datastructures.TreeKey{datastructures.NewStringTreeKey("a"), datastructures.NewStringTreeKey("b")})
		MatchOnce(g, generatedGroup, []datastructures.TreeKey{datastructures.NewStringTreeKey("a"), datastructures.NewStringTreeKey("c")})
		MatchOnce(g, generatedGroup, []datastructures.TreeKey{datastructures.NewStringTreeKey("a"), datastructures.NewStringTreeKey("d")})
		MatchOnce(g, generatedGroup, []datastructures.TreeKey{datastructures.NewStringTreeKey("a"), datastructures.NewStringTreeKey("e")})

		// b group
		MatchOnce(g, generatedGroup, []datastructures.TreeKey{datastructures.NewStringTreeKey("b"), datastructures.NewStringTreeKey("c")})
		MatchOnce(g, generatedGroup, []datastructures.TreeKey{datastructures.NewStringTreeKey("b"), datastructures.NewStringTreeKey("d")})
		MatchOnce(g, generatedGroup, []datastructures.TreeKey{datastructures.NewStringTreeKey("b"), datastructures.NewStringTreeKey("e")})

		// c group
		MatchOnce(g, generatedGroup, []datastructures.TreeKey{datastructures.NewStringTreeKey("c"), datastructures.NewStringTreeKey("d")})
		MatchOnce(g, generatedGroup, []datastructures.TreeKey{datastructures.NewStringTreeKey("c"), datastructures.NewStringTreeKey("e")})

		// d group
		MatchOnce(g, generatedGroup, []datastructures.TreeKey{datastructures.NewStringTreeKey("d"), datastructures.NewStringTreeKey("e")})
	})

	t.Run("it returns all 3 pair elements", func(t *testing.T) {
		generatedGroup := helpers.GenerateGroupPairs(group)

		// a group
		MatchOnce(g, generatedGroup, []datastructures.TreeKey{datastructures.NewStringTreeKey("a"), datastructures.NewStringTreeKey("b"), datastructures.NewStringTreeKey("c")})
		MatchOnce(g, generatedGroup, []datastructures.TreeKey{datastructures.NewStringTreeKey("a"), datastructures.NewStringTreeKey("b"), datastructures.NewStringTreeKey("d")})
		MatchOnce(g, generatedGroup, []datastructures.TreeKey{datastructures.NewStringTreeKey("a"), datastructures.NewStringTreeKey("b"), datastructures.NewStringTreeKey("e")})
		MatchOnce(g, generatedGroup, []datastructures.TreeKey{datastructures.NewStringTreeKey("a"), datastructures.NewStringTreeKey("c"), datastructures.NewStringTreeKey("d")})
		MatchOnce(g, generatedGroup, []datastructures.TreeKey{datastructures.NewStringTreeKey("a"), datastructures.NewStringTreeKey("c"), datastructures.NewStringTreeKey("e")})
		MatchOnce(g, generatedGroup, []datastructures.TreeKey{datastructures.NewStringTreeKey("a"), datastructures.NewStringTreeKey("d"), datastructures.NewStringTreeKey("e")})

		// b group
		MatchOnce(g, generatedGroup, []datastructures.TreeKey{datastructures.NewStringTreeKey("b"), datastructures.NewStringTreeKey("c"), datastructures.NewStringTreeKey("d")})
		MatchOnce(g, generatedGroup, []datastructures.TreeKey{datastructures.NewStringTreeKey("b"), datastructures.NewStringTreeKey("c"), datastructures.NewStringTreeKey("e")})
		MatchOnce(g, generatedGroup, []datastructures.TreeKey{datastructures.NewStringTreeKey("b"), datastructures.NewStringTreeKey("d"), datastructures.NewStringTreeKey("e")})

		// c group
		MatchOnce(g, generatedGroup, []datastructures.TreeKey{datastructures.NewStringTreeKey("c"), datastructures.NewStringTreeKey("d"), datastructures.NewStringTreeKey("e")})
	})

	t.Run("it returns all 4 pair elements", func(t *testing.T) {
		generatedGroup := helpers.GenerateGroupPairs(group)

		// a group
		MatchOnce(g, generatedGroup, []datastructures.TreeKey{datastructures.NewStringTreeKey("a"), datastructures.NewStringTreeKey("b"), datastructures.NewStringTreeKey("c"), datastructures.NewStringTreeKey("d")})
		MatchOnce(g, generatedGroup, []datastructures.TreeKey{datastructures.NewStringTreeKey("a"), datastructures.NewStringTreeKey("b"), datastructures.NewStringTreeKey("c"), datastructures.NewStringTreeKey("e")})
		MatchOnce(g, generatedGroup, []datastructures.TreeKey{datastructures.NewStringTreeKey("a"), datastructures.NewStringTreeKey("b"), datastructures.NewStringTreeKey("d"), datastructures.NewStringTreeKey("e")})
		MatchOnce(g, generatedGroup, []datastructures.TreeKey{datastructures.NewStringTreeKey("a"), datastructures.NewStringTreeKey("c"), datastructures.NewStringTreeKey("d"), datastructures.NewStringTreeKey("e")})

		// b group
		MatchOnce(g, generatedGroup, []datastructures.TreeKey{datastructures.NewStringTreeKey("b"), datastructures.NewStringTreeKey("c"), datastructures.NewStringTreeKey("d"), datastructures.NewStringTreeKey("e")})
	})

	t.Run("it returns all 5 pair elements", func(t *testing.T) {
		generatedGroup := helpers.GenerateGroupPairs(group)

		// a group
		MatchOnce(g, generatedGroup, []datastructures.TreeKey{datastructures.NewStringTreeKey("a"), datastructures.NewStringTreeKey("b"), datastructures.NewStringTreeKey("c"), datastructures.NewStringTreeKey("d"), datastructures.NewStringTreeKey("e")})
	})

	t.Run("it has the proper size", func(t *testing.T) {
		generatedGroup := helpers.GenerateGroupPairs(group)

		// also matches total number of tests above
		g.Expect(len(generatedGroup)).To(Equal(31))
	})
}
