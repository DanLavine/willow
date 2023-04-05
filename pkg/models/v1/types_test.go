package v1

import (
	"fmt"
	"testing"

	"github.com/DanLavine/willow/internal/datastructures"
	. "github.com/onsi/gomega"
)

func TestStrings_Sort(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("properly sorts a list of strings", func(t *testing.T) {
		strings := Strings{"b", "c", "d", "a"}
		strings.Sort()

		g.Expect(strings).To(Equal(Strings{"a", "b", "c", "d"}))
	})
}

func TestStrings_Each(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("runs the callback function for each element in the Strings", func(t *testing.T) {
		strings := Strings{"b", "c", "d", "a"}

		count := 0
		seenTree := Strings{}
		callback := func(index int, key datastructures.TreeKey) bool {
			g.Expect(index).To(Equal(count))
			seenTree = append(seenTree, String(fmt.Sprintf("%v", key)))
			count++
			return true
		}

		strings.Each(callback)
		g.Expect(seenTree).To(Equal(strings))
	})

	t.Run("the loop stops processing if the callback returns false", func(t *testing.T) {
		strings := Strings{"b", "c", "d", "a"}

		count := 0
		callback := func(index int, key datastructures.TreeKey) bool {
			count++
			return false
		}

		strings.Each(callback)
		g.Expect(count).To(Equal(1))
	})
}
