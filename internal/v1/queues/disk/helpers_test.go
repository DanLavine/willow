package disk_test

import (
	"context"
	"os"
	"time"

	"github.com/DanLavine/willow/internal/v1/models"
	"github.com/DanLavine/willow/internal/v1/queues/disk"
	"github.com/DanLavine/willow/internal/v1/queues/disk/encoder"
	v1 "github.com/DanLavine/willow/pkg/models/v1"
	"github.com/onsi/gomega"

	. "github.com/onsi/gomega"
)

var defaultTags = []string{"one", "two"}

type diskQueueTestManager struct {
	dataDir string

	dq              *disk.DiskQueue
	deadLetterQueue *disk.DiskDeadLetterQueue
	encoderQueue    *encoder.EncoderQueue

	readers []chan *models.Location
	tags    []string
}

func SetupDiskQueue(g *gomega.WithT, tags []string) *diskQueueTestManager {
	testDir, err := os.MkdirTemp("", "")
	g.Expect(err).ToNot(HaveOccurred())

	if tags == nil {
		tags = defaultTags
	}

	// general create queue params
	queueParams := &v1.QueueParams{MaxSize: 5, RetryCount: 2}

	// general encoder queue
	encoderQueue, err := encoder.NewEncoderQueue(testDir, tags)
	g.Expect(err).ToNot(HaveOccurred())

	// general dead letter queue
	diskEncoder, err := encoder.NewEncoderDeadLetter(testDir, tags)
	g.Expect(err).ToNot(HaveOccurred())

	deadLetterQueue, err := disk.NewDiskDeadLetterQueue(&v1.DeadLetterQueueParams{MaxSize: 5}, diskEncoder)
	g.Expect(err).ToNot(HaveOccurred())

	// general readers
	readers := []chan *models.Location{make(chan *models.Location)}

	dq, dqErr := disk.NewDiskQueue(tags, queueParams, encoderQueue, deadLetterQueue, readers)
	g.Expect(dqErr).ToNot(HaveOccurred())
	g.Expect(dq).ToNot(BeNil())

	return &diskQueueTestManager{
		dataDir:         testDir,
		dq:              dq,
		deadLetterQueue: deadLetterQueue,
		encoderQueue:    encoderQueue,
		readers:         readers,
		tags:            tags,
	}
}

func (dqtm *diskQueueTestManager) GetMessage(g *gomega.WithT) *v1.DequeueMessage {
	cdl, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second))

	var location *models.Location
	g.Eventually(func() *models.Location {
		select {
		case <-cdl.Done():
		case location = <-dqtm.readers[0]:
		}

		return location
	}, time.Second, 100*time.Millisecond, cdl).ShouldNot(BeNil())
	cancel()

	dequeueMessage, err := location.Process()
	g.Expect(err).ToNot(HaveOccurred())

	return dequeueMessage
}

func (dqtm *diskQueueTestManager) Clean() {
	dqtm.dq.Close()
	dqtm.encoderQueue.Close()
	dqtm.deadLetterQueue.Close()

	for index, _ := range dqtm.readers {
		close(dqtm.readers[index])
	}

	os.RemoveAll(dqtm.dataDir)
}
