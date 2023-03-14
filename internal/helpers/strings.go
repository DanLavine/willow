package helpers

import (
	"strings"
)

// GenerateStringPairs can be used to go through a list of strings and create all unique ordered groupings
// Example:
// *	group [a,b,c,d,e] -> [a,b,c,d,e,ab,ac,ad,ae,abc,abd,abe,acd,ace,ade,bc,bd,be,bcd,bce,cd,cde,de]
//
// NOTE that as part of willow we assume all requests with tags to be sorted and so this is the reason we
// only care about the in order groupings.
func GenerateStringPairs(group []string) []string {
	var allGroups []string
	groupLen := len(group)

	switch groupLen {
	case 0:
		return allGroups
	case 1:
		allGroups = append(allGroups, group[0])
	default:
		// add the first index each time. Will recurse through original group shrinking by [0] each time to capture all elements
		allGroups = append(allGroups, group[0])

		for i := 1; i < len(group); i++ {
			// generate all n[0] + n[x] groupings. I.E ["ab", "ac", "ad", "ae"]
			newGrouping := strings.Join(joinSlices([]string{group[0]}, []string{group[i]}), "")
			allGroups = append(allGroups, generateStringPairsGroups(joinSlices([]string{newGrouping}, group[i+1:]))...)
		}

		// advance to nex subset ["b", "c", "d", "e"]
		allGroups = append(allGroups, GenerateStringPairs(group[1:])...)
	}

	return allGroups
}

func generateStringPairsGroups(group []string) []string {
	var allGroups []string
	groupLen := len(group)

	// add initial combined key
	allGroups = append(allGroups, group[0])

	for i := 1; i < groupLen; i++ {
		newGrouping := strings.Join(joinSlices([]string{group[0]}, []string{group[i]}), "")
		allGroups = append(allGroups, generateStringPairsGroups(joinSlices([]string{newGrouping}, group[i+1:]))...)
	}

	return allGroups
}

// join slices into a unique slice. append(...) unfortuniitly changes the value of 1st element so not safe
// to use when you want to keep original slice value
func joinSlices(prefix, suffix []string) []string {
	newSlice := make([]string, len(prefix)+len(suffix))
	copy(newSlice, prefix)
	copy(newSlice[len(prefix):], suffix)

	return newSlice
}
