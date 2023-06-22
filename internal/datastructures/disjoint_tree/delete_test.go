package disjointtree

//import (
//	"testing"
//
//	. "github.com/DanLavine/willow/internal/datastructures/disjoint_tree/testhelpers"
//	"github.com/DanLavine/willow/pkg/models/datatypes"
//	. "github.com/onsi/gomega"
//)
//
//func TestDisjointTree_Delete_Params(t *testing.T) {
//	g := NewGomegaWithT(t)
//
//	t.Run("returns an error if keys are nil", func(t *testing.T) {
//		disjointTree := New()
//		err := disjointTree.Delete(nil, nil)
//		g.Expect(err).To(HaveOccurred())
//		g.Expect(err.Error()).To(Equal("EnumberableTreeKeys must have at least 1 element"))
//	})
//
//	t.Run("returns an error if there are no keys", func(t *testing.T) {
//		disjointTree := New()
//		err := disjointTree.Delete(datatypes.Ints{}, nil)
//		g.Expect(err).To(HaveOccurred())
//		g.Expect(err.Error()).To(Equal("EnumberableTreeKeys must have at least 1 element"))
//	})
//}
//
//func TestDisjointTree_Delete(t *testing.T) {
//	g := NewGomegaWithT(t)
//
//	setup := func(g *GomegaWithT) *disjointTree {
//		disjointTree := New()
//
//		_, err := disjointTree.CreateOrFind(datatypes.Ints{1}, nil, NewBTreeTester("1"))
//		g.Expect(err).ToNot(HaveOccurred())
//
//		return disjointTree
//	}
//
//	t.Run("does not delete the value if canDelete returns false", func(t *testing.T) {
//		disjointTree := setup(g)
//		err := disjointTree.Delete(datatypes.Ints{1}, func(item any) bool { return false })
//		g.Expect(err).ToNot(HaveOccurred())
//
//		g.Expect(disjointTree.tree).ToNot(BeNil())
//
//		value, err := disjointTree.Find(datatypes.Ints{1}, nil)
//		g.Expect(err).ToNot(HaveOccurred())
//		g.Expect(value).ToNot(BeNil())
//		g.Expect(value.(*BTreeTester).Value).To(Equal("1"))
//	})
//
//	t.Run("does delete the value if canDelete returns true", func(t *testing.T) {
//		disjointTree := setup(g)
//		err := disjointTree.Delete(datatypes.Ints{1}, func(item any) bool { return true })
//		g.Expect(err).ToNot(HaveOccurred())
//
//		g.Expect(disjointTree.tree).To(BeNil())
//	})
//}
//
//func TestDisjointTree_Delete_SingleDepth(t *testing.T) {
//	g := NewGomegaWithT(t)
//
//	setup := func(g *GomegaWithT) *disjointTree {
//		disjointTree := New()
//
//		_, err := disjointTree.CreateOrFind(datatypes.Ints{1}, nil, NewBTreeTester("1"))
//		g.Expect(err).ToNot(HaveOccurred())
//
//		return disjointTree
//	}
//
//	t.Run("deleting all values sets the root value to nil", func(t *testing.T) {
//		disjointTree := setup(g)
//		err := disjointTree.Delete(datatypes.Ints{1}, nil)
//		g.Expect(err).ToNot(HaveOccurred())
//
//		g.Expect(disjointTree.tree).To(BeNil())
//	})
//
//	t.Run("when there are multiple keys", func(t *testing.T) {
//		setup = func(g *GomegaWithT) *disjointTree {
//			disjointTree := New()
//
//			_, err := disjointTree.CreateOrFind(datatypes.Ints{1}, nil, NewBTreeTester("1"))
//			g.Expect(err).ToNot(HaveOccurred())
//
//			_, err = disjointTree.CreateOrFind(datatypes.Ints{2}, nil, NewBTreeTester("2"))
//			g.Expect(err).ToNot(HaveOccurred())
//
//			return disjointTree
//		}
//
//		t.Run("it keeps the tree when there are still values", func(t *testing.T) {
//			disjointTree := setup(g)
//			err := disjointTree.Delete(datatypes.Ints{1}, nil)
//			g.Expect(err).ToNot(HaveOccurred())
//
//			value, err := disjointTree.Find(datatypes.Ints{1}, nil)
//			g.Expect(err).ToNot(HaveOccurred())
//			g.Expect(value).To(BeNil())
//
//			value, err = disjointTree.Find(datatypes.Ints{2}, nil)
//			g.Expect(err).ToNot(HaveOccurred())
//			g.Expect(value).ToNot(BeNil())
//			g.Expect(value.(*BTreeTester).Value).To(Equal("2"))
//		})
//
//		t.Run("it sets the root value to nil when all values are removed", func(t *testing.T) {
//			disjointTree := setup(g)
//			g.Expect(disjointTree.Delete(datatypes.Ints{1}, nil)).ToNot(HaveOccurred())
//			g.Expect(disjointTree.Delete(datatypes.Ints{2}, nil)).ToNot(HaveOccurred())
//
//			g.Expect(disjointTree.tree).To(BeNil())
//		})
//	})
//}
//
//func TestDisjointTree_Delete_MultipleDepth(t *testing.T) {
//	g := NewGomegaWithT(t)
//
//	setup := func(g *GomegaWithT) *disjointTree {
//		disjointTree := New()
//
//		_, err := disjointTree.CreateOrFind(datatypes.Ints{1, 2}, nil, NewBTreeTester("2"))
//		g.Expect(err).ToNot(HaveOccurred())
//
//		return disjointTree
//	}
//
//	t.Run("deleting all child values sets the root value to nil", func(t *testing.T) {
//		disjointTree := setup(g)
//		err := disjointTree.Delete(datatypes.Ints{1, 2}, nil)
//		g.Expect(err).ToNot(HaveOccurred())
//
//		g.Expect(disjointTree.tree).To(BeNil())
//	})
//
//	t.Run("when deleting an internal node", func(t *testing.T) {
//		t.Run("removes the value of the internal node, but no children", func(t *testing.T) {
//			disjointTree := setup(g)
//			_, err := disjointTree.CreateOrFind(datatypes.Ints{1}, nil, NewBTreeTester("1"))
//			g.Expect(err).ToNot(HaveOccurred())
//
//			err = disjointTree.Delete(datatypes.Ints{1}, nil)
//			g.Expect(err).ToNot(HaveOccurred())
//
//			g.Expect(disjointTree.tree).ToNot(BeNil())
//
//			// innter node value should be gone
//			value, err := disjointTree.Find(datatypes.Ints{1}, nil)
//			g.Expect(err).ToNot(HaveOccurred())
//			g.Expect(value).To(BeNil())
//
//			// the child shuld still be available
//			value, err = disjointTree.Find(datatypes.Ints{1, 2}, nil)
//			g.Expect(err).ToNot(HaveOccurred())
//			g.Expect(value).ToNot(BeNil())
//			g.Expect(value.(*BTreeTester).Value).To(Equal("2"))
//		})
//
//		t.Run("deleting an internal node does nothing if it has children", func(t *testing.T) {
//			disjointTree := setup(g)
//			err := disjointTree.Delete(datatypes.Ints{1}, nil)
//			g.Expect(err).ToNot(HaveOccurred())
//			g.Expect(disjointTree.tree).ToNot(BeNil())
//
//			value, err := disjointTree.Find(datatypes.Ints{1, 2}, nil)
//			g.Expect(err).ToNot(HaveOccurred())
//			g.Expect(value).ToNot(BeNil())
//			g.Expect(value.(*BTreeTester).Value).To(Equal("2"))
//		})
//	})
//}
