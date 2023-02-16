package datastructures

import (
	"sync"
)

// TagTree is used to generate a tree like struture for a group of tag enqueue messages
// One of the main advantages is to try and find any kund of "subset" for a group of tags
// Example:
//
// If multi$iple requests come in with a list of tags such as (tags must always be sorted):
// 1. [a, b, c]
// 2. [a, c, e]
// 3. [b, c]
//
// It would generate the following tre structure:
/*
					root (no tags)
				 /				\
				a         b
			/   \	     /
		 b     c    c
		/     /
	 c     e
*/
// so in this case we have 2 groups if we wanted to find the subset [b, c]
// There is the entire tree [b,c], but also the tree [a,b,c]. On lookups, we can
// ignore the group [a, c, e], since c > b, we can ignore the enire tree and any children
// in the node after

type TagTree struct {
	lock *sync.RWMutex

	// Tag for the specific tree
	tag string

	// Any enqueued values in the tree
	values []any

	// Any child trees
	children []*TagTree
}

func NewTagTree() *TagTree {
	return &TagTree{
		lock:     new(sync.RWMutex),
		tag:      "",
		values:   nil,
		children: nil,
	}
}

func newTagTreeNode(tag string, value any) *TagTree {
	return &TagTree{
		lock:     new(sync.RWMutex),
		tag:      tag,
		values:   []any{value},
		children: nil,
	}
}

// add a value to the interface
func (tt *TagTree) AddValue(tags []string, value any) {
	// at the node to add a value
	if len(tags) == 0 {
		tt.lock.Lock()
		defer tt.lock.Unlock()

		tt.values = append(tt.values, value)
		return
	}

	// try recursing into child node
	tt.lock.RLock()
	pivotIndex := 0
	for _, child := range tt.children {
		if child.tag == tags[0] {
			defer tt.lock.RUnlock()
			child.AddValue(tags[1:], value)
			return
		}

		// if child tag larger than index we are after, break early
		if child.tag > tags[0] {
			break
		}

		// update index where we are possibly ending
		pivotIndex++
	}
	tt.lock.RUnlock() // didn't find the node, so *possibly* need to add. could have 2 simltanious requests trying to make the same tree

	// need to *possibly* create the new node. double check that another request didn't make the node
	tt.lock.Lock()
	defer tt.lock.Unlock()

	// adding the first index
	if pivotIndex == 0 && len(tt.children) == 0 {
		tt.children = []*TagTree{newTagTreeNode(tags[0], value)}
		return
	}

	// adding the last index
	if len(tt.children) == pivotIndex {
		tt.children = append(tt.children, newTagTreeNode(tags[0], value))
		return
	}

	switch tt.children[pivotIndex].tag > tags[0] {
	case true:
		// recurse backwards
		for searchIndex := pivotIndex; searchIndex >= 0; searchIndex-- {
			// another request creted the tags
			if tt.children[searchIndex].tag == tags[0] {
				tt.children[searchIndex].AddValue(tags[1:], value)
				return
			}

			// instert in the middle
			if tt.children[searchIndex].tag < tags[0] {
				// recreate children
				newChildren := append(tt.children[:searchIndex+1], tt.children[searchIndex:]...) // grow the children slice by 1 where we are inserting the new index
				newChildren[searchIndex+1] = newTagTreeNode(tags[0], value)
				tt.children = newChildren
				return
			}

			// insert at the front
			if searchIndex == 0 {
				tt.children = append([]*TagTree{newTagTreeNode(tags[0], value)}, tt.children...)
				return
			}
		}
	default:
		// recurse forwards
		for searchIndex, child := range tt.children[pivotIndex:] {
			// another request creted the tags
			if child.tag == tags[0] {
				child.AddValue(tags[1:], value)
				return
			}

			// insert in the middle
			if child.tag > tags[0] {
				// recreate children
				newChildren := append(tt.children[:pivotIndex+searchIndex+1], tt.children[pivotIndex+searchIndex:]...) // grow the children slice by 1 where we are inserting the new index
				newChildren[pivotIndex+searchIndex] = newTagTreeNode(tags[0], value)
				tt.children = newChildren
				return
			}
		}

		// insert at the end
		tt.children = append(tt.children, newTagTreeNode(tags[0], value))
	}
}
