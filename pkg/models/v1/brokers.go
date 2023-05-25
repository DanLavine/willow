package v1

import (
	"sort"

	"github.com/DanLavine/willow/pkg/models/datatypes"
)

type BrokerType uint32

const (
	Queue BrokerType = iota
)

// Converet the broker type to a String
func (bt BrokerType) ToString() string {
	switch bt {
	case Queue:
		return "queue"
	default:
		return "unknown"
	}
}

type BrokerInfo struct {
	// specific queue name for the message
	Name datatypes.String

	// Type of broker
	// NOTE: not currently used
	//BrokerType BrokerType

	// possible tags used by the broker
	// leaving this empty (or nil) results to the default tag set
	Tags datatypes.StringMap
}

func (b BrokerInfo) validate() *Error {
	if b.Name == "" {
		return InvalidRequestBody.With("Name to be provided", "Name is the empty string")
	}

	return nil
}

// GenerateGroupPairs can be used to go through a list of strings and create all unique ordered groupings
// Example:
// *	Tags{"b":"2", "a":"1", "c":"3"} -> [{"a":"1"}, {"b":"2"}, {"c":"3"}, {"a":"1", "b":"2"}, {"a":"1", "c":"3"}, {"b":"2", "c":"3"}, {"a":"1", "b":"2", "c":"3"}]
func (b BrokerInfo) GenerateTagPairs() []datatypes.StringMap {
	groupPairs := b.generateTagPairs(b.SortedTags())

	sort.Slice(groupPairs, func(i, j int) bool {
		return len(groupPairs[i]) < len(groupPairs[j])
	})

	groupsAsMaps := []datatypes.StringMap{}
	for _, group := range groupPairs {
		newMap := datatypes.StringMap{}

		for index := 0; index < len(group); index += 2 {
			newMap[group[index]] = group[index+1]
		}

		groupsAsMaps = append(groupsAsMaps, newMap)
	}

	return groupsAsMaps
}

func (b BrokerInfo) generateTagPairs(group datatypes.Strings) []datatypes.Strings {
	var allGroupPairs []datatypes.Strings

	switch len(group) {
	case 0:
		// nothing to do here
	case 1:
		// set to the default tag
		allGroupPairs = append(allGroupPairs, group)
	case 2:
		// this is the last key value pair of original tags
		allGroupPairs = append(allGroupPairs, group)
	default:
		// add the first index each time. Will recurse through original group shrinking by [0.1] each time to capture all elements
		allGroupPairs = append(allGroupPairs, datatypes.Strings{group[0], group[1]})

		// advance to nex subset ["b", "c", "d", "e"]
		allGroupPairs = append(allGroupPairs, b.generateTagPairs(group[2:])...)

		for i := 2; i < len(group); i += 2 {
			// generate all n[0,1] + n[x,x+1] groupings. I.E [{"a":"1","b":"2"}, {"a":"1","c":"3"}, ....]
			newGrouping := datatypes.Strings{group[0], group[1], group[i], group[i+1]}
			allGroupPairs = append(allGroupPairs, b.generateTagGroups(newGrouping, group[i+2:])...)
		}
	}

	return allGroupPairs

}

func (b BrokerInfo) generateTagGroups(prefix, suffix datatypes.Strings) []datatypes.Strings {
	var allGroupPairs []datatypes.Strings
	groupLen := len(suffix)

	// add initial combined slice
	allGroupPairs = append(allGroupPairs, prefix)

	// recurse building up to n size
	for i := 0; i < groupLen; i += 2 {
		newGrouping := datatypes.Strings{}
		for _, element := range prefix {
			newGrouping = append(newGrouping, element)
		}

		newGrouping = append(newGrouping, suffix[i])
		newGrouping = append(newGrouping, suffix[i+1])
		allGroupPairs = append(allGroupPairs, b.generateTagGroups(newGrouping, suffix[i+2:])...)
	}

	return allGroupPairs
}

func (b BrokerInfo) SortedTags() datatypes.Strings {
	tagsLen := len(b.Tags)
	if tagsLen == 0 {
		return datatypes.Strings{"default"}
	}

	// 1st, sort all key values
	tagKeys := make(datatypes.Strings, 0, tagsLen)
	for key, _ := range b.Tags {
		tagKeys = append(tagKeys, key)
	}
	tagKeys.Sort()

	// 2nd, get the values for all the keys and place them properly
	tags := make(datatypes.Strings, tagsLen*2, tagsLen*2)
	for i := 0; i < tagsLen; i++ {
		tags[i*2] = tagKeys[i]             // shift the key to proper location
		tags[(i*2)+1] = b.Tags[tagKeys[i]] // grab the value for the key and place it in the proper location
	}

	return tags
}

//// GenerateGroupPairs can be used to go through a list of strings and create all unique ordered groupings
//// Example:
//// *	Tags{"b":"2", "a":"1", "c":"3"} -> [{"a":"1"}, {"b":"2"}, {"c":"3"}, {"a":"1", "b":"2"}, {"a":"1", "c":"3"}, {"b":"2", "c":"3"}, {"a":"1", "b":"2", "c":"3"}]
////
//// NOTE that as part of willow we assume all requests with tags to be sorted and so this is the reason we
//// only care about the in order groupings.
////
//// TODO: move this onto v1.Broker's Tag functionality
//func (b BrokerInfo) GenerateTagPairs() []datatypes.Strings {
//	groupPairs := b.generateTagPairs(b.SortedTags())
//
//	sort.Slice(groupPairs, func(i, j int) bool {
//		return len(groupPairs[i]) < len(groupPairs[j])
//	})
//
//	return groupPairs
//}
//
//func (b BrokerInfo) generateTagPairs(group datatypes.Strings) []datatypes.Strings {
//	var allGroupPairs []datatypes.Strings
//
//	switch len(group) {
//	case 0:
//		// nothing to do here
//	case 1:
//		// set to the default tag
//		allGroupPairs = append(allGroupPairs, group)
//	case 2:
//		// this is the last key value pair of original tags
//		allGroupPairs = append(allGroupPairs, group)
//	default:
//		// add the first index each time. Will recurse through original group shrinking by [0.1] each time to capture all elements
//		allGroupPairs = append(allGroupPairs, datatypes.Strings{group[0], group[1]})
//
//		// advance to nex subset ["b", "c", "d", "e"]
//		allGroupPairs = append(allGroupPairs, b.generateTagPairs(group[2:])...)
//
//		for i := 2; i < len(group); i += 2 {
//			// generate all n[0,1] + n[x,x+1] groupings. I.E [{"a":"1","b":"2"}, {"a":"1","c":"3"}, ....]
//			newGrouping := datatypes.Strings{group[0], group[1], group[i], group[i+1]}
//			allGroupPairs = append(allGroupPairs, b.generateTagGroups(newGrouping, group[i+2:])...)
//		}
//	}
//
//	return allGroupPairs
//
//}
//
//func (b BrokerInfo) generateTagGroups(prefix, suffix datatypes.Strings) []datatypes.Strings {
//	var allGroupPairs []datatypes.Strings
//	groupLen := len(suffix)
//
//	// add initial combined slice
//	allGroupPairs = append(allGroupPairs, prefix)
//
//	// recurse building up to n size
//	for i := 0; i < groupLen; i += 2 {
//		newGrouping := datatypes.Strings{}
//		for _, element := range prefix {
//			newGrouping = append(newGrouping, element)
//		}
//
//		newGrouping = append(newGrouping, suffix[i])
//		newGrouping = append(newGrouping, suffix[i+1])
//		allGroupPairs = append(allGroupPairs, b.generateTagGroups(newGrouping, suffix[i+2:])...)
//	}
//
//	return allGroupPairs
//}
//
//func (b BrokerInfo) SortedTags() datatypes.Strings {
//	tagsLen := len(b.Tags)
//	if tagsLen == 0 {
//		return datatypes.Strings{"default"}
//	}
//
//	// 1st, sort all key values
//	tagKeys := make(datatypes.Strings, 0, tagsLen)
//	for key, _ := range b.Tags {
//		tagKeys = append(tagKeys, key)
//	}
//	tagKeys.Sort()
//
//	// 2nd, get the values for all the keys and place them properly
//	tags := make(datatypes.Strings, tagsLen*2, tagsLen*2)
//	for i := 0; i < tagsLen; i++ {
//		tags[i*2] = tagKeys[i]             // shift the key to proper location
//		tags[(i*2)+1] = b.Tags[tagKeys[i]] // grab the value for the key and place it in the proper location
//	}
//
//	return tags
//}
