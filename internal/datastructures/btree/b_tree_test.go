package btree

import (
	"fmt"
	"math"
	"testing"

	"github.com/DanLavine/willow/pkg/models/datatypes"
	. "github.com/onsi/gomega"
)

func validateTree(g *GomegaWithT, bNode *bNode, parentKey datatypes.CompareType, less bool) {
	if bNode == nil {
		return
	}

	var index int
	for index = 0; index < bNode.numberOfValues-1; index++ {
		// check parent key
		if parentKey != nil {
			if less {
				g.Expect(bNode.values[index].key.Less(parentKey)).To(BeTrue())
			} else {
				g.Expect(bNode.values[index].key.Less(parentKey)).To(BeFalse())
			}
		}

		// check current vllue is less than the next index
		g.Expect(bNode.values[index].key.Less(bNode.values[index+1].key)).To(BeTrue())

		if bNode.numberOfChildren != 0 {
			validateTree(g, bNode.children[index], bNode.values[index].key, true)
		}
	}

	// if there are any children, we need to check the last 2 indexes
	if bNode.numberOfChildren != 0 {
		validateTree(g, bNode.children[index], bNode.values[index].key, true)
		validateTree(g, bNode.children[index+1], bNode.values[index].key, false)
	}
}

func TestBTree_New(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("returns an error if the nodeSize is to small", func(t *testing.T) {
		bTree, err := New(1)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(Equal("nodeSize must be greater than 1 for BTree"))
		g.Expect(bTree).To(BeNil())
	})

	t.Run("returns an error if the nodeSize is to large", func(t *testing.T) {
		bTree, err := New(math.MaxInt)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(Equal(fmt.Sprintf("nodeSize must be 2 less than %d", math.MaxInt)))
		g.Expect(bTree).To(BeNil())
	})
}
