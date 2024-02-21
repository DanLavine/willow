package willowclient

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/DanLavine/willow/pkg/clients"
	"github.com/DanLavine/willow/pkg/models/api"
	"github.com/DanLavine/willow/pkg/models/datatypes"

	"github.com/DanLavine/willow/pkg/models/api/common/errors"
	v1willow "github.com/DanLavine/willow/pkg/models/api/willow/v1"
)

type WillowItem interface {
	Done() <-chan struct{}

	SetHeartbeatErrorCallback(callback func(err error))

	Data() []byte

	Ack(passed bool) error
}

type Item struct {
	doneOnce *sync.Once
	done     chan struct{}

	url    string
	client clients.HttpClient

	data             []byte
	itemID           string
	keyValues        datatypes.KeyValues
	queueName        string
	heartbeatTimeout time.Duration

	heartbeatErrorLock     *sync.RWMutex
	heartbeatErrorCallback func(err error)
}

func newItem(url string, client clients.HttpClient, queueName string, dequeueItem *v1willow.DequeueQueueItem) *Item {
	item := &Item{
		doneOnce: new(sync.Once),
		done:     make(chan struct{}),

		url:    url,
		client: client,

		data:             dequeueItem.Item,
		itemID:           dequeueItem.ItemID,
		keyValues:        dequeueItem.KeyValues,
		queueName:        queueName,
		heartbeatTimeout: dequeueItem.TimeoutDuration,

		heartbeatErrorLock: new(sync.RWMutex),
	}

	go func() {
		defer item.stop()

		// set ticker to be ((timeout - 10%) /3). This way we try and heartbeat at least 3 times before a failure occurs
		ticker := time.NewTicker((item.heartbeatTimeout - (item.heartbeatTimeout / 10)) / 3)
		lastTick := time.Now()

		for {
			select {
			case <-item.done:
				// ack was called can just escape
				return
			case <-ticker.C:
				// in this case, we know that we have timed out and don't need to heartbeat anymore
				if item.heartbeatTimeout > time.Since(lastTick) {
					return
				}

				resp, err := item.client.Do(&clients.RequestData{
					Method: "POST",
					Path:   fmt.Sprintf("%s/v1/queues/%s/channels/items/heartbeat", item.url, item.queueName),
					Model: &v1willow.Heartbeat{
						ItemID:    item.itemID,
						KeyValues: item.keyValues,
					},
				})

				// error making the request. This should not happen
				if err != nil {
					select {
					case <-item.done:
						//nothing to do here. race between ack and heartbeat
					default:
						item.forwardError(err)
						continue
					}
				}

				// parse the response
				switch resp.StatusCode {
				case http.StatusOK:
					// sent a heartbeat
					lastTick = time.Now()
				case http.StatusBadRequest, http.StatusConflict, http.StatusInternalServerError:
					// faild to heartbeat for some reason
					apiError := &errors.Error{}
					if err := api.DecodeAndValidateHttpResponse(resp, apiError); err != nil {
						select {
						case <-item.done:
							//nothing to do here. race between ack and heartbeat
						default:
							item.forwardError(err)
						}
					} else {
						select {
						case <-item.done:
							//nothing to do here. race between ack and heartbeat
						default:
							item.forwardError(apiError)
						}
					}
				// case:
				default:
					item.forwardError(fmt.Errorf("unexpected status code while heartbeating: %d", resp.StatusCode))
				}
			}
		}
	}()

	return item
}

// add or update the callback function for when a heartbeat happens
func (item *Item) SetHeartbeatErrorCallback(callback func(err error)) {
	item.heartbeatErrorLock.Lock()
	defer item.heartbeatErrorLock.Unlock()

	item.heartbeatErrorCallback = callback
}

// done can be used to monitor if the heartbeater has failed and the item should be counted as a failure
func (item *Item) Done() <-chan struct{} {
	return item.done
}

func (item *Item) ACK(passed bool) error {
	item.stop()

	resp, err := item.client.Do(&clients.RequestData{
		Method: "POST",
		Path:   fmt.Sprintf("%s/v1/queues/%s/channels/items/ack", item.url, item.queueName),
		Model: &v1willow.ACK{
			ItemID:    item.itemID,
			KeyValues: item.keyValues,
			Passed:    passed,
		},
	})

	if err != nil {
		return err
	}

	// parse the response
	switch resp.StatusCode {
	case http.StatusOK:
		// nothg to do here
		return nil
	case http.StatusBadRequest, http.StatusNotFound, http.StatusConflict, http.StatusInternalServerError:
		// faild to ack for some reason
		apiError := &errors.Error{}
		if err := api.DecodeAndValidateHttpResponse(resp, apiError); err != nil {
			return err
		}

		return apiError
	default:
		item.forwardError(fmt.Errorf("unexpected status code while heartbeating: %d", resp.StatusCode))
	}

	return nil
}

// get the dequeued item
func (item *Item) Data() []byte {
	return item.data
}

func (item *Item) forwardError(err error) {
	item.heartbeatErrorLock.RLock()
	defer item.heartbeatErrorLock.RUnlock()

	if item.heartbeatErrorCallback != nil {
		item.heartbeatErrorCallback(err)
	}
}

func (item *Item) stop() {
	item.doneOnce.Do(func() {
		close(item.done)
	})
}

// What I think could be useful for a reload  on a process that has stopped for an update and restarted. I.E K8S node's would still
// have docker images for running JOBS that could be picked up and restart heartbeating that they are still processing. In this case
// the joibs would have a long time to run, but that is to be expected in those use cases

// func (item *Item) ReleaseWithoutAck() { }}

// func (item *Item) Save(diskDir string) {}

// func LoadItem(diskDir string) *Item { }
