package btreeassociated

import (
	"fmt"

	"github.com/DanLavine/willow/internal/datastructures/set"
	"github.com/DanLavine/willow/pkg/models/datatypes"
)

// MatchPermutations can be used to find any combination of key values.
// I.E. suppose we we have keyValues = {"one": 1, 2: "something else", "float num": 3.42}
// to find all combinations, we can think of would be [{"one":1},{2:"something else"},{"float num": 3.42},{"one":1, 2: "something else"}, {"one":1, "float num": 3.42} ...]
//
// In the case of a large collection of KeyValues, a similar query would be massive to join all the possible combinations together.
// This is an optimization of performing those workflows in a reasonable way
//
// PARAMS:
// - keyValues - the possible key values to join together when searching for items in a tree
// - onQueryPagination - is the callback used for an items found in the tree. It will recive the objects' value saved in the tree (what were originally provided)
//
// RETURNS:
// - error - any errors encountered with the parameters
func (tsat *threadsafeAssociatedTree) MatchPermutations(keyValues KeyValues, onQueryPagination OnQueryPagination) error {
	if len(keyValues) == 0 {
		return fmt.Errorf("keyValues cannot be empty")
	}
	if onQueryPagination == nil {
		return fmt.Errorf("onQueryPagination cannot be nil")
	}

	// 1. first find all the nodes that we have a match for in the KeyValues
	idNodes := []*threadsafeIDNode{}
	for _, key := range keyValues.SortedKeys() {
		//for key, value := range keyValues {
		findValue := func(item any) {
			idNodes = append(idNodes, item.(*threadsafeIDNode))
		}

		findKey := func(item any) {
			valuesNode := item.(*threadsafeValuesNode)
			valuesNode.values.Find(keyValues[key], findValue)
		}

		if err := tsat.keys.Find(key, findKey); err != nil {
			return err
		}
	}

	// 2. For all the ID nodes, we need to group them together and find the values. Then run the callback on each found value.
	return tsat.generateIDPairs(idNodes, onQueryPagination)
}

func (tsat *threadsafeAssociatedTree) generateIDPairs(group []*threadsafeIDNode, onQueryPagination OnQueryPagination) error {
	switch len(group) {
	case 0:
		// nothing to do here
	case 1:
		// there is only 1 key value pair
		idNode := group[0]
		idNode.lock.RLock()
		defer idNode.lock.RUnlock()

		shouldContinue := false
		wrappedPagination := func(item any) {
			associatedKeyValues := item.(*AssociatedKeyValues)
			shouldContinue = onQueryPagination(associatedKeyValues)
		}

		// this should only ever loop 1 time atm. but just be safe and put in a loop for now
		if len(idNode.ids)-1 >= 0 {
			for _, id := range idNode.ids[0] {
				if err := tsat.associatedIDs.Find(datatypes.String(id), wrappedPagination); err != nil {
					return err
				}

				if !shouldContinue {
					return nil
				}
			}
		}
	default:
		// run the first index query
		idNode := group[0]
		idNode.lock.RLock()

		shouldContinue := false
		wrappedPagination := func(item any) {
			associatedKeyValues := item.(*AssociatedKeyValues)
			shouldContinue = onQueryPagination(associatedKeyValues)
		}

		if len(idNode.ids)-1 >= 0 {
			for _, id := range idNode.ids[0] { // this should only ever loop 1 time atm. but just be safe and put in a loop for now
				if err := tsat.associatedIDs.Find(datatypes.String(id), wrappedPagination); err != nil {
					idNode.lock.RUnlock()
					return err
				}

				if !shouldContinue {
					idNode.lock.RUnlock()
					return nil
				}
			}
		}
		idNode.lock.RUnlock()

		// drop a key and advance to the next subset
		if err := tsat.generateIDPairs(group[1:], onQueryPagination); err != nil {
			return err
		}

		// generate the N pair combinations
		for i := 1; i < len(group); i++ {
			err, shouldContinue := tsat.generateIDGroups(append([]*threadsafeIDNode{group[0]}, group[i]), group[i+1:], onQueryPagination)
			if err != nil {
				return err
			}

			if !shouldContinue {
				return nil
			}
		}
	}

	return nil
}

func (tsat *threadsafeAssociatedTree) generateIDGroups(prefix, suffix []*threadsafeIDNode, onQueryPagination OnQueryPagination) (error, bool) {
	// run the prefix combination
	idSet := set.New[string]()
	for index, node := range prefix {
		node.lock.RLock()

		if index == 0 {
			if len(node.ids)-1 >= len(prefix)-1 {
				idSet.AddBulk(node.ids[len(prefix)-1])
			} else {
				idSet.Clear()
			}
		} else {
			// this ensure that the idNode is safe to index, when trying to find large groupings. If the index only
			// contains small groupings, then we know there are no matching keys.
			//
			// I.E
			//  CREATE:
			//    {"one":"one"}
			//    {"two":"two","three","three","four":"four"}
			//
			//  SEARCH:
			//    {"one":"one", "two", "two"}
			//
			// both thos idNodes for "one" and "two" exist, but this is no grouping between them. Still the "one" node has a max node.ids[0] size and needs to be respected
			if len(node.ids)-1 >= len(prefix)-1 {
				idSet.Intersection(node.ids[len(prefix)-1])
			} else {
				idSet.Clear()
			}
		}

		// unlock the prefix idNodes now that we are done with it
		node.lock.RUnlock()
	}

	// loop through any found ids between all the prefix idNodes
	shouldContinue := false
	wrappedPagination := func(item any) {
		associatedKeyValues := item.(*AssociatedKeyValues)
		shouldContinue = onQueryPagination(associatedKeyValues)
	}

	for _, id := range idSet.Values() {
		if err := tsat.associatedIDs.Find(datatypes.String(id), wrappedPagination); err != nil {
			return err, false
		}

		if !shouldContinue {
			return nil, false
		}
	}

	// generate the N pair combinations
	for i := 0; i < len(suffix); i++ {
		tsat.generateIDGroups(append(prefix, suffix[i]), suffix[i+1:], onQueryPagination)
	}

	return nil, true
}
