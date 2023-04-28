package disjointtree

import (
	"testing"

	. "github.com/DanLavine/willow/internal/datastructures/disjoint_tree/testhelpers"
	"github.com/DanLavine/willow/pkg/models/datatypes"
	. "github.com/onsi/gomega"
)

func TestDisjointTree_CreateOrFind_Params(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("returns an error if keys are nil", func(t *testing.T) {
		disjointTree := New()
		treeItem, err := disjointTree.CreateOrFind(nil, nil, NewBTreeTester("boop"))
		g.Expect(treeItem).To(BeNil())
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(Equal("EnumberableTreeKeys must have at least 1 element"))
	})

	t.Run("returns an error if OnCreate callback is nil", func(t *testing.T) {
		disjointTree := New()
		treeItem, err := disjointTree.CreateOrFind(datatypes.Ints{1}, nil, nil)
		g.Expect(treeItem).To(BeNil())
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(Equal("Received a nil onCreate callback. Needs to not be nil"))
	})
}

func TestDisjointTree_CreateOrFind(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("it creates a single key at the root tree", func(t *testing.T) {
		disjointTree := New()
		treeItem, err := disjointTree.CreateOrFind(datatypes.Ints{1}, nil, NewBTreeTester("fin"))
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(treeItem).ToNot(BeNil())
		g.Expect(treeItem.(*BTreeTester).Value).To(Equal("fin"))

		disjointNode := disjointTree.tree.Find(datatypes.Int(1), nil).(*disjointNode)
		g.Expect(disjointNode.value).To(Equal(treeItem))
		g.Expect(disjointNode.children).To(BeNil())
	})

	t.Run("multiple keys create a tree with children for each key", func(t *testing.T) {
		disjointTree := New()
		treeItem, err := disjointTree.CreateOrFind(datatypes.Ints{1, 2}, nil, NewBTreeTester("second"))
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(treeItem).ToNot(BeNil())
		g.Expect(treeItem.(*BTreeTester).Value).To(Equal("second"))

		node := disjointTree.tree.Find(datatypes.Int(1), nil).(*disjointNode)
		g.Expect(node.value).To(BeNil())
		g.Expect(node.children).ToNot(BeNil())

		nodeChild := node.children.tree.Find(datatypes.Int(2), nil).(*disjointNode)
		g.Expect(nodeChild.value).To(Equal(treeItem))
		g.Expect(nodeChild.children).To(BeNil())
	})

	t.Run("values can be assigned to tree elements that have no value", func(t *testing.T) {
		disjointTree := New()
		_, err := disjointTree.CreateOrFind(datatypes.Ints{1, 2}, nil, NewBTreeTester("second"))
		g.Expect(err).ToNot(HaveOccurred())

		treeItem, err := disjointTree.CreateOrFind(datatypes.Ints{1}, nil, NewBTreeTester("first"))
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(treeItem).ToNot(BeNil())
		g.Expect(treeItem.(*BTreeTester).Value).To(Equal("first"))

		node := disjointTree.tree.Find(datatypes.Int(1), nil).(*disjointNode)
		g.Expect(node.value).To(Equal(treeItem))
		g.Expect(node.children).ToNot(BeNil())
	})

	t.Run("onFind is called for the value if it already exists", func(t *testing.T) {
		disjointTree := New()
		disjointTree.CreateOrFind(datatypes.Ints{1, 2}, nil, NewBTreeTester("first"))

		treeItem, err := disjointTree.CreateOrFind(datatypes.Ints{1, 2}, OnFindTest, NewBTreeTester("second"))
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(treeItem).ToNot(BeNil())
		g.Expect(treeItem.(*BTreeTester).Value).To(Equal("first"))
		g.Expect(treeItem.(*BTreeTester).OnFindCount).To(Equal(1))
	})
}
