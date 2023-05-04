package disjointtree

import (
	"fmt"

	"github.com/DanLavine/willow/internal/datastructures"
	"github.com/DanLavine/willow/pkg/models/datatypes"
)

// Find returns a value iff it exists within the disjoint tree. The Keys passed to Find
// must all exist in the disjoint tree in the order they are passed in.
//
// PARAMS:
// * keys - the set of keys for an item in the tree
// * onFind - (optional) callback to run when the item is found in the tree
//
// RETURNS
// * any - item found in the tree. If the item is not found this will be nil
// * error - any errors related to keys provided (such as nil, or an empty set)
func (dt *disjointTree) Find(keys datatypes.EnumerableCompareType, onFind datastructures.OnFind) (any, error) {
	if keys == nil || keys.Len() <= 0 {
		return nil, fmt.Errorf("EnumberableTreeKeys must have at least 1 element")
	}

	return dt.find(keys, onFind)
}

func (dt *disjointTree) find(keys datatypes.EnumerableCompareType, onFind datastructures.OnFind) (any, error) {
	key, keys := keys.Pop()

	// find the tree item
	treeItem := dt.tree.Find(key, dtReadLock)
	if treeItem == nil {
		return nil, nil
	}
	disjointNode := treeItem.(*disjointNode)
	defer disjointNode.lock.RUnlock()

	// we are at the final index
	if keys.Len() == 0 {
		if onFind != nil {
			onFind(disjointNode.value)
		}

		return disjointNode.value, nil
	}

	// recurse
	if disjointNode.children == nil {
		return nil, nil
	}

	return disjointNode.children.find(keys, onFind)
}

// SearchKeys returns all values in the disjoint tree whos values match. The Keys passed to Search
// Will check values recursively to find any potential matches.
//
// Consider adding a number of values to a disjoint tree like so:
//   - ["namespace","123","org","abc","project","testA"]
//   - ["namespace","123","org","abc","project","testB"]
//   - ["namespace","123","project","testF"]
//   - ["org","abc","project","testA"]
//   - ["project","testA"]
//
// We would end up with a disjointed tree like:
//
//	  (top level node)        (children under namespace)      (chidlren under "123")     (children under "org")  ...
//		      org                    "123"                           org, project                 "abc"
//		 /         \
//	namespace   project
//
// If we now wanted to find all values with that have a key "org" the we would need to check both the Org Tree + the Namespace Tree
// but we wouldn't need to search the project tree. This is because on insertion, we know the Keys(0,2,4..., even indexes) to be in
// a sorted order. This means we know that the "Project" Tree will not have any keys less than it. Using this knowledge we can
// recursively search through the tree's to find all possible values.
//
// PARAMS:
// * keys - the set of keys for an item in the tree
// * onFind - (optional) callback to run when the item is found in the tree
//
// RETURNS
// * any - item found in the tree. If the item is not found this will be nil
// * error - any errors related to keys provided (such as nil, or an empty set)
//func (dt *disjointTree) SearchKeys(keys datatypes.EnumerableCompareType, onFind datastructures.OnFind) ([]SearchResult, error) {
//	if keys == nil || keys.Len() == 0 {
//		return nil, fmt.Errorf("Cannot have empty search keys")
//	}
//
//	return dt.searchKeys(nil, keys, onFind), nil
//}
//
//func (dt *disjointTree) searchKeys(tags datatypes.EnumerableCompareType, keys datatypes.EnumerableCompareType, onFind datastructures.OnFind) []SearchResult {
//	searchResults := SearchResults{}
//	searchLock := new(sync.Mutex)
//
//	key, keys := keys.Pop()
//	item := dt.tree.Find(key, onFind)
//	node := item.(*disjointNode)
//
//	tags = append(tags, key)
//
//	if keys.Len() == 0 {
//		// we are at the final values
//		node.children.Iterate(func(childKey datatypes.CompareType, childItem any) {
//			// only add the disjoint node's that have a proper value
//			if node.value != nil {
//				searchResults = append(searchResults, SearchResult{Tags: append(tags, childKey.(datatypes.String)), Value: node.value})
//			}
//		})
//	} else {
//		// need to recurse to all the nodes
//
//		// if there are children on this node, we keep checking the children for the rest of the keys
//		if node.children != nil {
//			wg := new(sync.WaitGroup)
//
//			// for each value in the child tree
//			node.children.Iterate(func(childKey datatypes.CompareType, childItem any) {
//				childNode := childItem.(*disjointNode)
//
//				// check their values for the next key
//				if childNode.children != nil {
//					wg.Add(1)
//					go func(recuseTags datatypes.Strings, recurseNode *disjointNode) {
//						defer wg.Done()
//						childResults := recurseNode.children.searchKeys(recuseTags, keys, onFind)
//
//						// add all child values
//						searchLock.Lock()
//						searchResults = append(searchResults, childResults...)
//						searchLock.Unlock()
//					}(append(tags, childKey.(datatypes.String)), childNode)
//				}
//			})
//
//			// wait for all children to stop processing
//			wg.Wait()
//		}
//	}
//
//	return searchResults
//}
//
//// SearchKeyValues returns all values in the disjoint tree whos values match. The Keys passed to Search
//// Will check values recursively to find any potential matches.
////
//// Consider adding a number of values to a disjoint tree like so:
////   - ["namespace","123","org","abc","project","testA"]
////   - ["namespace","123","org","abc","project","testB"]
////   - ["namespace","123","project","testF"]
////   - ["org","abc","project","testA"]
////   - ["project","testA"]
////
//// We would end up with a disjointed tree like:
////
////	  (top level node)        (children under namespace)      (chidlren under "123")     (children under "org")  ...
////		      org                    "123"                           org, project                 "abc"
////		 /         \
////	namespace   project
////
//// If we now wanted to find all values with "org" = "abc" the we would need to check both the Org Tree + the Namespace Tree
//// but we wouldn't need to search the project tree. This is because on insertion, we know the Keys(0,2,4..., even indexes) to be in
//// a sorted order. This means we know that the "Project" TRee will not have any keys less than it. Using this knowledge we can
//// recursively search through the tree's to find all possible values.
////
//// PARAMS:
//// * keys - the set of keys for an item in the tree
//// * onFind - (optional) callback to run when the item is found in the tree
////
//// RETURNS
//// * any - item found in the tree. If the item is not found this will be nil
//// * error - any errors related to keys provided (such as nil, or an empty set)
//func (dt *disjointTree) SearchKeyValuess(keyValues datatypes.EnumerableCompareType, onFind datastructures.OnFind) ([]any, error) {
//	return nil, nil
//}
