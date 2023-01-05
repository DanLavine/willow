package disk

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sync"

	"github.com/DanLavine/gonotify"
	v1 "github.com/DanLavine/willow-message/protocol/v1"
	"github.com/DanLavine/willow/pkg/models"
)

// I think this works for validating base64 values.
// base64 cannot end with a '===' so we shouldn't need to look for those
var base64Regexp = regexp.MustCompile(`^([A-Za-z0-9+/]{4})*([A-Za-z0-9+/]{4}|[A-Za-z0-9+/]{3}=|[A-Za-z0-9+/]{2}==)$`)

type diskEncoder struct {
	lock *sync.RWMutex

	notifier *gonotify.Notify

	// details on where to read and write files to on disk
	baseDir    string
	brokerName string
	brokerTag  string

	/////// state tracking for location on disk we are writting to
	// what index we are working on
	nextInsertIndex int
	nextReadIndex   int
	// the location in the file we are writing at
	writeLocation int

	// list of all indexes that have yet to be processed
	indexes map[int]*location

	queue *os.File
}

// Create a new disk encoder.
//
// PARAMS:
// * baseDir - base volume that all disk based storage will use as a root
// * tag     - specific tag that coressponds to a queue name
//
// RETURNS:
// * diskEncoder - disk encoder
// * error       - any errors associated with creating the files or dirs for reading and writting
func NewDiskEncoder(baseDir, brokerName, brokerTag string) (*diskEncoder, *v1.Error) {
	queueDir := filepath.Join(baseDir, brokerName, brokerTag)

	filePath, err := os.Stat(queueDir)
	if os.IsPermission(err) || os.IsNotExist(err) {
		// create the dir
		if err = os.MkdirAll(queueDir, 0755); err != nil {
			return nil, &v1.Error{Message: fmt.Sprintf("Failed to create dir: %s", err.Error()), StatusCode: http.StatusInternalServerError}
		}
	} else if err != nil {
		// some other error encountered
		return nil, &v1.Error{Message: fmt.Sprintf("Failed to stat dir: %s", err.Error()), StatusCode: http.StatusInternalServerError}
	} else {
		// path already exists and is not dir?
		if !filePath.IsDir() {
			return nil, &v1.Error{Message: "Path already exist, but is not a directory", StatusCode: http.StatusInternalServerError}
		}
	}

	queueFile, err := os.OpenFile(filepath.Join(baseDir, brokerName, brokerTag, queueFile), os.O_CREATE|os.O_RDWR|os.O_APPEND, 0755)
	if err != nil {
		return nil, &v1.Error{Message: err.Error(), StatusCode: http.StatusInternalServerError}
	}

	return &diskEncoder{
		lock:            new(sync.RWMutex),
		notifier:        gonotify.New(),
		baseDir:         baseDir,
		brokerName:      brokerName,
		brokerTag:       brokerTag,
		nextInsertIndex: 0,
		nextReadIndex:   0,
		writeLocation:   0,
		indexes:         map[int]*location{},
		queue:           queueFile,
	}, nil
}

// Write a value to disk
func (de *diskEncoder) Enqueue(data []byte) *v1.Error {
	bytesLength := base64.StdEncoding.EncodedLen(len(data))
	encoded := make([]byte, bytesLength+1)
	base64.StdEncoding.Encode(encoded, data)
	encoded[bytesLength] = '.' // write the last index as the sepration char

	de.lock.Lock()
	defer de.lock.Unlock()

	n, err := de.queue.Write(encoded)
	if err != nil {
		return &v1.Error{Message: err.Error(), StatusCode: http.StatusInternalServerError}
	}

	// add the index to unprocessed entities
	de.indexes[de.nextInsertIndex] = &location{start: de.writeLocation, size: n - 1}

	// update the next Enqueue location
	de.writeLocation += n
	de.nextInsertIndex++
	_ = de.notifier.Add() // even if shutting down and this returns an error, thats fine. will load on a restart

	return nil
}

func (de *diskEncoder) Get(id int) (*v1.DequeueMessage, *v1.Error) {
	de.lock.Lock()
	defer de.lock.Unlock()

	// find the location
	location, ok := de.indexes[id]
	if !ok {
		err := &v1.Error{Message: "Index does not exist", StatusCode: http.StatusBadRequest}
		return nil, err.Expected(fmt.Sprintf("ID %d to not be empty", id))
	}

	// read the data from disk
	data := make([]byte, location.size)
	if _, err := de.queue.ReadAt(data, int64(location.start)); err != nil {
		return nil, &v1.Error{Message: err.Error(), StatusCode: http.StatusInternalServerError}
	}

	// decode data we read from disk
	dataLen := base64.StdEncoding.DecodedLen(len(data))
	decoded := make([]byte, dataLen)
	_, err := base64.StdEncoding.Decode(decoded, data)
	if err != nil {
		return nil, &v1.Error{Message: fmt.Sprintf("Failed decoding: %s", err.Error()), StatusCode: http.StatusInternalServerError}
	}

	return &v1.DequeueMessage{
		ID:         uint64(id),
		BrokerName: de.brokerName,
		BrokerTag:  de.brokerTag,
		Data:       decoded,
	}, nil
}

func (de *diskEncoder) Requeue(id int) *v1.Error {
	de.lock.Lock()
	defer de.lock.Unlock()

	if item, ok := de.indexes[id]; ok {
		item.retryCount++
		return nil
	}

	err := &v1.Error{Message: "Index does not exist", StatusCode: http.StatusBadRequest}
	return err.Expected(fmt.Sprintf("ID %d to not be empty", id))
}

func (de *diskEncoder) Remove(id int) *v1.Error {
	de.lock.Lock()
	defer de.lock.Unlock()

	delete(de.indexes, id)

	return nil
}

func (de *diskEncoder) Next() (*v1.DequeueMessage, *v1.Error) {
	select {
	case _, ok := <-de.notifier.Ready():
		if ok {
			message, err := de.Get(de.nextReadIndex)
			if err != nil {
				return nil, err
			}

			// NOTE don't need a lock since only 1 of these should be running at once
			de.nextReadIndex++

			return message, nil
		}

		// closed
		return nil, nil
	}
}

func (de *diskEncoder) Metrics() models.QueueMetrics {
	de.lock.Lock()
	defer de.lock.Unlock()

	return models.QueueMetrics{
		Ready:      len(de.indexes),
		Processing: 0,
	}
}

// Close shuts down and waits for any pending ACK messages to complete, but
// it won't try to drain the queue since it will be loaded on the restart
func (de *diskEncoder) Close() error {
	// TODO
	return nil
}
