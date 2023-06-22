package btree

import (
	"fmt"
	"testing"

	"github.com/DanLavine/willow/pkg/models/datatypes"
	. "github.com/onsi/gomega"
)

func TestBTree_Find(t *testing.T) {
	g := NewGomegaWithT(t)

	setupTree := func(g *GomegaWithT) *threadSafeBTree {
		bTree, err := NewThreadSafe(2)
		g.Expect(err).ToNot(HaveOccurred())

		for i := 0; i < 1_024; i++ {
			g.Expect(bTree.CreateOrFind(datatypes.Int(i), NewBTreeTester(fmt.Sprintf("%d", i)), OnFindTest)).ToNot(HaveOccurred())
		}

		return bTree
	}

	t.Run("it returns an error if the key is invalid", func(t *testing.T) {
		bTree := setupTree(g)

		err := bTree.Find(datatypes.EncapsulatedData{DataType: -1, Value: "bad"}, OnFindTest)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("key is invalid:"))
	})

	t.Run("it returns an error if the onFind callback is nil", func(t *testing.T) {
		bTree, err := NewThreadSafe(2)
		g.Expect(err).ToNot(HaveOccurred())

		err = bTree.Find(Key1, nil)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("onFind is nil"))
	})

	t.Run("it does not run the callback if the item is not found in the tree", func(t *testing.T) {
		bTree := setupTree(g)

		called := false
		onFind := func(item any) {
			called = true
		}

		g.Expect(bTree.Find(datatypes.Int64(-1), onFind)).ToNot(HaveOccurred())
		g.Expect(called).To(BeFalse())
	})

	t.Run("returns the item in the tree", func(t *testing.T) {
		bTree := setupTree(g)

		called := false
		onFind := func(item any) {
			btt := item.(*BTreeTester)
			g.Expect(btt.Value).To(Equal("768"))
			called = true
		}

		bTree.Find(datatypes.Int(768), onFind)
		g.Expect(called).To(BeTrue())
	})
}
