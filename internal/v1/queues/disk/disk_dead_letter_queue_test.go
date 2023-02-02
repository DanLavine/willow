package disk_test

import (
	"os"
	"testing"

	"github.com/DanLavine/willow/internal/config"
	"github.com/DanLavine/willow/internal/errors"
	"github.com/DanLavine/willow/internal/v1/queues/disk"
	v1 "github.com/DanLavine/willow/pkg/models/v1"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
)

func setupDeadLetterQueueParams(dir string, g *gomega.WithT) *disk.DiskDeadLetterQueue {
	configQueue := config.ConfigQueue{DeadLetterQueueMaxEntries: 1, ConfigDisk: &config.ConfigDisk{StorageDir: testDir}}
	g.Expect(configQueue.Validate()).ToNot(HaveOccurred())

	createParams := v1.Create{BrokerType: v1.Queue, BrokerTags: []string{"test"}, DeadLetterQueueParams: &v1.DeadLetterQueueParams{MaxSize: 1}}
	g.Expect(createParams.Validate()).ToNot(HaveOccurred())

	diskDeadLetterQueue, err := disk.DiskDeadLetterQueue(configQueue, createParams)
	g.Expect(err).ToNot(HaveOccurred())

	return diskDeadLetterQueue
}

func TestDeadLetterQueue_CanPerformBasicQueueOperations(t *testing.T) {
	g := NewGomegaWithT(t)

	testDir, dirErr = os.MkdirTemp("", "")
	g.Expect(dirErr).ToNot(HaveOccurred())
	defer os.RemoveAll(testDiir)

	diskDeadLetterQueue := setupDeadLetterQueueParams(testDir, g)
	defer diskDeadLetterQueue.Close()

	// check count and set data
	g.Expect(diskDeadLetterQueue.Count()).To(Equal(uint64(0)))
	g.Expect(diskDeadLetterQueue.Enqueue([]byte(`Zmlyc3Q=`))).ToNot(HaveOccurred())

	// check count
	g.Expect(diskDeadLetterQueue.Count()).To(Equal(uint64(1)))

	// read data
	dequeueMessage, err := diskDeadLetterQueue.Get(0, []string{"test"})
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(dequeueMessage.Data).To(Equal([]byte(`first`)))

	// clear data
	g.Expect(diskDeadLetterQueue.Clear()).ToNot(HaveOccurred())

	// check data one last time
	g.Expect(diskDeadLetterQueue.Count()).To(Equal(uint64(0)))
}

func TestDeadLetterQueue_Enqueue_ErrorsIfTheQueueIsFull(t *testing.T) {
	g := NewGomegaWithT(t)

	testDir, dirErr = os.MkdirTemp("", "")
	g.Expect(dirErr).ToNot(HaveOccurred())
	defer os.RemoveAll(testDiir)

	diskDeadLetterQueue := setupDeadLetterQueueParams(testDir, g)
	defer diskDeadLetterQueue.Close()

	err = diskDeadLetterQueue.Enqueue([]byte(`Zmlyc3Q=`))
	g.Expect(err).ToNot(HaveOccurred())

	err = diskDeadLetterQueue.Enqueue([]byte(`c2Vjb25k`))
	g.Expect(err).To(HaveOccurred())
	g.Expect(err).To(Equal(errors.DeadLetterQueueFull))
}

func TestDeadLetterQueue_Get_ReturnsAnErrorOnABadIndex(t *testing.T) {
	g := NewGomegaWithT(t)

	testDir, dirErr = os.MkdirTemp("", "")
	g.Expect(dirErr).ToNot(HaveOccurred())
	defer os.RemoveAll(testDiir)

	diskDeadLetterQueue := setupDeadLetterQueueParams(testDir, g)
	defer diskDeadLetterQueue.Close()

	_, err = diskDeadLetterQueue.Get(3, []string{"test"})
	g.Expect(err).To(HaveOccurred())
	g.Expect(err).To(Equal(errors.DeadLetterItemNotfound))
}
