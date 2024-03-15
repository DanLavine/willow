package btree

import (
	"fmt"
	"math"
	"testing"

	v1common "github.com/DanLavine/willow/pkg/models/api/common/v1"
	"github.com/DanLavine/willow/pkg/models/datatypes"
	. "github.com/onsi/gomega"
)

var noTypesRestriction = v1common.TypeRestrictions{MinDataType: datatypes.T_uint8, MaxDataType: datatypes.T_any}

func validateThreadSafeTree(g *GomegaWithT, bNode *threadSafeBNode) {
	// root must be nil
	if bNode == nil {
		return
	}

	var index int
	for index = 0; index < bNode.numberOfValues; index++ {
		// check current value is less than the next index
		if index < bNode.numberOfValues-1 {
			g.Expect(bNode.keyValues[index].key.LessMatchType(bNode.keyValues[index+1].key)).To(BeTrue())
		}

		// check all less than children
		if bNode.numberOfChildren != 0 {
			validateThreadSafeNode(g, bNode.children[index], bNode.keyValues[index].key, true)
		}
	}

	// if there are any children, we need to check greater than indexes
	if bNode.numberOfChildren != 0 {
		validateThreadSafeNode(g, bNode.children[index], bNode.keyValues[index-1].key, false)
	}
}

func validateThreadSafeNode(g *GomegaWithT, bNode *threadSafeBNode, parentKey datatypes.EncapsulatedValue, less bool) {
	if bNode == nil {
		return
	}

	var index int
	for index = 0; index < bNode.numberOfValues; index++ {
		if less {
			// ensure parent key is greater
			g.Expect(bNode.keyValues[index].key.LessMatchType(parentKey)).To(BeTrue())
		} else {
			// ensure parent key is less
			g.Expect(bNode.keyValues[index].key.LessMatchType(parentKey)).To(BeFalse())
		}

		// check current value is less than the next index
		if index < bNode.numberOfValues-1 {
			g.Expect(bNode.keyValues[index].key.LessMatchType(bNode.keyValues[index+1].key)).To(BeTrue())
		}

		// check all less than children
		if bNode.numberOfChildren != 0 {
			validateThreadSafeNode(g, bNode.children[index], bNode.keyValues[index].key, true)
		}
	}

	// if there are any children, we need to check the last 2 indexes
	if bNode.numberOfChildren != 0 {
		validateThreadSafeNode(g, bNode.children[index], bNode.keyValues[index-1].key, false)
	}
}

func TestBTree_NewThreadSafe(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("returns an error if the nodeSize is to small", func(t *testing.T) {
		bTree, err := NewThreadSafe(1)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(Equal("nodeSize must be greater than 1 for BTree"))
		g.Expect(bTree).To(BeNil())
	})

	t.Run("returns an error if the nodeSize is to large", func(t *testing.T) {
		bTree, err := NewThreadSafe(math.MaxInt)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(Equal(fmt.Sprintf("nodeSize must be 2 less than %d", math.MaxInt)))
		g.Expect(bTree).To(BeNil())
	})
}
