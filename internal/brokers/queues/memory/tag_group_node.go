package memory

import (
	"sync"

	"github.com/DanLavine/willow/internal/brokers/tags"
	"github.com/DanLavine/willow/pkg/models/datatypes"
	v1 "github.com/DanLavine/willow/pkg/models/v1"
)

type tag = func() *v1.DequeueItemResponse

type tagNode struct {
	lock *sync.Mutex

	strictChannel  chan tags.Tag // only used by the tag group if there is one
	generalChannel chan tags.Tag // used for any tag groups that have the provided tags

	// TODO: This is not used yet. But I think will be required when trying to remove items from the tree
	counter int

	tagGroup *tagGroup
}

func tagNodeLock(item any) {
	tagNode := item.(*tagNode)
	tagNode.lock.Lock()
}

// create new generalChannels for all possible tag combinations
func (q *Queue) tagNodeNewGeneralChannel(channels *[]chan tags.Tag) func() (any, error) {
	return func() (any, error) {
		generalChannel := make(chan tags.Tag)
		*channels = append(*channels, generalChannel)

		return &tagNode{
			lock:           new(sync.Mutex),
			strictChannel:  nil,
			generalChannel: generalChannel,
			counter:        0,
			tagGroup:       nil,
		}, nil
	}
}

// create new generalChannels for all possible tag combinations
func (q *Queue) tagNodeNewGeneralChannelRead(channels *[]<-chan tags.Tag) func() (any, error) {
	return func() (any, error) {
		generalChannel := make(chan tags.Tag)
		*channels = append(*channels, generalChannel)

		return &tagNode{
			lock:           new(sync.Mutex),
			strictChannel:  nil,
			generalChannel: generalChannel,
			counter:        0,
			tagGroup:       nil,
		}, nil
	}
}

// create new generalChannels for all possible tag combinations
func (q *Queue) tagNodeNewStrictChannel(channels *[]<-chan tags.Tag) func() (any, error) {
	return func() (any, error) {
		strictChannel := make(chan tags.Tag)
		*channels = append(*channels, strictChannel)

		return &tagNode{
			lock:           new(sync.Mutex),
			strictChannel:  strictChannel,
			generalChannel: make(chan tags.Tag),
			counter:        0,
			tagGroup:       nil,
		}, nil
	}
}

func (q *Queue) newTagNode(tagPairs datatypes.StringMap, channels *[]chan tags.Tag) func() (any, error) {
	return func() (any, error) {
		strictChannel := make(chan tags.Tag)
		generalChannel := make(chan tags.Tag)
		*channels = append(*channels, strictChannel)
		*channels = append(*channels, generalChannel)

		lock := new(sync.Mutex)
		lock.Lock()

		tagNode := &tagNode{
			lock:           lock,
			strictChannel:  strictChannel,
			generalChannel: generalChannel,
			counter:        0,
			tagGroup:       newTagGroup(tagPairs, *channels),
		}

		// start running the tag group in the background to process messages
		return tagNode, q.taskManager.AddExecuteTask("", tagNode.tagGroup)
	}
}

// add a strict channel to a node that already exists
func (q *Queue) tagNodeUpdateStrictChannel(channels *[]chan tags.Tag) func(item any) {
	return func(item any) {
		node := item.(*tagNode)
		node.lock.Lock()
		defer node.lock.Unlock()

		if node.strictChannel == nil {
			node.strictChannel = make(chan tags.Tag)
		}

		*channels = append(*channels, node.strictChannel)
	}
}

// create a tag group for a node that already exists
func (q *Queue) tagNodeUpdateTagGroup(tagPairs datatypes.StringMap, channels *[]chan tags.Tag) func(item any) {
	return func(item any) {
		node := item.(*tagNode)
		node.lock.Lock()

		if node.strictChannel == nil {
			node.strictChannel = make(chan tags.Tag)
		}

		*channels = append(*channels, node.generalChannel)
		*channels = append(*channels, node.strictChannel)
		if node.tagGroup == nil {
			// create a new tagGroup
			node.tagGroup = newTagGroup(tagPairs, *channels)
			// start running the tag group in the background to process messages
			q.taskManager.AddExecuteTask("", node.tagGroup)
		}
	}
}

// used to find the generalChannels for all possible tag combinations
func (q *Queue) tagNodeGetGeneralChannel(channels *[]chan tags.Tag) func(item any) {
	return func(item any) {
		node := item.(*tagNode)
		*channels = append(*channels, node.generalChannel)
	}
}

func (q *Queue) tagNodeGetGeneralChannelRead(channels *[]<-chan tags.Tag) func(item any) {
	return func(item any) {
		node := item.(*tagNode)
		*channels = append(*channels, node.generalChannel)
	}
}

func (q *Queue) tagNodeGetStrictChannel(channels *[]<-chan tags.Tag) func(item any) {
	return func(item any) {
		node := item.(*tagNode)
		node.lock.Lock()
		defer node.lock.Unlock()

		if node.strictChannel == nil {
			node.strictChannel = make(chan tags.Tag)
		}

		*channels = append(*channels, node.strictChannel)
	}
}
