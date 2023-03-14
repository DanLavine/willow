package tags

import (
	"sync/atomic"

	"github.com/DanLavine/willow/internal/datastructures"
	"github.com/DanLavine/willow/internal/helpers"
)

// Individual tag tree keeps track of any readers for associated tags. These tags can be used for clients
// requesting "any" readers on a match restrcition
type Readers interface {
	// Create tags for a new TagGroup
	CreateTagsGroup(tags []string) []chan any

	// Get the global reader for all tag groups
	GetGlobalReader() chan any

	// Add a collection of tags to the tag tree
	GetReaders(tags []string) []chan any

	// Add a specific tag to the tag treeand return the reader. If the tag already exists, the usage ount is increased by 1
	GetReader(tag string) chan any

	// remove a tag if it is no longer being used by any other tag groups
	RemoveTag(tag string)

	// remove any tags if they ate no longer being used by any other groups
	RemoveTags(tag []string)

	// remove all tags for a specific tag group
	RemoveTagsGroup(tags []string)
}

// reader is a "tag" for a queue that a client might be interested in reading from
type reader struct {
	count   *atomic.Int32
	channel chan any
}

func newReader() *reader {
	return &reader{
		count:   new(atomic.Int32),
		channel: make(chan any),
	}
}

func (r *reader) OnFind() {
	r.count.Add(1)
}

func (r *reader) CanDestroy() bool {
	total := r.count.Add(-1)
	if total <= 0 {
		return true
	}

	return false
}

// readerTree is used to structure any possible "tags" or "tag combinations" for a queue and save them.
// They should be automatically removed when there are no more clients/queues for a particular "tag"
type readerTree struct {
	global chan any
	tree   datastructures.BTree
}

func NewReaderTree() *readerTree {
	tree, err := datastructures.NewBTree(2)
	if err != nil {
		panic(err)
	}

	return &readerTree{
		global: make(chan any),
		tree:   tree,
	}
}

// [a,b,c,d,e]
// a,b,c,d,e,ab,ac,ad,ae,abc,abd,abe,acd,ace,ade,abcd,abde,acde,abcde,bc,bd,be,bcd,bce,bde,cd,ce,cde,de -> 1 channel?
// then on a new group
// [a,b,c,d,e,f]
// if any subsets overlap, create a new channel for them all and return 2 channels. This way at most N unique channels, but keep overal count low
func (r *readerTree) CreateTagsGroup(tags []string) []chan any {
	newReader := newReader()
	channels := map[chan any]struct{}{newReader.channel: struct{}{}, r.global: struct{}{}}

	usingReader := false
	for _, tag := range helpers.GenerateStringPairs(tags) {
		treeItem := r.tree.FindOrCreate(datastructures.NewStringTreeKey(tag), newReader)
		if treeItem != newReader {
			channels[treeItem.(*reader).channel] = struct{}{}
		} else {
			usingReader = true
		}
	}

	if !usingReader {
		// all tags already exists. can garbage collect the newReader and the channel
		delete(channels, newReader.channel)
		close(newReader.channel)
	}

	chans := make([]chan any, 0, len(channels))
	for channel, _ := range channels {
		chans = append(chans, channel)
	}

	return chans
}

func (r *readerTree) GetGlobalReader() chan any {
	return r.global
}

func (r *readerTree) GetReaders(tags []string) []chan any {
	readers := map[chan any]struct{}{}
	for _, tag := range tags {
		readers[r.GetReader(tag)] = struct{}{}
	}

	returnReaders := make([]chan any, 0, len(readers))
	for reader, _ := range readers {
		returnReaders = append(returnReaders, reader)
	}

	return returnReaders
}

func (r *readerTree) GetReader(tag string) chan any {
	treeItem := r.tree.FindOrCreate(datastructures.NewStringTreeKey(tag), newReader())
	reader := treeItem.(*reader)

	return reader.channel
}

func (r *readerTree) RemoveTags(tags []string) {
	panic("todo")
	//for _, tag := range tags {
	//	r.RemoveTag(tag)
	//}
}

func (r *readerTree) RemoveTag(tag string) {
	panic("todo")
	//_ := r.tree.Remove(datastructures.NewStringTreeKey(tag))
}

func (r *readerTree) RemoveTagsGroup(tags []string) {
	panic("todo")
}
