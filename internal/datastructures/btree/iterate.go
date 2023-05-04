package btree

import "github.com/DanLavine/willow/internal/datastructures"

// Iterate over each node with a thread safe read lock and call the iterate function when the value != nil
//
// PARAMS:
// * callback - function is called when a Tree's Node value != nil. The Iterate callback is passed the Node's value
func (btree *bTree) Iterate(callback datastructures.Iterate) {
	if callback == nil {
		panic("callback is nil")
	}

	btree.lock.RLock()
	defer btree.lock.RUnlock()

	if btree.root != nil {
		btree.root.iterate(callback)
	}
}

func (bn *bNode) iterate(callback datastructures.Iterate) {
	bn.lock.RLock()
	defer bn.lock.RUnlock()

	for i := 0; i < bn.numberOfValues; i++ {
		callback(bn.values[i].key, bn.values[i].item)
	}

	for i := 0; i < bn.numberOfChildren; i++ {
		bn.children[i].iterate(callback)
	}
}
