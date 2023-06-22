package btreeshared

//import (
//	"github.com/DanLavine/willow/internal/datastructures"
//	"github.com/DanLavine/willow/pkg/models/datatypes"
//)
//
//// bTreeShared is  way of grouping arbitrary key values into a unique searchable data set.
////
//// The tree is structured with these rules:
//// 1. The root of the tree is all the 'keys' which are searchable via a string.
//// 2. each KeyNode for a 'key' is another bTree with possible 'values'.
//// 3. each ValueNode is a struct contains a slice of [][]uint64, where the 1sst index is how many keys comprise a unique index.
//// I.E:
////
////	0 -> 1, so will only have 1 id ever.
////	1 -> 2, so will need to look for another tag (intersecction) to know if the pair share an ID. OR a union to find all IDs that have a particular KeyValue pair
////	2 -> 3, etc
////	...
////
//// Example (tree root):
////
////	       d
////	    /      \
////	   c       f,h
////	 /  \    /  |  \
////	a    b  e   g   i
////
//// If we were to investigate the tree of 'a' it could just be a list of all words that begin with 'a'.
////
////			apple
////	   /    \
////	ant     axe
////
//// at this point, any value will have a structure of:
////
////	type unique_ids struct {
////	  ids: [][]uint64
////	}
////
//// So if we wanted to just find the map[string]EncapsulatedData{'a':'ant'}
//// This would correspond to ids[0][0] -> jeneral ant info it the kye value pair (or whatever is saved)
////
//// but if we wanted something like large ant colonies, we could find something like map[string]EncapsulatedData{'a':'ant', 'colony size':'large'}
//// with these, we could do an intersection of ant.ids[1] n 'colony size'.ids[1] -> would output all intersected ids for large an colonies
//// if that is how we decided to store the data.
////
//// With this flexibility, we can find any type of unique groupings, and query a generalized key value data set. (TODO coming eventually)
//type bTreeShared interface {
//	Find(keyValuePairs datatypes.StringMap, onFind datastructures.OnFind) error
//
//	CreateOrFind(keyValuePairs datatypes.StringMap, onCreate datastructures.OnCreate, onFind datastructures.OnFind) error
//
//	Iterate(callback datastructures.OnFind) error
//
//	Delete(keyValuePairs datatypes.StringMap, canDelete datastructures.CanDelete) error
//}
