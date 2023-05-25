package idtree

import (
	"github.com/DanLavine/willow/internal/datastructures"
	"github.com/DanLavine/willow/pkg/models/datatypes"
)

func (idt *IDTree) Iterate(callback datastructures.Iterate) {
	idt.root.iterate(callback)
}

func (n *node) iterate(callback datastructures.Iterate) {
	// no value here
	if n.value != nil {
		callback(datatypes.Uint64(n.id), n.value)
	}

	if n.left != nil {
		n.left.iterate(callback)
	}

	if n.right != nil {
		n.right.iterate(callback)
	}
}
