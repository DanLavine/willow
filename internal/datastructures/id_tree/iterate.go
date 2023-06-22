package idtree

import (
	"github.com/DanLavine/willow/internal/datastructures"
)

func (idt *IDTree) Iterate(callback datastructures.OnFind) {
	if idt.root != nil {
		idt.root.iterate(callback)
	}
}

func (n *node) iterate(callback datastructures.OnFind) {
	if n.value != nil {
		callback(n.value)
	}

	if n.left != nil {
		n.left.iterate(callback)
	}

	if n.right != nil {
		n.right.iterate(callback)
	}
}
