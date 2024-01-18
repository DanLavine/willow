package memory

import (
	"sync"
	"time"

	"github.com/DanLavine/willow/internal/heartbeater"
)

type item struct {
	// lock is used to guard all the child objects
	lock             *sync.RWMutex
	data             []byte
	updateable       bool
	retryCount       uint64
	maxRetryAttempts uint64
	retryPosition    string
	heartbeatTimeout time.Duration

	// heartbeater is used to setup and manage the heartbeat process
	heartbeatLock    *sync.RWMutex
	heartbeatProcess heartbeater.Heartbeater
}

func newItem(data []byte, updateable bool, maxRetryAttempts uint64, retryPosition string, heartbeatTimeout time.Duration) *item {
	item := &item{
		lock:             new(sync.RWMutex),
		data:             data,
		updateable:       updateable,
		retryCount:       0,
		maxRetryAttempts: maxRetryAttempts,
		retryPosition:    retryPosition,
		heartbeatTimeout: heartbeatTimeout,
		heartbeatLock:    new(sync.RWMutex),
	}

	return item
}

// onTimeout needs to eventually call queue_channels_client.deleteChannel() callback
func (item *item) CreateHeartbeater(onShutdown, onTimeout func()) heartbeater.Heartbeater {
	item.heartbeatLock.Lock()
	defer item.heartbeatLock.Unlock()

	if item.heartbeatProcess != nil {
		panic("heartbeat process already running")
	}

	heartbeatProcess, err := heartbeater.New(item.heartbeatTimeout, onShutdown, onTimeout)
	if err != nil {
		panic(err)
	}

	item.heartbeatProcess = heartbeatProcess
	return heartbeatProcess
}

// UnsetHeartbeater needs to be called when adding the heartbeater to the async task manager fails
func (item *item) UnsetHeartbeater() {
	item.heartbeatLock.Lock()
	defer item.heartbeatLock.Unlock()

	item.heartbeatProcess = nil
}

func (item *item) StartHeartbeater() bool {
	item.heartbeatLock.Lock()
	defer item.heartbeatLock.Unlock()

	if item.heartbeatProcess != nil {
		return item.heartbeatProcess.Start()
	}

	return false
}

func (item *item) StopHeartbeater() bool {
	item.heartbeatLock.Lock()
	defer item.heartbeatLock.Unlock()

	if item.heartbeatProcess != nil {
		stopped := item.heartbeatProcess.Stop()
		item.heartbeatProcess = nil
		return stopped
	}

	return false
}

func (item *item) Heartbeat() bool {
	item.heartbeatLock.Lock()
	defer item.heartbeatLock.Unlock()

	if item.heartbeatProcess != nil {
		return item.heartbeatProcess.Heartbeat()
	}

	return false
}
