package helpers_test

import (
	"fmt"
	"runtime"
	"testing"

	"github.com/DanLavine/willow/internal/helpers"
	. "github.com/onsi/gomega"
)

func MatchOnce(g *GomegaWithT, group []string, val string) {
	count := 0
	for _, groupVal := range group {
		if groupVal == val {
			count++
		}
	}

	_, _, line, _ := runtime.Caller(1)
	g.Expect(count).To(Equal(1), fmt.Sprintf("line: %d, vale: %s", line, val))
}

func TestGenerateStringPairs(t *testing.T) {
	g := NewGomegaWithT(t)

	group := []string{"a", "b", "c", "d", "e"}

	t.Run("it returns all individual elements", func(t *testing.T) {
		generatedGroup := helpers.GenerateStringPairs(group)
		MatchOnce(g, generatedGroup, "a")
		MatchOnce(g, generatedGroup, "b")
		MatchOnce(g, generatedGroup, "c")
		MatchOnce(g, generatedGroup, "d")
		MatchOnce(g, generatedGroup, "e")
	})

	t.Run("it returns all 2 pair elements", func(t *testing.T) {
		generatedGroup := helpers.GenerateStringPairs(group)

		// a group
		MatchOnce(g, generatedGroup, "ab")
		MatchOnce(g, generatedGroup, "ac")
		MatchOnce(g, generatedGroup, "ad")
		MatchOnce(g, generatedGroup, "ae")

		// b group
		MatchOnce(g, generatedGroup, "bc")
		MatchOnce(g, generatedGroup, "bd")
		MatchOnce(g, generatedGroup, "be")

		// c group
		MatchOnce(g, generatedGroup, "cd")
		MatchOnce(g, generatedGroup, "ce")

		// d group
		MatchOnce(g, generatedGroup, "de")
	})

	t.Run("it returns all 3 pair elements", func(t *testing.T) {
		generatedGroup := helpers.GenerateStringPairs(group)

		// a group
		MatchOnce(g, generatedGroup, "abc")
		MatchOnce(g, generatedGroup, "abd")
		MatchOnce(g, generatedGroup, "abe")
		MatchOnce(g, generatedGroup, "acd")
		MatchOnce(g, generatedGroup, "ace")
		MatchOnce(g, generatedGroup, "ade")

		// b group
		MatchOnce(g, generatedGroup, "bcd")
		MatchOnce(g, generatedGroup, "bce")
		MatchOnce(g, generatedGroup, "bde")

		// c group
		MatchOnce(g, generatedGroup, "cde")
	})

	t.Run("it returns all 4 pair elements", func(t *testing.T) {
		generatedGroup := helpers.GenerateStringPairs(group)

		// a group
		MatchOnce(g, generatedGroup, "abcd")
		MatchOnce(g, generatedGroup, "abce")
		MatchOnce(g, generatedGroup, "abde")
		MatchOnce(g, generatedGroup, "acde")

		// b group
		MatchOnce(g, generatedGroup, "bcde")
	})

	t.Run("it returns all 5 pair elements", func(t *testing.T) {
		generatedGroup := helpers.GenerateStringPairs(group)

		// a group
		MatchOnce(g, generatedGroup, "abcde")
	})

	t.Run("it has the proper size", func(t *testing.T) {
		generatedGroup := helpers.GenerateStringPairs(group)

		// also matches total number of tests above
		g.Expect(len(generatedGroup)).To(Equal(31))
	})
}
