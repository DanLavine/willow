package datastructures

import (
	"fmt"

	. "github.com/onsi/gomega"
)

var (
	key0   = IntTreeKey(0)
	key1   = IntTreeKey(1)
	key2   = IntTreeKey(2)
	key3   = IntTreeKey(3)
	key4   = IntTreeKey(4)
	key5   = IntTreeKey(5)
	key6   = IntTreeKey(6)
	key7   = IntTreeKey(7)
	key8   = IntTreeKey(8)
	key9   = IntTreeKey(9)
	key10  = IntTreeKey(10)
	key20  = IntTreeKey(20)
	key30  = IntTreeKey(30)
	key35  = IntTreeKey(35)
	key38  = IntTreeKey(38)
	key40  = IntTreeKey(40)
	key50  = IntTreeKey(50)
	key60  = IntTreeKey(60)
	key70  = IntTreeKey(70)
	key75  = IntTreeKey(75)
	key78  = IntTreeKey(78)
	key80  = IntTreeKey(80)
	key90  = IntTreeKey(90)
	key100 = IntTreeKey(100)
	key110 = IntTreeKey(110)
	key120 = IntTreeKey(120)
	key130 = IntTreeKey(130)
)

type bTreeTester struct {
	onFindCount int
	value       string
}

func newBTreeTester(value string) func() (any, error) {
	return func() (any, error) {
		return &bTreeTester{
			onFindCount: 0,
			value:       value,
		}, nil
	}
}
func newBTreeTesterWithError() (any, error) {
	return nil, fmt.Errorf("failure")
}

func (btt *bTreeTester) OnFind() {
	btt.onFindCount++
}

func validateTree(g *GomegaWithT, bNode *bNode, parentKey TreeKey, less bool) {
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
