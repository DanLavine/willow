package btreeassociated

import (
	"github.com/DanLavine/willow/internal/datastructures"
	"github.com/DanLavine/willow/pkg/models/datatypes"
)

// BTreeAssociated is  way of grouping arbitrary key values into a unique searchable data set.
//
// The tree can be broken into 3 node types:
//  1. groupedKeyValueAssociation - The root level of the Composite Tree is a BTree of Integer Keys and each node is a compositeColumn.
//     Another BTree which contains an the number of key+value pairs == Integer Keys. Using this info we can gurantee
//     that each sub tree at the Integer Node is unique.
//  2. keyValuePairs - The values of the compositeColumn's BTree are then the Keys
//  3. idholders - the values for a unique entire
//
// Example (tree root):
//
//	       4
//	    /      \
//	   2       6,8
//	 /  \    /  | \
//	1    3   5  7  9
//
// If we were to investigate the tree of 3 for something like unique project details, we would see all the 'keys' at this stage
// and the tree would look at a minimum something like: (because it is 3, there needs to be at least 3 keys value groups per item)
// (groupedKeyValueAssociation) - (for tree 3)
//
//	  organization
//	   /        \
//	namespace  project
//
// Where finaly the last tree under 'namespace', which are all the possible values for a namespace could look something like:
// (keyValuePairs) - (index is city)
//
//				  default,staging
//	    /         |        \
//	  active	 pending    test
//
// At this point the Value's under any 'namespace' is a list of unique ID's. Using a set, we can search for any arbitray tags + values
// and do a number of filters to find a particualr subset of data.
//
// For example, getting all three values for map[string]string{organization:123, namespace:default, project:willow} will generate
// 1 unique ID that points for all those search criteria (can be done by using a union for all data between the values returned from each tree).
//
// Similarly, if we instead say something like map[string]string{project:willow} we could just get the list of ID's from project willow. and
// now we have any possible entry that has the project:willow key value pairing
//
// Lastly, we can do something like map[string]string{namespace:*, project:willow} (where star means anything). This could again return all values
// where project:willow key value pairing exists iff they also have a namespace tag.
//
// There are some other constraints that need to be accounted for as well. For example, I need a way of specifying 'key+value limit = 3' otherwise
// we would also need to search the 4-9 trees for any of those values as well since they are an arbitrary collection of tags. But that can come later
// as query params. For now this structre should give us everything we need
type BTreeAssociated interface {
	// Get value associated with a collection of key value pairs
	Find(keyValuePairs datatypes.StringMap, onFind datastructures.OnFind) error

	// CreateOrFind a value associated with a collection of key value pairs
	CreateOrFind(keyValuePairs datatypes.StringMap, onCreate datastructures.OnCreate, onFind datastructures.OnFind) error

	// Iterate over the tree and for each value found invoke the callback with the node's value
	Iterate(callback datastructures.OnFind) error

	// Remove a item from the tree
	Delete(keyValuePairs datatypes.StringMap, canDelete datastructures.CanDelete) error
}
