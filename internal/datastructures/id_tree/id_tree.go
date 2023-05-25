package idtree

type direction int

const (
	left  direction = 0
	right direction = 1
)

// TODO: make sure this works for both big + little Endian
// TODO: add a lock

// ID Generator is used to generate a binary tree with a consistent placement of ints.
// The main features of this binary tree are:
//	1. Nodes are automatically removed if they are a leaf on delete
//  2. Nodes keep track of a missing left or right leaf node after a removal for the entire height of the tree.
//  3. Inserts replace missing leaf nodes so we don't need to worry about an index reaching the `math.MaxInt`
//
/* Example tree
height |						   ID
h0 - 1 (first index in this row) |                 1
                                 |         /               \
h1 - 2 (first index in this row) |        2                3
                                 |     /      \         /      \
h2 - 4 (first index in this row) |    4        6       5        7
                                 |  /   \    /   \   /   \    /   \
h3 - 8 (first index in this row) |  8   12  10	 14  9   13  11   15
*/
//  A left Node can always be calculated as: ((1 << parent.height) + parent.id)
//  A right Node can always be calculated as: (((1 << parent.height)*2) + parent.id)
//
//  The hight of any node can also becalculated like so:
/*
**  height := 0
**  val := node.id
**  for {
**    if val == 1 {
**      break
**    }
**    val = (val >> 1)
**    height++
**  }
 */

type IDTree struct {
	root *node
}

type node struct {
	id    uint64
	value any

	height int

	minLeft  uint64
	minRight uint64

	parent *node
	left   *node
	right  *node
}

func NewIDTree() *IDTree {
	return &IDTree{
		root: &node{
			height:   0,
			id:       1,
			minLeft:  2,
			minRight: 3,
		},
	}
}

// determine the ID of the node being added
func id(parent *node, direction direction) uint64 {
	if direction == left {
		return (1 << parent.height) + parent.id
	} else {
		return ((1 << parent.height) * 2) + parent.id
	}
}

// determine if the nextIndex should be placed on the left or the right
func (n *node) assignDirection() direction {
	if n.minLeft < n.minRight {
		return left
	}

	return right
}

// get the min from left and right values
func (n *node) min() uint64 {
	if n.minLeft < n.minRight {
		return n.minLeft
	}

	return n.minRight
}

// find a node at a particular index
func (n *node) findDirection(index uint64) direction {
	if (index/(1<<n.height))%2 == 0 {
		return left
	}

	return right
}

func newIndexNode(val any, parent *node, direction direction) (*node, uint64) {
	id := id(parent, direction)

	minLeft := (1 << (parent.height + 1)) + id
	minRight := ((1 << (parent.height + 1)) * 2) + id

	node := &node{
		id:     id,
		value:  val,
		height: parent.height + 1,

		minLeft:  minLeft,
		minRight: minRight,

		parent: parent,
	}

	return node, minLeft
}

// add a value and get the ID for where the item was inserted
func (idt *IDTree) Add(val any) uint64 {
	if idt.root == nil {
		idt.root = &node{id: 1, value: val, height: 0, minLeft: 2, minRight: 3, parent: nil}
		return 1
	} else {
		index, _ := idt.root.add(val, idt.root)
		return index
	}
}

// add an new index, or replace a missing index
//
// RETURNS:
// * uint64 - the ID of the index value was placed int
// * uint64 - the least index either left or right
func (n *node) add(value any, parent *node) (uint64, uint64) {
	var nodeID, min uint64

	if n.value == nil {
		// found an empty index, add the new value
		n.value = value
		nodeID = n.id

		return n.id, n.min()
	} else {
		// create or recurse into node
		if n.assignDirection() == left {
			if n.left == nil {
				// assign new index left
				n.left, min = newIndexNode(value, n, left)
				nodeID = n.left.id
			} else {
				// recurse into left index
				nodeID, min = n.left.add(value, n)
			}

			if n.minLeft < min {
				n.minLeft = min
			}
		} else {
			if n.right == nil {
				// assign new index right
				n.right, min = newIndexNode(value, n, right)
				nodeID = n.right.id
			} else {
				// recurse into right index
				nodeID, min = n.right.add(value, n)
			}

			if n.minRight < min {
				n.minRight = min
			}
		}
	}

	return nodeID, n.min()
}

// Get an item at a specific index
func (idt *IDTree) Get(index uint64) any {
	node := idt.root

	if node == nil {
		return nil
	}

	// Don't worry about cleanup here. Not possible to hit, Remove() takes care of that
	value, _, _ := node.find(index, false)
	return value
}

// remove an item at the desired index
func (idt *IDTree) Remove(index uint64) any {
	node := idt.root

	if node == nil {
		return nil
	}

	value, min, cleanup := node.find(index, true)
	if cleanup {
		if node.findDirection(index) == left {
			node.left = nil
			node.minLeft = min
		} else {
			node.right = nil
			node.minRight = min
		}

		if node.left == nil && node.right == nil {
			idt.root = nil
		}
	}

	return value
}

// RETURNS
// * any - the value for the particular index
// * uint64 - id of the removed index. 0 means no action to take
// * bool - if the node should be set to nil in the parent
func (n *node) find(index uint64, remove bool) (any, uint64, bool) {
	// found current index, return value and break recursion
	if n.id == index {
		value := n.value

		// clear out the node and figure out if it should be deleted
		if remove {
			n.value = nil
			return value, n.id, n.left == nil && n.right == nil
		}

		return value, n.id, false
	}

	if n.findDirection(index) == left {
		// the index is on the left

		// left is already nil. This node might need to be cleaned up as well
		if n.left == nil {
			return nil, n.id, n.value == nil && n.right == nil
		}

		value, minLeft, deleteChild := n.left.find(index, remove)

		// check to see if all children are nil
		if deleteChild {
			n.left = nil

			// this can be removed as well since all children are nil
			if n.value == nil && n.right == nil {
				return value, n.id, true
			}
		}

		if remove {
			if n.minLeft > minLeft {
				n.minLeft = minLeft
			}
		}

		return value, n.minLeft, false
	} else {
		// the index is on the right

		// right is already nil. This node might need to be cleaned up as well
		if n.right == nil {
			return nil, n.id, n.value == nil && n.right == nil
		}

		value, minRight, deleteChild := n.right.find(index, remove)

		// check to see if all children are nil
		if deleteChild {
			n.right = nil

			// this can be removed as well since all children are nil
			if n.value == nil && n.left == nil {
				return value, n.id, true
			}
		}

		if remove {
			if n.minRight > minRight {
				n.minRight = minRight
			}
		}

		return value, n.minRight, false
	}
}
