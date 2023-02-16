package datastructures_test

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/DanLavine/willow/internal/datastructures"
	. "github.com/onsi/gomega"
)

func TestIDGenerator_Add(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("adds values and returns ids in chronilogical order", func(t *testing.T) {
		itemTracker := datastructures.NewIDTree()

		for i := 1; i <= 1024; i++ {
			g.Expect(itemTracker.Add(i)).To(Equal(uint64(i)))
		}
	})

	t.Run("replaces nodes at proper hights on the left", func(t *testing.T) {
		itemTracker := datastructures.NewIDTree()
		/* generate a tree like so:
		height |						   ID
		h0 - 1 (first index in this row) |                1
		h1 - 2 (first index in this row) |        2					       3
		h2 - 4 (first index in this row) |		4			  6			   5				7
		h3 - 8 (first index in this row) |	8		12	10	14	 9	 13	  11  15
		*/
		for i := 1; i <= 15; i++ {
			g.Expect(itemTracker.Add(fmt.Sprintf("%d", i))).To(Equal(uint64(i)))
		}

		// remove leaf then node
		g.Expect(itemTracker.Remove(8)).To(Equal("8"))
		g.Expect(itemTracker.Remove(4)).To(Equal("4"))

		// re-add leaf, should add the missing node 4
		g.Expect(itemTracker.Add("new-8")).To(Equal(uint64(4)))
		g.Expect(itemTracker.Get(4)).To(Equal("new-8"))

		// tring to add 4, should then add the missing 8
		g.Expect(itemTracker.Add("new-4")).To(Equal(uint64(8)))
		g.Expect(itemTracker.Get(8)).To(Equal("new-4"))

		// we can go in the reverse order as well. remove the node, then the leaf
		g.Expect(itemTracker.Remove(4)).To(Equal("new-8"))
		g.Expect(itemTracker.Remove(8)).To(Equal("new-4"))

		// re-add the 4 index
		g.Expect(itemTracker.Add("new-4-2")).To(Equal(uint64(4)))
		g.Expect(itemTracker.Get(4)).To(Equal("new-4-2"))

		// re-add the 8 index
		g.Expect(itemTracker.Add("new-8-2")).To(Equal(uint64(8)))
		g.Expect(itemTracker.Get(8)).To(Equal("new-8-2"))

		// adding another index should go back up to 10
		g.Expect(itemTracker.Add("16")).To(Equal(uint64(16)))
		g.Expect(itemTracker.Get(16)).To(Equal("16"))
	})

	t.Run("replaces nodes at proper hights on the right", func(t *testing.T) {
		itemTracker := datastructures.NewIDTree()
		/* generate a tree like so:
		height |						   ID
		h0 - 1 (first index in this row) |                1
		h1 - 2 (first index in this row) |        2					       3
		h2 - 4 (first index in this row) |		4			  6			   5				7
		h3 - 8 (first index in this row) |	8		12	10	14	 9	 13	  11  15
		*/
		for i := 1; i <= 15; i++ {
			g.Expect(itemTracker.Add(fmt.Sprintf("%d", i))).To(Equal(uint64(i)))
		}

		// remove leaf then node
		g.Expect(itemTracker.Remove(15)).To(Equal("15"))
		g.Expect(itemTracker.Remove(7)).To(Equal("7"))

		// re-add leaf, should add the missing node 4
		g.Expect(itemTracker.Add("new-15")).To(Equal(uint64(7)))
		g.Expect(itemTracker.Get(7)).To(Equal("new-15"))

		// tring to add 4, should then add the missing 8
		g.Expect(itemTracker.Add("new-7")).To(Equal(uint64(15)))
		g.Expect(itemTracker.Get(15)).To(Equal("new-7"))

		// we can go in the reverse order as well. remove the node, then the leaf
		g.Expect(itemTracker.Remove(15)).To(Equal("new-7"))
		g.Expect(itemTracker.Remove(7)).To(Equal("new-15"))

		// re-add the 4 index
		g.Expect(itemTracker.Add("new-7-2")).To(Equal(uint64(7)))
		g.Expect(itemTracker.Get(7)).To(Equal("new-7-2"))

		// re-add the 8 index
		g.Expect(itemTracker.Add("new-15-2")).To(Equal(uint64(15)))
		g.Expect(itemTracker.Get(15)).To(Equal("new-15-2"))

		// adding another index should go back up to 10
		g.Expect(itemTracker.Add("16")).To(Equal(uint64(16)))
		g.Expect(itemTracker.Get(16)).To(Equal("16"))
	})

	t.Run("replaces nodes at proper indexes on an imbalanced tree", func(t *testing.T) {
		itemTracker := datastructures.NewIDTree()
		/* generate a tree like so:
		height |						   ID
		h0 - 1 (first index in this row) |                1
		h1 - 2 (first index in this row) |        2					       3
		h2 - 4 (first index in this row) |		4			  6			   5				7
		h3 - 8 (first index in this row) |	8		12	10	14	 9	 13	  11  15
		*/
		for i := 1; i <= 15; i++ {
			g.Expect(itemTracker.Add(fmt.Sprintf("%d", i))).To(Equal(uint64(i)))
		}

		// remove a bunch of various nodes, leave one max height
		// should still have 1->2->4->12
		// should still have 1->3(nil)->7(nil)->11
		// next index is at 6.. but thats not true. left has an improper balance now of +3
		g.Expect(itemTracker.Remove(8)).To(Equal("8"))
		g.Expect(itemTracker.Remove(3)).To(Equal("3"))
		g.Expect(itemTracker.Remove(6)).To(Equal("6"))
		g.Expect(itemTracker.Remove(5)).To(Equal("5"))
		g.Expect(itemTracker.Remove(7)).To(Equal("7"))
		g.Expect(itemTracker.Remove(10)).To(Equal("10"))
		g.Expect(itemTracker.Remove(14)).To(Equal("14"))
		g.Expect(itemTracker.Remove(9)).To(Equal("9"))
		g.Expect(itemTracker.Remove(13)).To(Equal("13"))
		g.Expect(itemTracker.Remove(15)).To(Equal("15"))

		g.Expect(itemTracker.Add("3")).To(Equal(uint64(3)))
		g.Expect(itemTracker.Add("5")).To(Equal(uint64(5)))
		g.Expect(itemTracker.Add("6")).To(Equal(uint64(6)))
		g.Expect(itemTracker.Add("7")).To(Equal(uint64(7)))
		g.Expect(itemTracker.Add("8")).To(Equal(uint64(8)))
		g.Expect(itemTracker.Add("9")).To(Equal(uint64(9)))
		g.Expect(itemTracker.Add("10")).To(Equal(uint64(10)))
		g.Expect(itemTracker.Add("13")).To(Equal(uint64(13)))
		g.Expect(itemTracker.Add("14")).To(Equal(uint64(14)))
		g.Expect(itemTracker.Add("15")).To(Equal(uint64(15)))
	})
}

func TestIDGenerator_Get(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("returns nil if nothing has been added", func(t *testing.T) {
		itemTracker := datastructures.NewIDTree()
		g.Expect(itemTracker.Get(1)).To(BeNil())
	})

	t.Run("returns the value at the desired index", func(t *testing.T) {
		itemTracker := datastructures.NewIDTree()

		for i := 1; i <= 1024; i++ {
			g.Expect(itemTracker.Add(fmt.Sprintf("%d", i))).To(Equal(uint64(i)))
		}

		for i := 1; i <= 1024; i++ {
			g.Expect(itemTracker.Get(uint64(i))).To(Equal(fmt.Sprintf("%d", i)))
		}
	})
}

func TestIDGenerator_Remove(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("returns nil if nothing has been added", func(t *testing.T) {
		itemTracker := datastructures.NewIDTree()
		g.Expect(itemTracker.Remove(1)).To(BeNil())
	})

	t.Run("clears the root if everything has been removed", func(t *testing.T) {
		itemTracker := datastructures.NewIDTree()

		g.Expect(itemTracker.Add("1")).To(Equal(uint64(1)))
		g.Expect(itemTracker.Add("2")).To(Equal(uint64(2)))
		g.Expect(itemTracker.Add("3")).To(Equal(uint64(3)))

		g.Expect(itemTracker.Remove(1)).To(Equal("1"))
		g.Expect(itemTracker.Remove(2)).To(Equal("2"))
		g.Expect(itemTracker.Remove(3)).To(Equal("3"))
	})

	t.Run("returns the value at the desired index", func(t *testing.T) {
		itemTracker := datastructures.NewIDTree()

		for i := 1; i <= 1024; i++ {
			g.Expect(itemTracker.Add(fmt.Sprintf("%d", i))).To(Equal(uint64(i)))
		}

		for i := 1; i <= 1024; i++ {
			g.Expect(itemTracker.Remove(uint64(i))).To(Equal(fmt.Sprintf("%d", i)))
		}
	})

	t.Run("removing a left leaf node can be replaced properly", func(t *testing.T) {
		itemTracker := datastructures.NewIDTree()

		/* generate a tree like so:
		 *
		 *                1
		 *         2						3
		 *		4				6			5				7
		 *	8							9
		 */
		for i := 1; i <= 9; i++ {
			g.Expect(itemTracker.Add(fmt.Sprintf("%d", i))).To(Equal(uint64(i)))
		}

		// remove leaf then node
		g.Expect(itemTracker.Remove(8)).To(Equal("8"))

		// tring to add 8, should then add the missing 8
		g.Expect(itemTracker.Add("new-8")).To(Equal(uint64(8)))
		g.Expect(itemTracker.Get(8)).To(Equal("new-8"))

		// adding another index should go back up to 10
		g.Expect(itemTracker.Add("10")).To(Equal(uint64(10)))
		g.Expect(itemTracker.Get(10)).To(Equal("10"))
	})

	t.Run("removing a right leaf node can be replaced properly", func(t *testing.T) {
		itemTracker := datastructures.NewIDTree()

		/* generate a tree like so:
		 *
		 *                1
		 *         2						3
		 *		4				6			5				7
		 *	8							9
		 */
		for i := 1; i <= 9; i++ {
			g.Expect(itemTracker.Add(fmt.Sprintf("%d", i))).To(Equal(uint64(i)))
		}

		// remove leaf then node
		g.Expect(itemTracker.Remove(7)).To(Equal("7"))

		// tring to add 8, should then add the missing 8
		g.Expect(itemTracker.Add("new-7")).To(Equal(uint64(7)))
		g.Expect(itemTracker.Get(7)).To(Equal("new-7"))

		// adding another index should go back up to 10
		g.Expect(itemTracker.Add("10")).To(Equal(uint64(10)))
		g.Expect(itemTracker.Get(10)).To(Equal("10"))
	})

	t.Run("removing a left child and parent allow fo indexes to still be recreated properly", func(t *testing.T) {
		itemTracker := datastructures.NewIDTree()

		/* generate a tree like so:
		 *
		 *                1
		 *         2						3
		 *		4				6			5				7
		 *	8							9
		 */
		for i := 1; i <= 9; i++ {
			g.Expect(itemTracker.Add(fmt.Sprintf("%d", i))).To(Equal(uint64(i)))
		}

		// remove leaf then node
		g.Expect(itemTracker.Remove(8)).To(Equal("8"))
		g.Expect(itemTracker.Remove(4)).To(Equal("4"))

		// re-add leaf, should add the missing node 4
		g.Expect(itemTracker.Add("new-8")).To(Equal(uint64(4)))
		g.Expect(itemTracker.Get(4)).To(Equal("new-8"))

		// tring to add 4, should then add the missing 8
		g.Expect(itemTracker.Add("new-4")).To(Equal(uint64(8)))
		g.Expect(itemTracker.Get(8)).To(Equal("new-4"))

		// we can go in the reverse order as well. remove the node, then the leaf
		g.Expect(itemTracker.Remove(4)).To(Equal("new-8"))
		g.Expect(itemTracker.Remove(8)).To(Equal("new-4"))

		// re-add the 4 index
		g.Expect(itemTracker.Add("new-4-2")).To(Equal(uint64(4)))
		g.Expect(itemTracker.Get(4)).To(Equal("new-4-2"))

		// re-add the 8 index
		g.Expect(itemTracker.Add("new-8-2")).To(Equal(uint64(8)))
		g.Expect(itemTracker.Get(8)).To(Equal("new-8-2"))

		// adding another index should go back up to 10
		g.Expect(itemTracker.Add("10")).To(Equal(uint64(10)))
		g.Expect(itemTracker.Get(10)).To(Equal("10"))
	})

	t.Run("removing a right child and parent allow fo indexes to still be recreated properly", func(t *testing.T) {
		itemTracker := datastructures.NewIDTree()

		/* generate a tree like so:
		 *
		 *                1
		 *         2						3
		 *		4				6			5				7
		 *	8							9
		 */
		for i := 1; i <= 9; i++ {
			g.Expect(itemTracker.Add(fmt.Sprintf("%d", i))).To(Equal(uint64(i)))
		}

		// remove leaf then node
		g.Expect(itemTracker.Remove(7)).To(Equal("7"))
		g.Expect(itemTracker.Remove(3)).To(Equal("3"))

		// re-add leaf, should add the missing node 3
		g.Expect(itemTracker.Add("new-7")).To(Equal(uint64(3)))
		g.Expect(itemTracker.Get(3)).To(Equal("new-7"))

		// tring to add 7, should then add the missing 7
		g.Expect(itemTracker.Add("new-3")).To(Equal(uint64(7)))
		g.Expect(itemTracker.Get(7)).To(Equal("new-3"))

		// we can go in the reverse order as well. remove the node, then the leaf
		g.Expect(itemTracker.Remove(3)).To(Equal("new-7"))
		g.Expect(itemTracker.Remove(7)).To(Equal("new-3"))

		// re-add the 43 index
		g.Expect(itemTracker.Add("new-3-2")).To(Equal(uint64(3)))
		g.Expect(itemTracker.Get(3)).To(Equal("new-3-2"))

		// re-add the 8 index
		g.Expect(itemTracker.Add("new-7-2")).To(Equal(uint64(7)))
		g.Expect(itemTracker.Get(7)).To(Equal("new-7-2"))

		// adding another index should go back up to 10
		g.Expect(itemTracker.Add("10")).To(Equal(uint64(10)))
		g.Expect(itemTracker.Get(10)).To(Equal("10"))
	})

	t.Run("removing an index means the next add can re-insert to that empty index", func(t *testing.T) {
		itemTracker := datastructures.NewIDTree()

		for i := 1; i <= 1024; i++ {
			g.Expect(itemTracker.Add(fmt.Sprintf("%d", i))).To(Equal(uint64(i)))
		}

		rand.Seed(time.Now().UnixNano())
		for i := 0; i < 100; i++ {
			randInt := rand.Intn(1024) + 1
			g.Expect(itemTracker.Remove(uint64(randInt))).To(Equal(fmt.Sprintf("%d", randInt)))
			g.Expect(itemTracker.Add(fmt.Sprintf("%d", randInt))).To(Equal(uint64(randInt)))
		}
	})

	t.Run("removing the same index multiple times, does not mess up insert count", func(t *testing.T) {
		itemTracker := datastructures.NewIDTree()

		/* generate a tree like so:
		 *
		 *                1
		 *         2						3
		 *		4				6			5				7
		 *	8							9
		 */
		for i := 1; i <= 9; i++ {
			g.Expect(itemTracker.Add(fmt.Sprintf("%d", i))).To(Equal(uint64(i)))
		}

		// remove leaf then node
		g.Expect(itemTracker.Remove(8)).To(Equal("8"))
		g.Expect(itemTracker.Remove(8)).To(BeNil())

		// re-add leaf
		g.Expect(itemTracker.Add("new-8")).To(Equal(uint64(8)))

		// adding another index should go back up to 10
		g.Expect(itemTracker.Add("10")).To(Equal(uint64(10)))
		g.Expect(itemTracker.Get(10)).To(Equal("10"))

		// remove leaf then node
		g.Expect(itemTracker.Remove(7)).To(Equal("7"))
		g.Expect(itemTracker.Remove(7)).To(BeNil())

		// re-add leaf
		g.Expect(itemTracker.Add("new-7")).To(Equal(uint64(7)))

		// adding another index should go back up to 11
		g.Expect(itemTracker.Add("11")).To(Equal(uint64(11)))
		g.Expect(itemTracker.Get(11)).To(Equal("11"))
	})
}
