package disjointtree

//import (
//	"github.com/DanLavine/willow/internal/datastructures"
//	"github.com/DanLavine/willow/pkg/models/datatypes"
//)
//
//func (dt *disjointTree) Iterate(callback datastructures.Iterate) {
//	if callback == nil {
//		panic("callback is nil")
//	}
//
//	iterator := func(key datatypes.CompareType, value any) {
//		node := value.(*disjointNode)
//		if node.value != nil {
//			callback(key, node.value)
//		}
//
//		if node.children != nil {
//			node.children.Iterate(callback)
//		}
//	}
//
//	dt.tree.Iterate(iterator)
//}
