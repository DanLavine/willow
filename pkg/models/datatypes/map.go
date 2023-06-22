package datatypes

import "sort"

type StringMap map[string]EncapsulatedData

func (sm StringMap) Keys() []string {
	keys := []string{}

	for key, _ := range sm {
		keys = append(keys, key)
	}

	return keys
}

func (sm StringMap) SoretedKeys() []string {
	keys := []string{}

	for key, _ := range sm {
		keys = append(keys, key)
	}

	sort.Strings(keys)
	return keys
}

// GenerateGroupPairs can be used to go through a list of strings and create all unique ordered groupings.
// The returned slice is sorted on length of key value pairs with the longest value always last
//
// Example:
// *	Tags{"b":"2", "a":"1", "c":"3"} -> [{"a":"1"}, {"b":"2"}, {"c":"3"}, {"a":"1", "b":"2"}, {"a":"1", "c":"3"}, {"b":"2", "c":"3"}, {"a":"1", "b":"2", "c":"3"}]
func (sm StringMap) GenerateTagPairs() []StringMap {
	groupPairs := sm.generateTagPairs(sm.Keys())

	sort.Slice(groupPairs, func(i, j int) bool {
		return len(groupPairs[i]) < len(groupPairs[j])
	})

	return groupPairs
}

func (sm StringMap) generateTagPairs(group []string) []StringMap {
	var allGroupPairs []StringMap

	switch len(group) {
	case 0:
		// nothing to do here
	case 1:
		// there is only 1 key value pair
		allGroupPairs = append(allGroupPairs, StringMap{group[0]: sm[group[0]]})
	default:
		// add the first index each time. Will recurse through original group shrinking by 1 each time to capture all elements
		allGroupPairs = append(allGroupPairs, StringMap{group[0]: sm[group[0]]})

		// drop a key and advance to the next subset ["a", "b", "c"] -> ["b", "c"]
		allGroupPairs = append(allGroupPairs, sm.generateTagPairs(group[1:])...)

		for i := 1; i < len(group); i++ {
			// generate all n[0,1] + n[x,x+1] groupings. I.E [{"a":"1","b":"2"}, {"a":"1","c":"3"}, ....]
			newGrouping := []string{group[0], group[i]}
			allGroupPairs = append(allGroupPairs, sm.generateTagGroups(newGrouping, group[i+1:])...)
		}
	}

	return allGroupPairs
}

func (sm StringMap) generateTagGroups(prefix, suffix []string) []StringMap {
	allGroupPairs := []StringMap{}

	// add initial combined slice
	baseGrouping := StringMap{}
	for _, prefixKey := range prefix {
		baseGrouping[prefixKey] = sm[prefixKey]
	}
	allGroupPairs = append(allGroupPairs, baseGrouping)

	// recurse building up to n size
	for i := 0; i < len(suffix); i++ {
		allGroupPairs = append(allGroupPairs, sm.generateTagGroups(append(prefix, suffix[i]), suffix[i+1:])...)
	}

	return allGroupPairs
}
