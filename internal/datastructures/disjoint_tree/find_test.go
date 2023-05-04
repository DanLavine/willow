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

//func TestDisjointTree_Search(t *testing.T) {
//	g := NewGomegaWithT(t)
//
//	t.Run("it returns an error if the keys are nil", func(t *testing.T) {
//		disjointTree := New()
//
//		searchResults, err := disjointTree.searchKeys(nil, nil)
//		g.Expect(err).ToNot(BeNil())
//		g.Expect(err.Error()).To(ContainSubstring("cannot have empty search keys"))
//		g.Expect(searchResults).To(BeNil())
//	})
//
//	t.Run("it returns an error if the keys are empty", func(t *testing.T) {
//		disjointTree := New()
//
//		searchResults, err := disjointTree.searchKeys(datatypes.Ints{}, nil)
//		g.Expect(err).ToNot(BeNil())
//		g.Expect(err.Error()).To(ContainSubstring("cannot have empty search keys"))
//		g.Expect(searchResults).To(BeNil())
//	})
//
//	t.Run("it returns an empty list if there are no results", func(t *testing.T) {
//		disjointTree := New()
//
//		searchResults, err := disjointTree.searchKeys(datatypes.Ints{1, 2}, nil)
//		g.Expect(err).ToNot(HaveOccurred())
//		g.Expect(searchResults).To(Equal(SearchResults))
//	})
//
//	t.Run("when the tree has a number of values inserted", func(t *testing.T) {
//		setup := func(g *GomegaWithT) *disjointTree {
//			disjointTree := New()
//
//			var err error
//			_, err = disjointTree.CreateOrFind(datatypes.Ints{1, 1}, nil, NewBTreeTester("1"))
//			g.Expect(err).ToNot(HaveOccurred())
//			_, err = disjointTree.CreateOrFind(datatypes.Ints{1, 2}, nil, NewBTreeTester("2"))
//			g.Expect(err).ToNot(HaveOccurred())
//			_, err = disjointTree.CreateOrFind(datatypes.Ints{1, 3}, nil, NewBTreeTester("3"))
//			g.Expect(err).ToNot(HaveOccurred())
//			_, err = disjointTree.CreateOrFind(datatypes.Ints{1, 4}, nil, NewBTreeTester("4"))
//			g.Expect(err).ToNot(HaveOccurred())
//			_, err = disjointTree.CreateOrFind(datatypes.Ints{1, 5, 20, 30}, nil, NewBTreeTester("30"))
//			g.Expect(err).ToNot(HaveOccurred())
//
//			return disjointTree
//		}
//
//		t.Run("it returns all the properly tagged values, and not internal nodes", func(t *testing.T) {
//			disjointTree := setup(g)
//
//			searchResults, err := disjointTree.searchKeys(datatypes.Ints{}, nil)
//			g.Expect(err).ToNot(HaveOccurred())
//
//			g.Expect(len(searchResults)).To(Equal(5))
//		})
//	})
//}
