// normaly I like to have tests in their own package test the "API" of a class. But this
// I think it makes more sense to test the internals of the data structure
package datastructures

//import (
//	"fmt"
//	"sync"
//	"testing"
//
//	. "github.com/onsi/gomega"
//)
//
//func TestTagTree__AddValue_Single(t *testing.T) {
//	g := NewGomegaWithT(t)
//
//	t.Run("it accepts no tags", func(t *testing.T) {
//		tagTree := NewTagTree()
//		tagTree.AddValue(nil, "one")
//		tagTree.AddValue([]string{}, "two")
//
//		g.Expect(len(tagTree.values)).To(Equal(2))
//		g.Expect(tagTree.values[0]).To(Equal("one"))
//		g.Expect(tagTree.values[1]).To(Equal("two"))
//	})
//
//	t.Run("it can accept a single tag", func(t *testing.T) {
//		tagTree := NewTagTree()
//		tagTree.AddValue([]string{"a"}, "one")
//
//		g.Expect(len(tagTree.values)).To(Equal(0))
//		g.Expect(len(tagTree.children)).To(Equal(1))
//
//		child := tagTree.children[0]
//		g.Expect(child.tag).To(Equal("a"))
//		g.Expect(len(child.values)).To(Equal(1))
//		g.Expect(child.values[0]).To(Equal("one"))
//	})
//
//	t.Run("it can insert any number of children all to the left", func(t *testing.T) {
//		tagTree := NewTagTree()
//		tagTree.AddValue([]string{"b"}, "b")
//		tagTree.AddValue([]string{"a"}, "a")
//		tagTree.AddValue([]string{"0"}, "0")
//
//		g.Expect(len(tagTree.values)).To(Equal(0))
//		g.Expect(len(tagTree.children)).To(Equal(3))
//
//		// child 0
//		child := tagTree.children[0]
//		g.Expect(child.tag).To(Equal("0"))
//		g.Expect(len(child.values)).To(Equal(1))
//		g.Expect(child.values[0]).To(Equal("0"))
//
//		// child 1
//		child = tagTree.children[1]
//		g.Expect(child.tag).To(Equal("a"))
//		g.Expect(len(child.values)).To(Equal(1))
//		g.Expect(child.values[0]).To(Equal("a"))
//
//		// child 2
//		child = tagTree.children[2]
//		g.Expect(child.tag).To(Equal("b"))
//		g.Expect(len(child.values)).To(Equal(1))
//		g.Expect(child.values[0]).To(Equal("b"))
//	})
//
//	t.Run("it can insert any number of children all to the right", func(t *testing.T) {
//		tagTree := NewTagTree()
//		tagTree.AddValue([]string{"0"}, "0")
//		tagTree.AddValue([]string{"a"}, "a")
//		tagTree.AddValue([]string{"b"}, "b")
//
//		g.Expect(len(tagTree.values)).To(Equal(0))
//		g.Expect(len(tagTree.children)).To(Equal(3))
//
//		// child 0
//		child := tagTree.children[0]
//		g.Expect(child.tag).To(Equal("0"))
//		g.Expect(len(child.values)).To(Equal(1))
//		g.Expect(child.values[0]).To(Equal("0"))
//
//		// child 1
//		child = tagTree.children[1]
//		g.Expect(child.tag).To(Equal("a"))
//		g.Expect(len(child.values)).To(Equal(1))
//		g.Expect(child.values[0]).To(Equal("a"))
//
//		// child 2
//		child = tagTree.children[2]
//		g.Expect(child.tag).To(Equal("b"))
//		g.Expect(len(child.values)).To(Equal(1))
//		g.Expect(child.values[0]).To(Equal("b"))
//	})
//
//	t.Run("it can insert a middle index left", func(t *testing.T) {
//		tagTree := NewTagTree()
//		tagTree.AddValue([]string{"b"}, "b")
//		tagTree.AddValue([]string{"0"}, "0")
//		tagTree.AddValue([]string{"a"}, "a")
//
//		g.Expect(len(tagTree.values)).To(Equal(0))
//		g.Expect(len(tagTree.children)).To(Equal(3))
//
//		// child 0
//		child := tagTree.children[0]
//		g.Expect(child.tag).To(Equal("0"))
//		g.Expect(len(child.values)).To(Equal(1))
//		g.Expect(child.values[0]).To(Equal("0"))
//
//		// child 1
//		child = tagTree.children[1]
//		g.Expect(child.tag).To(Equal("a"))
//		g.Expect(len(child.values)).To(Equal(1))
//		g.Expect(child.values[0]).To(Equal("a"))
//
//		// child 2
//		child = tagTree.children[2]
//		g.Expect(child.tag).To(Equal("b"))
//		g.Expect(len(child.values)).To(Equal(1))
//		g.Expect(child.values[0]).To(Equal("b"))
//	})
//
//	// This is hard to test consitently since it happens on multi threaded checks
//	// so its more of a best effor atttempt that shoudn't be flaky.
//	t.Run("it can insert a middle index right", func(t *testing.T) {
//		wg := new(sync.WaitGroup)
//		tagTree := NewTagTree()
//
//		for i := 0; i < 1024; i++ {
//			wg.Add(1)
//			go func(num int, wg *sync.WaitGroup) {
//				defer wg.Done()
//				tagTree.AddValue([]string{fmt.Sprintf("%d", num)}, num)
//			}(i, wg)
//		}
//
//		wg.Wait()
//		g.Expect(len(tagTree.values)).To(Equal(0))
//		g.Expect(len(tagTree.children)).To(Equal(1024))
//	})
//}
//
//func TestTagTree_AddValue_nested(t *testing.T) {
//	g := NewGomegaWithT(t)
//
//	t.Run("it can accept a slice of values", func(t *testing.T) {
//		tagTree := NewTagTree()
//		tagTree.AddValue([]string{"a", "b", "c"}, "one")
//
//		g.Expect(len(tagTree.values)).To(Equal(0))
//		g.Expect(len(tagTree.children)).To(Equal(1))
//
//		child := tagTree.children[0]
//		g.Expect(child.tag).To(Equal("a"))
//		g.Expect(len(child.values)).To(Equal(0))
//		g.Expect(len(child.children)).To(Equal(1))
//
//		child = child.children[0]
//		g.Expect(child.tag).To(Equal("b"))
//		g.Expect(len(child.values)).To(Equal(0))
//		g.Expect(len(child.children)).To(Equal(1))
//
//		child = child.children[0]
//		g.Expect(child.tag).To(Equal("c"))
//		g.Expect(len(child.values)).To(Equal(1))
//		g.Expect(len(child.children)).To(Equal(0))
//		g.Expect(child.values[0]).To(Equal("one"))
//	})
//}
