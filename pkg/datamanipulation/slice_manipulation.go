package datamanipulation

import (
	"fmt"
	"sort"
)

//	PARAMETERS:
//	- strings - slice that will make all permutations from
//	- minKeys - minimum number of elements that make up a permutation
//	- maxKeys - maximum number of elements that make up a permutation. Use -1 to mean unlimited
//
//	RETURNS:
//	- [][]string - slice of permutations in sorted order by: 1. number of keys. 2. keys are sorted against other combinations
//	- error - if an errors were encountered with the parameters
//
// GenerateStringPermutations can be used to construct all possible permutations for a slice of strings.
//
// NOTE: This can be a dangerous operation with even a small initial slice. So when calling this function,
// ensure you have done the due diligence beforehand to make the strings to be the minimum needed set. The number
// of combinations is equal to list [1, 3, 7, 15, 31, ((n-1)*2 + 1)...]
//
// Example:
// - Strings{"a", "b", "c"} -> [{"a"}, {"b"}, {"c"}, {"a", "b"}, {"a", "c"}, {"b", "c"}, {"a", "b", "c"}]
func GenerateStringPermutations(strings []string, minKeys, maxKeys int) ([][]string, error) {
	if minKeys > maxKeys && maxKeys != -1 {
		return nil, fmt.Errorf("minKeys is greater than maxKeys")
	}

	// generate all the possible keys possible
	keyPermutations := generateStringPairs(strings, minKeys, maxKeys)

	// sort the keys
	// 1. according to length
	// 2. according to the keys in the slice
	sort.SliceStable(keyPermutations, func(i, j int) bool {
		if len(keyPermutations[i]) < len(keyPermutations[j]) {
			return true
		}

		if len(keyPermutations[i]) == len(keyPermutations[j]) {
			sort.Strings(keyPermutations[i])
			sort.Strings(keyPermutations[j])

			for index, value := range keyPermutations[i] {
				if value < keyPermutations[j][index] {
					return true
				} else if keyPermutations[j][index] < value {
					return false
				}
			}

			// at this point, all values must be in the proper order
			return true
		}

		return false
	})

	return keyPermutations, nil
}

func generateStringPairs(keys []string, minKeys, maxKeys int) [][]string {
	keyPairs := [][]string{}

	switch len(keys) {
	case 0:
		// nothing to do here
	case 1:
		// there is only 1 key to add
		if 1 < minKeys || (1 > maxKeys && maxKeys != -1) {
			return keyPairs
		}

		keyPairs = append(keyPairs, keys)
	default:
		// add the first index each time. Will recurse through original group shrinking by 1 each time to capture all elements
		if minKeys <= 1 && (maxKeys >= 1 || maxKeys == -1) {
			keyPairs = append(keyPairs, []string{keys[0]})
		}

		// drop a key and advance to the next subset ["a", "b", "c"] -> ["b", "c"]
		keyPairs = append(keyPairs, generateStringPairs(keys[1:], minKeys, maxKeys)...)

		for i := 1; i < len(keys); i++ {
			// generate all n[0,1] + n[x,x+1] groupings. I.E [{"a":"1","b":"2"}, {"a":"1","c":"3"}, ....]
			newKeysGrouping := []string{keys[0], keys[i]}
			keyPairs = append(keyPairs, generateStringGroups(newKeysGrouping, keys[i+1:], minKeys, maxKeys)...)
		}
	}

	return keyPairs
}

func generateStringGroups(prefix, suffix []string, minKeys, maxKeys int) [][]string {
	prefixLen := len(prefix)

	// This is super important to copy the slice. append on line 285, can add the values into prefix
	// as part of the append operation. So we want to setup an exact slice so a new value is allocated each time
	prefixCopy := make([]string, prefixLen, prefixLen)
	copy(prefixCopy, prefix)

	allKeyPairs := [][]string{}
	if prefixLen < minKeys {
		// do nothing, but need to continue growing the possible combinations
	} else if prefixLen >= minKeys && (prefixLen <= maxKeys || maxKeys == -1) {
		// add the combination and continue
		allKeyPairs = append(allKeyPairs, prefixCopy)
	} else if prefixLen > maxKeys && maxKeys != -1 {
		// can break since we are over the max size
		return allKeyPairs
	}

	// recurse building up to n size
	for i := 0; i < len(suffix); i++ {
		allKeyPairs = append(allKeyPairs, generateStringGroups(append(prefixCopy, suffix[i]), suffix[i+1:], minKeys, maxKeys)...)
	}

	return allKeyPairs
}
