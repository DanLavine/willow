package disjointtree

import "github.com/DanLavine/willow/internal/datastructures"

func (dt *disjointTree) Iterate(callback datastructures.Iterate) {
	if callback == nil {
		panic("callback is nil")
	}

	iterator := func(value any) {
		node := value.(*disjointNode)
		if node.value != nil {
			callback(node.value)
		}

		if node.children != nil {
			node.children.Iterate(callback)
		}
	}

	dt.tree.Iterate(iterator)
}
