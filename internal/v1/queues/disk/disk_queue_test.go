package disk

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/DanLavine/willow/internal/v1/models"
	"github.com/DanLavine/willow/internal/v1/queues/disk/encoder"
	. "github.com/onsi/gomega"
)

func TestQueue_NewDiskQueue(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("returns an error if the readers is empty", func(t *testing.T) {
		testDir, err := os.MkdirTemp("", "")
		g.Expect(err).ToNot(HaveOccurred())
		defer os.RemoveAll(testDir)

		dq, dqErr := NewDiskQueue(testDir, []string{}, nil)
		g.Expect(dqErr).To(HaveOccurred())
		g.Expect(dqErr.Error()).To(ContainSubstring("No readers received"))
		g.Expect(dq).To(BeNil())
	})

	t.Run("returns an error if any readers are nil", func(t *testing.T) {
		testDir, err := os.MkdirTemp("", "")
		g.Expect(err).ToNot(HaveOccurred())
		defer os.RemoveAll(testDir)

		dq, dqErr := NewDiskQueue(testDir, []string{}, []chan *models.Location{make(chan *models.Location), nil})
		g.Expect(dqErr).To(HaveOccurred())
		g.Expect(dqErr.Error()).To(ContainSubstring("Null reader received"))
		g.Expect(dq).To(BeNil())
	})

	t.Run("creates all inital encoder files at their proper location", func(t *testing.T) {
		testDir, err := os.MkdirTemp("", "")
		g.Expect(err).ToNot(HaveOccurred())
		reader := []chan *models.Location{make(chan *models.Location)}

		dq, dqErr := NewDiskQueue(testDir, []string{"tags1"}, reader)
		g.Expect(dqErr).ToNot(HaveOccurred())
		g.Expect(dq).ToNot(BeNil())

		defer func() {
			dq.Close()
			close(reader[0])
			os.RemoveAll(testDir)
		}()

		// queue item file
		fileInfo, err := os.Stat(filepath.Join(testDir, encoder.EncodeStrings([]string{"tags1"}), "0.idx"))
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(fileInfo.IsDir()).To(BeFalse())

		// queue item state file
		fileInfo, err = os.Stat(filepath.Join(testDir, encoder.EncodeStrings([]string{"tags1"}), "0_processing.idx"))
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(fileInfo.IsDir()).To(BeFalse())

		// queue item update file
		fileInfo, err = os.Stat(filepath.Join(testDir, encoder.EncodeStrings([]string{"tags1"}), "update.idx"))
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(fileInfo.IsDir()).To(BeFalse())
	})

	t.Run("creates a queue with empty tags", func(t *testing.T) {
		testDir, err := os.MkdirTemp("", "")
		g.Expect(err).ToNot(HaveOccurred())
		reader := []chan *models.Location{make(chan *models.Location)}

		dq, dqErr := NewDiskQueue(testDir, nil, reader)
		g.Expect(dqErr).ToNot(HaveOccurred())
		g.Expect(dq).ToNot(BeNil())

		g.Eventually(dq.Drain()).Should(BeClosed())
		close(reader[0])
		os.RemoveAll(testDir)
	})

	t.Run("creates a queue with no tags", func(t *testing.T) {
		testDir, err := os.MkdirTemp("", "")
		g.Expect(err).ToNot(HaveOccurred())
		reader := []chan *models.Location{make(chan *models.Location)}

		dq, dqErr := NewDiskQueue(testDir, []string{}, reader)
		g.Expect(dqErr).ToNot(HaveOccurred())
		g.Expect(dq).ToNot(BeNil())

		g.Eventually(dq.Drain()).Should(BeClosed())
		close(reader[0])
		os.RemoveAll(testDir)
	})

	t.Run("creates a queue with multiple tags", func(t *testing.T) {
		testDir, err := os.MkdirTemp("", "")
		g.Expect(err).ToNot(HaveOccurred())
		reader := []chan *models.Location{make(chan *models.Location)}

		dq, dqErr := NewDiskQueue(testDir, []string{"tag1", "tag2"}, reader)
		g.Expect(dqErr).ToNot(HaveOccurred())
		g.Expect(dq).ToNot(BeNil())

		defer func() {
			g.Eventually(dq.Drain()).Should(BeClosed())
			close(reader[0])
			os.RemoveAll(testDir)
		}()

		// queue item file
		fileInfo, err := os.Stat(filepath.Join(testDir, encoder.EncodeStrings([]string{"tag1", "tag2"}), "0.idx"))
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(fileInfo.IsDir()).To(BeFalse())
	})
}

func TestQueue_Enqueueue(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("properly updates the ready count", func(t *testing.T) {
		testDir, err := os.MkdirTemp("", "")
		g.Expect(err).ToNot(HaveOccurred())
		reader := []chan *models.Location{make(chan *models.Location)}

		dq, dqErr := NewDiskQueue(testDir, []string{"tag1", "tag2"}, reader)
		g.Expect(dqErr).ToNot(HaveOccurred())
		g.Expect(dq).ToNot(BeNil())

		defer func() {
			dq.Close()
			close(reader[0])
			os.RemoveAll(testDir)
		}()

		g.Expect(dq.Enqueue([]byte("first"))).ToNot(HaveOccurred())
		metrics := dq.Metrics()
		g.Expect(metrics.Ready).To(Equal(uint64(1)))
		g.Expect(metrics.Processing).To(Equal(uint64(0)))

		g.Expect(dq.Enqueue([]byte("second"))).ToNot(HaveOccurred())
		metrics = dq.Metrics()
		g.Expect(metrics.Ready).To(Equal(uint64(2)))
		g.Expect(metrics.Processing).To(Equal(uint64(0)))
	})

	t.Run("allows the enqueued item to be returned on a read channel", func(t *testing.T) {
		testDir, err := os.MkdirTemp("", "")
		g.Expect(err).ToNot(HaveOccurred())
		reader := make(chan *models.Location)
		readers := []chan *models.Location{reader}

		dq, dqErr := NewDiskQueue(testDir, []string{"tag1", "tag2"}, readers)
		g.Expect(dqErr).ToNot(HaveOccurred())
		g.Expect(dq).ToNot(BeNil())

		defer func() {
			dq.Close()
			close(reader)
			os.RemoveAll(testDir)
		}()

		g.Expect(dq.Enqueue([]byte("first"))).ToNot(HaveOccurred())

		context, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second))
		defer cancel()

		var location *models.Location
		g.Eventually(func() *models.Location {
			select {
			case <-context.Done():
			case location = <-reader:
			}

			return location
		}, time.Second, 100*time.Millisecond, context).ShouldNot(BeNil())
	})

	t.Run("returns enqueued items in the order they were added", func(t *testing.T) {
		testDir, err := os.MkdirTemp("", "")
		g.Expect(err).ToNot(HaveOccurred())
		reader := make(chan *models.Location)
		readers := []chan *models.Location{reader}

		dq, dqErr := NewDiskQueue(testDir, []string{}, readers)
		g.Expect(dqErr).ToNot(HaveOccurred())
		g.Expect(dq).ToNot(BeNil())

		defer func() {
			dq.Close()
			close(reader)
			os.RemoveAll(testDir)
		}()

		g.Expect(dq.Enqueue([]byte("first"))).ToNot(HaveOccurred())
		g.Expect(dq.Enqueue([]byte("second"))).ToNot(HaveOccurred())

		var location *models.Location
		cdl, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second))
		g.Eventually(func() *models.Location {
			select {
			case <-cdl.Done():
			case location = <-reader:
			}

			return location
		}, time.Second, 100*time.Millisecond, cdl).ShouldNot(BeNil())
		cancel()

		dequeueMessage, locationErr := location.Process()
		g.Expect(locationErr).ToNot(HaveOccurred())
		g.Expect(dequeueMessage.Data).To(Equal([]byte(`first`)))

		cdl, cancel = context.WithDeadline(context.Background(), time.Now().Add(time.Second))
		g.Eventually(func() *models.Location {
			select {
			case <-cdl.Done():
			case location = <-reader:
			}

			return location
		}, time.Second, 100*time.Millisecond, cdl).ShouldNot(BeNil())
		cancel()

		dequeueMessage, locationErr = location.Process()
		g.Expect(locationErr).ToNot(HaveOccurred())
		g.Expect(dequeueMessage.Data).To(Equal([]byte(`second`)))
	})

	t.Run("updates the metric counts properly", func(t *testing.T) {
		testDir, err := os.MkdirTemp("", "")
		g.Expect(err).ToNot(HaveOccurred())
		reader := make(chan *models.Location)
		readers := []chan *models.Location{reader}

		dq, dqErr := NewDiskQueue(testDir, []string{"tag1", "tag2"}, readers)
		g.Expect(dqErr).ToNot(HaveOccurred())
		g.Expect(dq).ToNot(BeNil())

		defer func() {
			dq.Close()
			close(reader)
			os.RemoveAll(testDir)
		}()

		g.Expect(dq.Enqueue([]byte("first"))).ToNot(HaveOccurred())

		context, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second))
		defer cancel()

		var location *models.Location
		g.Eventually(func() *models.Location {
			select {
			case <-context.Done():
			case location = <-reader:
			}

			return location
		}, time.Second, 100*time.Millisecond, context).ShouldNot(BeNil())

		_, locationErr := location.Process()
		g.Expect(locationErr).ToNot(HaveOccurred())

		metrics := dq.Metrics()
		g.Expect(metrics.Ready).To(Equal(uint64(0)))
		g.Expect(metrics.Processing).To(Equal(uint64(1)))
	})
}
