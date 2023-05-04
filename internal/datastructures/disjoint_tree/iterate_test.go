package disjointtree

import (
	"testing"

	. "github.com/DanLavine/willow/internal/datastructures/disjoint_tree/testhelpers"
	"github.com/DanLavine/willow/pkg/models/datatypes"
	. "github.com/onsi/gomega"
)

func TestDisjointTree_Iterate(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("it runs the callback only on nodes with values", func(t *testing.T) {
		disjointTree := New()
		disjointTree.CreateOrFind(datatypes.Ints{1, 2}, nil, NewBTreeTester("first"))

		count := 0
		iterator := func(key datatypes.CompareType, value any) {
			treeTester := value.(*BTreeTester)
			g.Expect(treeTester.Value).To(Equal("first"))
			count++
		}

		disjointTree.Iterate(iterator)
		g.Expect(count).To(Equal(1))
	})
}
