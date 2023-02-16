package disk_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/DanLavine/willow/internal/v1/queues/disk"
	v1 "github.com/DanLavine/willow/pkg/models/v1"
	. "github.com/onsi/gomega"
)

func setupTestDir(g *WithT, f func(tmpDir string)) {
	testDir, dirErr := os.MkdirTemp("", "")
	g.Expect(dirErr).ToNot(HaveOccurred())
	defer os.RemoveAll(testDir)

	f(testDir)
}

func setupWithQueue(g *WithT, f func(queue *disk.Queue)) {
	testDir, dirErr := os.MkdirTemp("", "")
	g.Expect(dirErr).ToNot(HaveOccurred())
	defer os.RemoveAll(testDir)

	create := &v1.Create{
		Name:         "test",
		QueueMaxSize: 5,
		RetryCount:   1,
		//DeadLetterQueueMaxSize
	}

	queue, diskErr := disk.NewQueue(testDir, create)
	g.Expect(diskErr).ToNot(HaveOccurred())
	defer queue.Stop()

	f(queue)
}

func TestDiskQueue_NewQueue(t *testing.T) {
	g := NewGomegaWithT(t)

	create := &v1.Create{
		Name: "test",
	}

	t.Run("it save the details about the queue", func(t *testing.T) {
		setupTestDir(g, func(testDir string) {
			queue, diskErr := disk.NewQueue(testDir, create)
			g.Expect(diskErr).ToNot(HaveOccurred())
			defer queue.Stop()

			fileData, err := os.ReadFile(filepath.Join(testDir, disk.EncodeString("test"), "queue.info"))
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(fileData).To(Equal([]byte(`{"Name":"test","QueueMaxSize":0,"RetryCount":0,"DeadLetterQueueMaxSize":null}`)))
		})
	})
}

func TestDiskQueue_LoadQeueue(t *testing.T) {
	g := NewGomegaWithT(t)

	create := &v1.Create{
		Name:         "test",
		QueueMaxSize: 5,
	}

	t.Run("it loads data from disk if the queue info exists", func(t *testing.T) {
		setupTestDir(g, func(testDir string) {
			// start and stop first queue to generate info
			queue, diskErr := disk.NewQueue(testDir, create)
			g.Expect(diskErr).ToNot(HaveOccurred())
			queue.Stop()

			// load second queue to read info from disk
			queue, diskErr = disk.LoadQueue(testDir, "test")
			g.Expect(diskErr).ToNot(HaveOccurred())
			defer queue.Stop()

			// check metrics to know things are proper
			metrics := queue.Metrics()
			g.Expect(metrics).ToNot(BeNil())
			g.Expect(metrics.Max).To(Equal(uint64(5))) // default in test setup
		})
	})

	t.Run("it returns an error if there is no queue info on disk", func(t *testing.T) {
		setupTestDir(g, func(testDir string) {
			queue, diskErr := disk.LoadQueue(testDir, "test")
			g.Expect(diskErr).To(HaveOccurred())
			g.Expect(queue).To(BeNil())
		})
	})
}

func TestDiskQueue_Enque(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("it returns", func(t *testing.T) {
		setupWithQueue(g, func(queue *disk.Queue) {
		})
	})
}
