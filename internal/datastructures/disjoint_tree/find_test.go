package disjointtree

import (
	"testing"

	. "github.com/DanLavine/willow/internal/datastructures/disjoint_tree/testhelpers"
	"github.com/DanLavine/willow/pkg/models/datatypes"
	. "github.com/onsi/gomega"
)

func TestDisjointTree_Find(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("it returns an error if the keys are nil", func(t *testing.T) {
		disjointTree := New()
		treeItem, err := disjointTree.Find(nil, nil)
		g.Expect(treeItem).To(BeNil())
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(Equal("EnumberableTreeKeys must have at least 1 element"))
	})

	t.Run("it returns an error if the keys don't have a length of 1", func(t *testing.T) {
		disjointTree := New()
		treeItem, err := disjointTree.Find(datatypes.Ints{}, nil)
		g.Expect(treeItem).To(BeNil())
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(Equal("EnumberableTreeKeys must have at least 1 element"))
	})

	t.Run("it returns the item if it exists at the specific key", func(t *testing.T) {
		disjointTree := New()
		disjointTree.CreateOrFind(datatypes.Ints{1, 2}, nil, NewBTreeTester("first"))

		treeItem, err := disjointTree.Find(datatypes.Ints{1, 2}, nil)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(treeItem).ToNot(BeNil())
	})

	t.Run("it calls the OnFind method if the item exists", func(t *testing.T) {
		disjointTree := New()
		disjointTree.CreateOrFind(datatypes.Ints{1, 2}, nil, NewBTreeTester("first"))

		treeItem, err := disjointTree.Find(datatypes.Ints{1, 2}, OnFindTest)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(treeItem).ToNot(BeNil())
		g.Expect(treeItem.(*BTreeTester).Value).To(Equal("first"))
		g.Expect(treeItem.(*BTreeTester).OnFindCount).To(Equal(1))
	})
}
