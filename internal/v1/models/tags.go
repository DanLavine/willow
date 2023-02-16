package models

import (
	"container/list"
	"strings"
	"sync"

	"github.com/DanLavine/willow/internal/datastructures"
)

type Tags struct {
	lock *sync.Mutex
	tags []string
	tag  string

	lastItem    any
	itemTracker *datastructures.IDTree
}

func NewTags(tags []string) *Tags {
	items := list.New()

	return &Tags{
		lock:        new(sync.Mutex),
		tags:        tags,
		tag:         strings.Join(tags),
		itemTracker: datastructures.NewIDTree(),
	}
}

func (t *Tags) Less(compare datastructures.TreeItem) bool {
	return t.tag < compare.(*Tags).tag
}

func (t *Tags) GetLast() any {
	return t.lastItem
}

// enqueue an item
func (t *Tags) Enqueue(item any) {
	t.lastItem = t.itemTracker.Add(item)
}

// update the last item
func (t *Tags) UpdateLast(item any) {
	if t.LastItem == nil {
		t.lastItem = t.itemTracker.Add(item)
		return
	}

	// remove the last item
	_ = t.itemTracker.Remove(t.lastItem)

	// re-asign to new item
	t.lastItem = t.itemTracker.Add(item)
}

func (t *Tags) Lock() {
	t.lock.Lock()
}

func (t *Tags) Unlock() {
	t.lock.Unlock()
}
