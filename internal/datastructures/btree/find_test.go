package btree

import (
	"fmt"
	"testing"

	. "github.com/DanLavine/willow/internal/datastructures/btree/testhelpers"
	"github.com/DanLavine/willow/pkg/models/datatypes"
	. "github.com/onsi/gomega"
)

func TestBTree_Find(t *testing.T) {
	g := NewGomegaWithT(t)

	setupTree := func(g *GomegaWithT) *bTree {
		bTree, err := New(2)
		g.Expect(err).ToNot(HaveOccurred())

		for i := 0; i < 1_024; i++ {
			_, _ = bTree.CreateOrFind(datatypes.Int(i), OnFindTest, NewBTreeTester(fmt.Sprintf("%d", i)))
		}

		return bTree
	}

	t.Run("returns nil if the item does not exist in the tree", func(t *testing.T) {
		bTree, err := New(2)
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(bTree.Find(Key1, OnFindTest)).To(BeNil())
	})

	t.Run("returns the item in the tree", func(t *testing.T) {
		bTree := setupTree(g)

		treeItem := bTree.Find(datatypes.Int(768), OnFindTest)
		g.Expect(treeItem).ToNot(BeNil())
		g.Expect(treeItem.(*BTreeTester).Value).To(Equal("768"))
		g.Expect(treeItem.(*BTreeTester).OnFindCount).To(Equal(1))
	})
}
