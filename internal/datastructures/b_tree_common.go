package datastructures

import "fmt"

// check to see if the key is greate than the current value
func (v *value) greater(compareKey TreeKey) bool {
	return !v.key.Less(compareKey)
}

// lastChild returns the last child in the bTree
func (bn *bNode) lastChild() *bNode {
	switch bn.numberOfChildren {
	case 0:
		return nil
	default:
		return bn.children[bn.numberOfChildren-1]
	}
}

// lastValue returns the last value in the bTree
func (bn *bNode) lastValue() *value {
	switch bn.numberOfValues {
	case 0:
		return nil
	default:
		return bn.values[bn.numberOfValues-1]
	}
}

// dropGreates is used to remove the rightmost value and children from a node
func (bn *bNode) dropGreatest() {
	bn.values[bn.numberOfValues-1] = nil
	bn.numberOfValues--

	if bn.numberOfChildren != 0 {
		bn.children[bn.numberOfChildren-1] = nil
		bn.numberOfChildren--
	}
}

// leafs will never have any children so just use this as the check
func (bn *bNode) isLeaf() bool {
	return bn.numberOfChildren == 0
}

// non leaf. non root
func (bn *bNode) minChildren() int {
	// same as math.Ceil(order / 2). but ceil only works for floats
	return (cap(bn.values) + 1) / 2
}

func (bn *bNode) maxChildren() int {
	return cap(bn.values)
}

// true leaf. non root
func (bn *bNode) minValues() int {
	return bn.minChildren() - 1
}

func (bn *bNode) maxValues() int {
	return cap(bn.values) - 1
}

// complete helper function for tests
func (bn *bNode) print(parentString string) {
	if parentString == "" {
		fmt.Println("tree")
	}
	passedString := parentString

	var index int
	for index = 0; index < bn.numberOfValues; index++ {
		parentString = passedString
		parentString = fmt.Sprintf("%s[%d]", parentString, index)

		if bn.children[index] != nil {
			bn.children[index].print(fmt.Sprintf("%s.child[%d]", parentString, index))
		}

		if bn.values[index] != nil {
			if index == 0 {
				fmt.Printf("%s key: %v, number of values: %d, number of children %d\n", parentString, bn.values[index].key, bn.numberOfValues, bn.numberOfChildren)
			} else {
				fmt.Printf("%s key: %v\n", parentString, bn.values[index].key)
			}
		}
	}

	if bn.children[index] != nil {
		bn.children[index].print(fmt.Sprintf("%s.child[%d]", parentString, index))
	}
}
