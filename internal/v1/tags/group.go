package tags

import (
	"fmt"
	"sync"

	"github.com/DanLavine/willow/internal/datastructures"
)

type TagsGroup interface {
	// FindOrCreateTagsGroup can be used to setup a tag group.
	FindOrCreateTagsGroup(tags []string, onCreate func() datastructures.TreeItem) (any, error)
}

// tags group tree is the tree structure for recording tags. Each tagsGroupTree is a BTree
// where each value in the tree is a tagsGroup structure. The Trees are recursive in nature when
// creating multiple tags.
//
// I.E:
//   - The Root tree could be structured like so
//     Cow
//     /      \
//     Bat     Farm
//
// Then Under Farm there could be another tree structure of:
//
//	    Group
//	  /        \
//	Berries   HayStacks
//
// On any Operations (Find, Create, Delete, etc). when calling []string{"Farm", "HayStack"} we can look at
// the root level for "Farm" and then recurse into the tree at that level to find "HayStack"
type tagsGroupTree struct {
	tree datastructures.BTree
}

type tagsGroup struct {
	lock     *sync.Mutex
	value    datastructures.TreeItem
	children *tagsGroupTree
}

func (tg *tagsGroup) OnFind() {
	tg.lock.Lock()

	if tg.value != nil {
		tg.value.OnFind()
	}
}

func NewTagsGroup() *tagsGroupTree {
	tree, err := datastructures.NewBTree(2)
	if err != nil {
		panic(err)
	}

	return &tagsGroupTree{
		tree: tree,
	}
}

// wraper to create a tagsGroup with a value
func (tgt *tagsGroupTree) newTagsGroupTreeWithValue(onCreate func() datastructures.TreeItem) func() datastructures.TreeItem {
	return func() datastructures.TreeItem {
		lock := new(sync.Mutex)
		lock.Lock()

		return &tagsGroup{
			lock:     lock,
			value:    onCreate(),
			children: NewTagsGroup(),
		}
	}
}

// wraper to create a tagsGroup without a value
func (tgt *tagsGroupTree) newTagsGroupTree() func() datastructures.TreeItem {
	return func() datastructures.TreeItem {
		lock := new(sync.Mutex)
		lock.Lock()

		return &tagsGroup{
			lock:     lock,
			value:    nil,
			children: NewTagsGroup(),
		}
	}
}

// FindOrCreateTagsGroup is used to setup a new TagsGroup if onde does not already exists. The onCreate()
// function is a callback to any queue provides to setup the actual tag group. If the tag group does exists,
// then the OnFind() method for that tag group will be called instead of the setup function
//
// PARAMS:
// * tags - all the tags to create the unique nested tree structure. On the first calue the value() func will be called and saved. On the n+ calls if the item exists, then thesaved value will be returned
// * onCreate - callback func provided from the queue implemention to create their own tag groups
//
// RETURNS
// * datastructures.TreeItem - the return value of the 'value' param
// * v1.Error - any errors encountered
func (tgt *tagsGroupTree) FindOrCreateTagsGroup(tags []string, onCreate func() datastructures.TreeItem) (datastructures.TreeItem, error) {
	switch len(tags) {
	case 0:
		return nil, fmt.Errorf("received an invalid tags length. Needs to be at least 1")
	case 1:
		treeItem := tgt.tree.FindOrCreate(datastructures.NewStringTreeKey(tags[0]), tgt.newTagsGroupTreeWithValue(onCreate))
		tagsGroup := treeItem.(*tagsGroup)
		defer tagsGroup.lock.Unlock()

		if tagsGroup.value == nil {
			tagsGroup.value = onCreate()
		}

		return tagsGroup.value, nil
	default:
		tagsGroup := tgt.tree.FindOrCreate(datastructures.NewStringTreeKey(tags[0]), tgt.newTagsGroupTree()).(*tagsGroup)
		defer tagsGroup.lock.Unlock()

		return tagsGroup.children.FindOrCreateTagsGroup(tags[1:], onCreate)
	}
}
