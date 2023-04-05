package tags

import (
	"github.com/DanLavine/willow/internal/datastructures"
	"github.com/DanLavine/willow/internal/helpers"
	v1 "github.com/DanLavine/willow/pkg/models/v1"
)

// TagReaders is used as a record keeper for possible tag combinations. This data structure is used by:
//  1. Queues
//     a. Setup all possible readers when an Item is enqueued as part of the Enqueue process
//     b. Any client trying to find possible tag groups, they pull fromavailable readers here. In some
//     cases they will create the readers if there are none that are usinig those readers yet. It is
//     important to remember that the clients will just hang in that case untill a queue message matches the reader options
type TagReaders interface {
	// Create all possible readers for a particular group of tags. This will create every possible set rather than just the tags provided
	// * used by readers to create a new tag group
	CreateGroup(tags v1.Strings) []chan<- Tag

	// Get the global reader for a particular queue. This can return nil if the queue does not exists
	// * used by clients
	GetGlobalReader() <-chan Tag

	// Get the strict reader for a particular set of tags. Will create the tag and readers if they do not yet exist
	// * used by clients
	GetStrictReader(tags v1.Strings) <-chan Tag

	// Get reader for a particular set of tags. Will create the tag and readers if they do not yet exist
	// * used by clients
	GetSubsetReader(tags v1.Strings) <-chan Tag

	// Get any readers that exists for all provided tags. Will create the tags and readeds if they do not yet exist
	// * used by clients
	GetAnyReaders(tags v1.Strings) []<-chan Tag

	// Remove any tags if they ate no longer being used by any other clients
	RemoveReaders(tag v1.Strings)

	// remove all tags for a specific tag group
	RemoveReadersGroup(tags v1.Strings)
}

// Root structure for the reader tree
type tagReadersTree struct {
	// globala reader all brokers using this TagReaders set
	globalReader chan Tag

	// each element must be of type *tagReadersNode
	readers datastructures.DisjointTree
}

func NewTagReaderTree() *tagReadersTree {
	return &tagReadersTree{
		globalReader: make(chan Tag),
		readers:      datastructures.NewDisjointTree(),
	}
}

// node for each element in the disjoint tree
type tagReadersNode struct {
	// total number of possible readers/clients using this specific tag. On 0 this can be deleted
	//count *atomic.Int64

	// reader for this specific tag
	reader chan Tag
	// strict reader for a queue that is created from all nested tags. Most of the time this will be nil
	strictReader chan Tag
}

func newTagReadersNode(reader chan Tag, setupStrict bool) func() (any, error) {
	return func() (any, error) {
		var strictReader chan Tag
		if setupStrict {
			strictReader = make(chan Tag)
		}

		return &tagReadersNode{
			//count:        new(atomic.Int64),
			reader:       reader,
			strictReader: strictReader,
		}, nil
	}
}

func (trn *tagReadersNode) OnUpdate() {
	if trn.strictReader == nil {
		trn.strictReader = make(chan Tag)
	}
}

// [a,b,c,d,e]
// a,b,c,d,e,ab,ac,ad,ae,abc,abd,abe,acd,ace,ade,abcd,abde,acde,abcde,bc,bd,be,bcd,bce,bde,cd,ce,cde,de -> 1 channel?
// then on a new group
// [a,b,c,d,e,f]
// if any subsets overlap, create a new channel for them all and return 2 channels. This way at most N unique channels, but keep overal count low
//
// PARAMS:
// * tags - string collection that will create all possible groups for
//
// RETURNS:
// * []chan<- Tag - a write only copy of all possible tag group channels. A new channel is used for all 1st time created channels
func (r *tagReadersTree) CreateGroup(tags v1.Strings) []chan<- Tag {
	tagGroups := helpers.GenerateGroupPairs(tags)
	channels := map[chan<- Tag]struct{}{r.globalReader: struct{}{}}

	// reader for all new tags in the provided group if they do not already exist
	reader := make(chan Tag)

	for _, tagGroup := range tagGroups {
		if len(tagGroup) == len(tags) {
			// cannot return an error on our callback so ignore it
			treeItem, _ := r.readers.FindOrCreate(tagGroup, "OnUpdate", newTagReadersNode(reader, true))
			node := treeItem.(*tagReadersNode)
			channels[node.strictReader] = struct{}{}
			channels[node.reader] = struct{}{}
		} else {
			// cannot return an error on our callback so ignore it
			treeItem, _ := r.readers.FindOrCreate(tagGroup, "", newTagReadersNode(reader, false))
			node := treeItem.(*tagReadersNode)
			channels[node.reader] = struct{}{}
		}
	}

	channelSlice := []chan<- Tag{}
	for ch, _ := range channels {
		channelSlice = append(channelSlice, ch)
	}

	// cleanup if all channels are already created
	if _, ok := channels[reader]; !ok {
		close(reader)
	}

	return channelSlice
}

// GetGlobalReader gets the global reader for the TagGroup
//
// RETURNS:
// * <-chan Tag - Read Only copy of the global reader
func (r *tagReadersTree) GetGlobalReader() <-chan Tag {
	return r.globalReader
}

// GeStrictReader gets a strict reader for the given tags
//
// RETURNS:
// * <-chan Tag - Read Only copy of the strict reader if it exists. Otherwise will be nil
func (r *tagReadersTree) GetStrictReader(tags v1.Strings) <-chan Tag {
	// won't return an error
	treeItem, _ := r.readers.FindOrCreate(tags, "OnUpdate", newTagReadersNode(make(chan Tag), true))
	return treeItem.(*tagReadersNode).strictReader
}

// GetSubsetReader gets a strict reader for the given tags
//
// RETURNS:
// * <-chan Tag - Read Only copy of the strict reader if it exists. Otherwise will be nil
func (r *tagReadersTree) GetSubsetReader(tags v1.Strings) <-chan Tag {
	// won't return an error
	treeItem, _ := r.readers.FindOrCreate(tags, "", newTagReadersNode(make(chan Tag), false))
	return treeItem.(*tagReadersNode).reader
}

// GetAnyReaders gets a strict reader for the given tags
//
// RETURNS:
// * []<-chan Tag - Read Only copy of any readers that match any tags
func (r *tagReadersTree) GetAnyReaders(tags v1.Strings) []<-chan Tag {
	potentialReader := map[<-chan Tag]struct{}{}
	readers := []<-chan Tag{}

	newReader := make(chan Tag)
	for _, tag := range tags {
		// won't return an error
		treeItem, _ := r.readers.FindOrCreate(v1.Strings{tag}, "", newTagReadersNode(newReader, false))
		potentialReader[treeItem.(*tagReadersNode).reader] = struct{}{}
	}

	for reader, _ := range potentialReader {
		readers = append(readers, reader)
	}

	return readers
}

func (r *tagReadersTree) RemoveReaders(tags v1.Strings) {
	panic("not implemented")
}
func (r *tagReadersTree) RemoveReadersGroup(tags v1.Strings) {
	panic("not implemented")
}
