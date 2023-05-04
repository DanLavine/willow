package memory

import (
	"sync"

	"github.com/DanLavine/willow/internal/v1/tags"
	"github.com/DanLavine/willow/pkg/models/datatypes"
	v1 "github.com/DanLavine/willow/pkg/models/v1"
)

type tag = func() *v1.DequeueItemResponse

type tagNode struct {
	lock *sync.Mutex

	channel chan tags.Tag
	counter int

	tagGroup *tagGroup
}

func (q *Queue) newTagNode(tags datatypes.Strings, channels []chan<- tags.Tag) func() (*tagNode, error) {
	return func() (*tagNode, error) {
		channel := make(chan Tag)
		channels = append(channel, channels)

		tagNode := &tagNode{
			channel:  channel,
			counter:  0,
			tagGroup: newTagGroup(tags, channels),
		}

		// start running the tag group in the background to process messages
		return tagNode, q.taskManager.AddExecuteTask("", tagNode.tagGroup)
	}
}

func (q *Queue) updateTagNode(tags datatypes.Strings, channels []chan<- tags.Tag) func(item any) {
	return func(item any) {
		node := item.(*tagNode)
		node.lock.Lock()
		defer node.lock.Unlock()

		if node.tagGroup == nil {
			node.tagGroup = newTagGroup(tags, channels)

			// start running the tag group in the background to process messages
			q.taskManager.AddExecuteTask("", tagNode.tagGroup)
		}
	}
}

// used to find the channels for all possible tag combinations
func (q *Queue) findChannels(channels []chan<- tags.Tag) func(item any) {
	return func(item any) {
		node := item.(*tagNode)
		channels = append(channels, node.channel)
	}
}

// create new channels for all possible tag combinations
func (q *Queue) newChannels(channels []chan<- tags.Tag) func() (*tagNode, error) {
	return func() (*tagNode, error) {
		channel := make(chan Tag)
		channels = append(channels, channel)

		return &tagNode{
			channel:  channel,
			counter:  0,
			tagGroup: nil,
		}, nil
	}
}
