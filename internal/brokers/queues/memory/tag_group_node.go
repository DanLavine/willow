package memory

import (
	"sync"

	"github.com/DanLavine/willow/internal/brokers/tags"
	"github.com/DanLavine/willow/internal/datastructures"
	"github.com/DanLavine/willow/pkg/models/datatypes"
	v1 "github.com/DanLavine/willow/pkg/models/v1"
)

type tag = func() *v1.DequeueItemResponse

type tagNode struct {
	lock *sync.RWMutex

	strictChannel  chan tags.Tag // only used by the tag group if there is one
	generalChannel chan tags.Tag // used for any tag groups that have the provided tags

	// TODO: This is not used yet. But I think will be required when trying to remove items from the tree
	counter int

	tagGroup *tagGroup
}

// create new generalChannels for all possible tag combinations
func (q *Queue) tagNodeNewGeneralChannel(channels *[]chan tags.Tag) func() any {
	return func() any {
		generalChannel := make(chan tags.Tag)
		*channels = append(*channels, generalChannel)

		return &tagNode{
			lock:           new(sync.RWMutex),
			strictChannel:  nil,
			generalChannel: generalChannel,
			counter:        0,
			tagGroup:       nil,
		}
	}
}

// create new generalChannels for all possible tag combinations
func (q *Queue) tagNodeNewGeneralChannelRead(channels *[]<-chan tags.Tag) func() any {
	return func() any {
		generalChannel := make(chan tags.Tag)
		*channels = append(*channels, generalChannel)

		return &tagNode{
			lock:           new(sync.RWMutex),
			strictChannel:  nil,
			generalChannel: generalChannel,
			counter:        0,
			tagGroup:       nil,
		}
	}
}

// create new generalChannels for all possible tag combinations
func (q *Queue) tagNodeNewStrictChannel(channels *[]<-chan tags.Tag) func() any {
	return func() any {
		strictChannel := make(chan tags.Tag)
		*channels = append(*channels, strictChannel)

		return &tagNode{
			lock:           new(sync.RWMutex),
			strictChannel:  strictChannel,
			generalChannel: make(chan tags.Tag),
			counter:        0,
			tagGroup:       nil,
		}
	}
}

func (q *Queue) newTagNode(tagPairs datatypes.StringMap, channels *[]chan tags.Tag, callback datastructures.OnFind) func() any {
	return func() any {
		strictChannel := make(chan tags.Tag)
		generalChannel := make(chan tags.Tag)
		*channels = append(*channels, strictChannel)
		*channels = append(*channels, generalChannel)

		tagNode := &tagNode{
			lock:           new(sync.RWMutex),
			strictChannel:  strictChannel,
			generalChannel: generalChannel,
			counter:        0,
			tagGroup:       newTagGroup(tagPairs, *channels),
		}

		// start running the tag group in the background to process messages
		_ = q.taskManager.AddExecuteTask("", tagNode.tagGroup)

		callback(tagNode)

		return tagNode
	}
}

// create a tag group for a node that already exists
func (q *Queue) tagNodeUpdateTagGroup(tagPairs datatypes.StringMap, channels *[]chan tags.Tag, callback datastructures.OnFind) func(item any) {
	return func(item any) {
		node := item.(*tagNode)
		node.lock.Lock()
		defer node.lock.Unlock()

		if node.strictChannel == nil {
			node.strictChannel = make(chan tags.Tag)
		}

		*channels = append(*channels, node.generalChannel)
		*channels = append(*channels, node.strictChannel)
		if node.tagGroup == nil {
			// create a new tagGroup
			node.tagGroup = newTagGroup(tagPairs, *channels)
			// start running the tag group in the background to process messages
			_ = q.taskManager.AddExecuteTask("", node.tagGroup)
		}

		callback(node)
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

func (q *Queue) canDeleteTagNode(item any) bool {
	tagNode := item.(*tagNode)

	if tagNode.tagGroup == nil {
		close(tagNode.generalChannel)
		if tagNode.strictChannel != nil {
			close(tagNode.strictChannel)
		}
		return true
	}

	tagNode.tagGroup.lock.Lock()
	defer tagNode.tagGroup.lock.Unlock()
	if tagNode.tagGroup.itemReadyCount.Load()+tagNode.tagGroup.itemProcessingCount.Load() == 0 {
		// no items in the tag node, so delete these
		tagNode.tagGroup.Stop()
		close(tagNode.generalChannel)

		if tagNode.strictChannel != nil {
			close(tagNode.strictChannel)
		}

		return true
	}

	return false
}
