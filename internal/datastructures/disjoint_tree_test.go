package datastructures

import (
	"testing"

	. "github.com/onsi/gomega"
)

type disjointTreeTester struct {
	value      string
	foundCount int
}

func (dtt *disjointTreeTester) OnFind() {
	dtt.foundCount++
}

func newDisjoinTreeTeseter(value string) func() (any, error) {
	return func() (any, error) {
		return &disjointTreeTester{
			value:      value,
			foundCount: 0,
		}, nil
	}
}

func TestDisjointTree_FindOrCreate(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("returns an error if keys are nil", func(t *testing.T) {
		disjointTree := NewDisjointTree()
		treeItem, err := disjointTree.FindOrCreate(nil, "", newDisjoinTreeTeseter("boop"))
		g.Expect(treeItem).To(BeNil())
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(Equal("EnumberableTreeKeys must have at least 1 element"))
	})

	t.Run("returns an error if onCreate callback is nil", func(t *testing.T) {
		disjointTree := NewDisjointTree()
		treeItem, err := disjointTree.FindOrCreate(EnumerableIntTreeKey{1}, "", nil)
		g.Expect(treeItem).To(BeNil())
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(Equal("Received a nil onCreate callback. Needs to not be nil"))
	})

	t.Run("it creates a single key at the root tree", func(t *testing.T) {
		disjointTree := NewDisjointTree()
		treeItem, err := disjointTree.FindOrCreate(EnumerableIntTreeKey{1}, "", newDisjoinTreeTeseter("fin"))
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(treeItem).ToNot(BeNil())
		g.Expect(treeItem.(*disjointTreeTester).value).To(Equal("fin"))

		disjointNode := disjointTree.tree.Find(IntTreeKey(1), "").(*disjointNode)
		g.Expect(disjointNode.value).To(Equal(treeItem))
		g.Expect(disjointNode.children).To(BeNil())
	})

	t.Run("multiple keys create a tree with children for each key", func(t *testing.T) {
		disjointTree := NewDisjointTree()
		treeItem, err := disjointTree.FindOrCreate(EnumerableIntTreeKey{1, 2}, "", newDisjoinTreeTeseter("second"))
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(treeItem).ToNot(BeNil())
		g.Expect(treeItem.(*disjointTreeTester).value).To(Equal("second"))

		node := disjointTree.tree.Find(IntTreeKey(1), "").(*disjointNode)
		g.Expect(node.value).To(BeNil())
		g.Expect(node.children).ToNot(BeNil())

		nodeChild := node.children.tree.Find(IntTreeKey(2), "").(*disjointNode)
		g.Expect(nodeChild.value).To(Equal(treeItem))
		g.Expect(nodeChild.children).To(BeNil())
	})

	t.Run("values can be assigned to tree elements that have no value", func(t *testing.T) {
		disjointTree := NewDisjointTree()
		_, err := disjointTree.FindOrCreate(EnumerableIntTreeKey{1, 2}, "", newDisjoinTreeTeseter("second"))
		g.Expect(err).ToNot(HaveOccurred())

		treeItem, err := disjointTree.FindOrCreate(EnumerableIntTreeKey{1}, "", newDisjoinTreeTeseter("first"))
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(treeItem).ToNot(BeNil())
		g.Expect(treeItem.(*disjointTreeTester).value).To(Equal("first"))

		node := disjointTree.tree.Find(IntTreeKey(1), "").(*disjointNode)
		g.Expect(node.value).To(Equal(treeItem))
		g.Expect(node.children).ToNot(BeNil())
	})

	t.Run("onFind is called for the value if it already exists", func(t *testing.T) {
		disjointTree := NewDisjointTree()
		disjointTree.FindOrCreate(EnumerableIntTreeKey{1, 2}, "", newDisjoinTreeTeseter("first"))

		treeItem, err := disjointTree.FindOrCreate(EnumerableIntTreeKey{1, 2}, "OnFind", newDisjoinTreeTeseter("second"))
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(treeItem).ToNot(BeNil())
		g.Expect(treeItem.(*disjointTreeTester).value).To(Equal("first"))
		g.Expect(treeItem.(*disjointTreeTester).foundCount).To(Equal(1))
	})
}

func TestDisjointTree_Find(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("it returns an error if the keys are nil", func(t *testing.T) {
		disjointTree := NewDisjointTree()
		treeItem, err := disjointTree.Find(nil, "")
		g.Expect(treeItem).To(BeNil())
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(Equal("EnumberableTreeKeys must have at least 1 element"))
	})

	t.Run("it returns an error if the keys don't have a length of 1", func(t *testing.T) {
		disjointTree := NewDisjointTree()
		treeItem, err := disjointTree.Find(EnumerableIntTreeKey{}, "")
		g.Expect(treeItem).To(BeNil())
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(Equal("EnumberableTreeKeys must have at least 1 element"))
	})

	t.Run("it returns an error if the keys cannot find an item", func(t *testing.T) {
		disjointTree := NewDisjointTree()
		treeItem, err := disjointTree.Find(EnumerableIntTreeKey{1}, "")
		g.Expect(treeItem).To(BeNil())
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(Equal("item not found"))
	})

	t.Run("it returns the item if it exists at the specific key", func(t *testing.T) {
		disjointTree := NewDisjointTree()
		disjointTree.FindOrCreate(EnumerableIntTreeKey{1, 2}, "", newDisjoinTreeTeseter("first"))

		treeItem, err := disjointTree.Find(EnumerableIntTreeKey{1, 2}, "")
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(treeItem).ToNot(BeNil())
	})

	t.Run("it calls the OnFind method if the item exists", func(t *testing.T) {
		disjointTree := NewDisjointTree()
		disjointTree.FindOrCreate(EnumerableIntTreeKey{1, 2}, "", newDisjoinTreeTeseter("first"))

		treeItem, err := disjointTree.Find(EnumerableIntTreeKey{1, 2}, "OnFind")
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(treeItem).ToNot(BeNil())
		g.Expect(treeItem.(*disjointTreeTester).value).To(Equal("first"))
		g.Expect(treeItem.(*disjointTreeTester).foundCount).To(Equal(1))
	})
}

func TestDisjointTree_Iterate(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("it runs the callback for only nodes with values", func(t *testing.T) {
		disjointTree := NewDisjointTree()
		disjointTree.FindOrCreate(EnumerableIntTreeKey{1, 2}, "", newDisjoinTreeTeseter("first"))

		count := 0
		iterator := func(value any) {
			treeTester := value.(*disjointTreeTester)
			g.Expect(treeTester.value).To(Equal("first"))
			count++
		}

		disjointTree.Iterate(iterator)
		g.Expect(count).To(Equal(1))
	})
}
