package disk_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/DanLavine/willow/internal/errors"
	"github.com/DanLavine/willow/internal/v1/models"
	"github.com/DanLavine/willow/internal/v1/queues/disk"
	"github.com/DanLavine/willow/internal/v1/queues/disk/encoder"
	v1 "github.com/DanLavine/willow/pkg/models/v1"
	. "github.com/onsi/gomega"
)

func TestQueue_NewDiskQueue(t *testing.T) {
	g := NewGomegaWithT(t)
	tags := []string{"test"}

	testDir, dirErr := os.MkdirTemp("", "")
	g.Expect(dirErr).ToNot(HaveOccurred())
	defer os.RemoveAll(testDir)

	// general create queue params
	queueParams := &v1.QueueParams{MaxSize: 5, RetryCount: 0}

	// general encoder queue
	encoderQueue, err := encoder.NewEncoderQueue(testDir, tags)
	g.Expect(err).ToNot(HaveOccurred())
	defer encoderQueue.Close()

	// general dead letter queue
	diskEncoder, err := encoder.NewEncoderDeadLetter(testDir, tags)
	g.Expect(err).ToNot(HaveOccurred())

	deadLetterQueue, err := disk.NewDiskDeadLetterQueue(&v1.DeadLetterQueueParams{MaxSize: 5}, diskEncoder)
	g.Expect(err).ToNot(HaveOccurred())
	defer deadLetterQueue.Close()

	// general readers
	readers := []chan *models.Location{make(chan *models.Location)}
	defer close(readers[0])

	t.Run("returns an error if CreatQueueParams is nil", func(t *testing.T) {
		_, dqErr := disk.NewDiskQueue(tags, nil, encoderQueue, deadLetterQueue, readers)
		g.Expect(dqErr).To(HaveOccurred())
		g.Expect(dqErr).To(Equal(errors.NoCreateQueueParams))
	})

	t.Run("returns an error if the encoder is nil", func(t *testing.T) {
		_, dqErr := disk.NewDiskQueue(tags, queueParams, nil, deadLetterQueue, readers)
		g.Expect(dqErr).To(HaveOccurred())
		g.Expect(dqErr).To(Equal(errors.NoEncoder))
	})

	t.Run("returns an error if the readers are nil", func(t *testing.T) {
		_, dqErr := disk.NewDiskQueue(tags, queueParams, encoderQueue, deadLetterQueue, nil)
		g.Expect(dqErr).To(HaveOccurred())
		g.Expect(dqErr).To(Equal(errors.NoReaders))
	})

	t.Run("returns an error if any readers are nil", func(t *testing.T) {
		_, dqErr := disk.NewDiskQueue(tags, queueParams, encoderQueue, deadLetterQueue, []chan *models.Location{nil})
		g.Expect(dqErr).To(HaveOccurred())
		g.Expect(dqErr).To(Equal(errors.NilReader))
	})

	t.Run("accepts a nil deadLetterQueue. This implies we don't want one configured", func(t *testing.T) {
		_, dqErr := disk.NewDiskQueue(tags, queueParams, encoderQueue, nil, readers)
		g.Expect(dqErr).ToNot(HaveOccurred())
	})
}

func TestQueue_Enqueueue(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("properly updates the ready count", func(t *testing.T) {
		dqtm := SetupDiskQueue(g, nil)
		defer dqtm.Clean()

		g.Expect(dqtm.dq.Enqueue([]byte("first"))).ToNot(HaveOccurred())
		metrics := dqtm.dq.Metrics()
		g.Expect(metrics.Ready).To(Equal(uint64(1)))
		g.Expect(metrics.Processing).To(Equal(uint64(0)))

		g.Expect(dqtm.dq.Enqueue([]byte("second"))).ToNot(HaveOccurred())
		metrics = dqtm.dq.Metrics()
		g.Expect(metrics.Ready).To(Equal(uint64(2)))
		g.Expect(metrics.Processing).To(Equal(uint64(0)))
	})

	t.Run("allows the enqueued item to be returned on a read channel", func(t *testing.T) {
		dqtm := SetupDiskQueue(g, nil)
		defer dqtm.Clean()

		g.Expect(dqtm.dq.Enqueue([]byte("first"))).ToNot(HaveOccurred())

		context, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second))
		defer cancel()

		var location *models.Location
		g.Eventually(func() *models.Location {
			select {
			case <-context.Done():
			case location = <-dqtm.readers[0]:
			}

			return location
		}, time.Second, 100*time.Millisecond, context).ShouldNot(BeNil())
	})

	t.Run("returns enqueued items in the order they were added", func(t *testing.T) {
		dqtm := SetupDiskQueue(g, nil)
		defer dqtm.Clean()

		g.Expect(dqtm.dq.Enqueue([]byte("first"))).ToNot(HaveOccurred())
		g.Expect(dqtm.dq.Enqueue([]byte("second"))).ToNot(HaveOccurred())

		dequeueMessage := dqtm.GetMessage(g)
		g.Expect(dequeueMessage.Data).To(Equal([]byte(`first`)))

		dequeueMessage = dqtm.GetMessage(g)
		g.Expect(dequeueMessage.Data).To(Equal([]byte(`second`)))
	})

	t.Run("updates the metric counts properly", func(t *testing.T) {
		dqtm := SetupDiskQueue(g, nil)
		defer dqtm.Clean()

		g.Expect(dqtm.dq.Enqueue([]byte("first"))).ToNot(HaveOccurred())

		dequeueMessage := dqtm.GetMessage(g)
		g.Expect(dequeueMessage.Data).To(Equal([]byte(`first`)))

		metrics := dqtm.dq.Metrics()
		g.Expect(metrics.Ready).To(Equal(uint64(0)))
		g.Expect(metrics.Processing).To(Equal(uint64(1)))
	})
}

func TestQueue_ACK(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("returns an error if the item is not found", func(t *testing.T) {
		dqtm := SetupDiskQueue(g, nil)
		defer dqtm.Clean()

		ackErr := dqtm.dq.ACK(32, false)
		g.Expect(ackErr).To(HaveOccurred())
		g.Expect(ackErr.Error()).To(ContainSubstring("Item not found"))
	})

	t.Run("returns an error if the item is not processing", func(t *testing.T) {
		dqtm := SetupDiskQueue(g, nil)
		defer dqtm.Clean()

		g.Expect(dqtm.dq.Enqueue([]byte(`hello world`))).ToNot(HaveOccurred())

		ackErr := dqtm.dq.ACK(1, false)
		g.Expect(ackErr).To(HaveOccurred())
		g.Expect(ackErr.Error()).To(ContainSubstring("Item not processing"))
	})

	t.Run("when passed it TRUE", func(t *testing.T) {
		t.Run("it removes the processing item and records that it has been deleted", func(t *testing.T) {
			dqtm := SetupDiskQueue(g, nil)
			defer dqtm.Clean()

			g.Expect(dqtm.dq.Enqueue([]byte(`hello world`))).ToNot(HaveOccurred())
			dequeueMessage := dqtm.GetMessage(g)

			g.Expect(dqtm.dq.ACK(dequeueMessage.ID, true)).ToNot(HaveOccurred())

			// queue item state file
			dataDirs := []string{dqtm.dataDir}
			dataDirs = append(dataDirs, encoder.EncodeStrings(dqtm.tags)...)
			dataDirs = append(dataDirs, "0_processing.idx")

			statFile, err := os.ReadFile(filepath.Join(dataDirs...))
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(statFile).To(Equal([]byte(`P1.D1.`))) // P1 (procsessing index 1) D1 (delete index 1)
		})

		t.Run("it records that the item was properly processed", func(t *testing.T) {
			dqtm := SetupDiskQueue(g, nil)
			defer dqtm.Clean()

			g.Expect(dqtm.dq.Enqueue([]byte(`hello world`))).ToNot(HaveOccurred())
			dequeueMessage := dqtm.GetMessage(g)

			g.Expect(dqtm.dq.ACK(dequeueMessage.ID, true)).ToNot(HaveOccurred())

			metrics := dqtm.dq.Metrics()
			g.Expect(metrics.Ready).To(Equal(uint64(0)))
			g.Expect(metrics.Processing).To(Equal(uint64(0)))
		})

		t.Run("next enqueue is set to index 1 again", func(t *testing.T) {
			dqtm := SetupDiskQueue(g, nil)
			defer dqtm.Clean()

			g.Expect(dqtm.dq.Enqueue([]byte(`hello world 1`))).ToNot(HaveOccurred())
			dequeueMessage := dqtm.GetMessage(g)
			g.Expect(dequeueMessage.ID).To(Equal(uint64(1)))
			g.Expect(dqtm.dq.ACK(dequeueMessage.ID, true)).ToNot(HaveOccurred())

			g.Expect(dqtm.dq.Enqueue([]byte(`hello world`))).ToNot(HaveOccurred())
			dequeueMessage = dqtm.GetMessage(g)
			g.Expect(dequeueMessage.ID).To(Equal(uint64(1)))
		})
	})

	t.Run("when passed it FALSE", func(t *testing.T) {
		t.Run("it requeues the processing item and records that it has failed", func(t *testing.T) {
			dqtm := SetupDiskQueue(g, nil)
			defer dqtm.Clean()

			g.Expect(dqtm.dq.Enqueue([]byte(`hello world`))).ToNot(HaveOccurred())
			dequeueMessage := dqtm.GetMessage(g)

			g.Expect(dqtm.dq.ACK(dequeueMessage.ID, false)).ToNot(HaveOccurred())

			// queue item state file
			dataDirs := []string{dqtm.dataDir}
			dataDirs = append(dataDirs, encoder.EncodeStrings(dqtm.tags)...)
			dataDirs = append(dataDirs, "0_processing.idx")

			statFile, err := os.ReadFile(filepath.Join(dataDirs...))
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(statFile).To(Equal([]byte(`P1.R1.`))) // P1 (procsessing index 1) D1 (delete index 1)
		})

		t.Run("it records that the item was properly processed", func(t *testing.T) {
			dqtm := SetupDiskQueue(g, nil)
			defer dqtm.Clean()

			g.Expect(dqtm.dq.Enqueue([]byte(`hello world`))).ToNot(HaveOccurred())
			dequeueMessage := dqtm.GetMessage(g)

			g.Expect(dqtm.dq.ACK(dequeueMessage.ID, false)).ToNot(HaveOccurred())

			metrics := dqtm.dq.Metrics()
			g.Expect(metrics.Ready).To(Equal(uint64(1)))
			g.Expect(metrics.Processing).To(Equal(uint64(0)))
		})

		t.Run("it is requeued with proper values", func(t *testing.T) {
			dqtm := SetupDiskQueue(g, nil)
			defer dqtm.Clean()

			g.Expect(dqtm.dq.Enqueue([]byte(`hello world`))).ToNot(HaveOccurred())
			dequeueMessage := dqtm.GetMessage(g)
			g.Expect(dqtm.dq.ACK(dequeueMessage.ID, false)).ToNot(HaveOccurred())

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

			g.Expect(location.RetryCount).To(Equal(uint64(1)))

			dequeueMessage, err := location.Process()
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(dequeueMessage.Data).To(Equal([]byte(`hello world`)))
		})

		t.Run("it is sent to the dead letter queue after retry count failure is reached", func(t *testing.T) {
			dqtm := SetupDiskQueue(g, nil)
			defer dqtm.Clean()

			g.Expect(dqtm.dq.Enqueue([]byte(`hello world`))).ToNot(HaveOccurred())

			// deafult retries is 2. so pull and fail 3 times
			dequeueMessage := dqtm.GetMessage(g)
			g.Expect(dqtm.dq.ACK(dequeueMessage.ID, false)).ToNot(HaveOccurred())

			dequeueMessage = dqtm.GetMessage(g)
			g.Expect(dqtm.dq.ACK(dequeueMessage.ID, false)).ToNot(HaveOccurred())

			dequeueMessage = dqtm.GetMessage(g)
			g.Expect(dqtm.dq.ACK(dequeueMessage.ID, false)).ToNot(HaveOccurred())

			// metrics should record the item has been removed
			metrics := dqtm.dq.Metrics()
			g.Expect(metrics.Ready).To(Equal(uint64(0)))
			g.Expect(metrics.Processing).To(Equal(uint64(0)))

			// check dead letter queue
			g.Expect(dqtm.deadLetterQueue.Count()).To(Equal(uint64(1)))
			deadMessage, err := dqtm.deadLetterQueue.Get(0)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(deadMessage.ID).To(Equal(uint64(0)))
			g.Expect(deadMessage.Data).To(Equal([]byte(`hello world`)))
		})
	})
}
