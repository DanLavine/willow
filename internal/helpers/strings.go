package helpers

import "github.com/DanLavine/willow/pkg/models/datatypes"

// GenerateGroupPairs can be used to go through a list of strings and create all unique ordered groupings
// Example:
// *	group [a,b,c,d,e] -> [[a], [b], [c], [d], [e], [a,b], [a,c], [a,d], [a,e], [a,b,c], [a,b,d], [a,b,e], [a,c,d], [a,c,e], [a,d,e], [a,b,c,d], [a,b,c,e], [a,c,d,e], [a,b,c,d,e], [b,c], ...]
//
// NOTE that as part of willow we assume all requests with tags to be sorted and so this is the reason we
// only care about the in order groupings.
func GenerateGroupPairs(group datatypes.Strings) []datatypes.Strings {
	var allGroupPairs []datatypes.Strings
	groupLen := len(group)

	switch groupLen {
	case 0:
		return allGroupPairs
	case 1:
		allGroupPairs = append(allGroupPairs, datatypes.Strings{group[0]})
	default:
		// add the first index each time. Will recurse through original group shrinking by [0] each time to capture all elements
		allGroupPairs = append(allGroupPairs, datatypes.Strings{group[0]})

		for i := 1; i < groupLen; i++ {
			// generate all n[0] + n[x] groupings. I.E [[a,b], [a,c], [a,d], [a,e]]
			newGrouping := datatypes.Strings{group[0], group[i]}
			allGroupPairs = append(allGroupPairs, generateGroupPairs(newGrouping, group[i+1:])...)
		}

		// advance to nex subset ["b", "c", "d", "e"]
		allGroupPairs = append(allGroupPairs, GenerateGroupPairs(group[1:])...)
	}

	return allGroupPairs
}

func generateGroupPairs(prefix, suffix datatypes.Strings) []datatypes.Strings {
	var allGroupPairs []datatypes.Strings
	groupLen := len(suffix)

	// add initial combined slice
	allGroupPairs = append(allGroupPairs, prefix)

	// recurse building up to n size
	for i := 0; i < groupLen; i++ {
		newGrouping := datatypes.Strings{}
		for _, element := range prefix {
			newGrouping = append(newGrouping, element)
		}

		newGrouping = append(newGrouping, suffix[i])
		allGroupPairs = append(allGroupPairs, generateGroupPairs(newGrouping, suffix[i+1:])...)
	}

	return allGroupPairs
}
