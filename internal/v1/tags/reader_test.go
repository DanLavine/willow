package tags

import (
	"testing"
	"time"

	v1 "github.com/DanLavine/willow/pkg/models/v1"
	. "github.com/onsi/gomega"
)

func numberOfDifferences(g *GomegaWithT, diffCount int, first, second []chan<- Tag) {
	for _, firstVal := range first {
		for index, secondVal := range second {
			if firstVal == secondVal {
				break
			}

			if index == len(second)-1 {
				diffCount--
			}
		}
	}

	g.Expect(diffCount).To(BeNumerically(">=", 0))
}

func TestTagReaders_GetGlobalReader(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("returns a global reader even if no tags have been  defined yet", func(t *testing.T) {
		readerTree := NewTagReaderTree()
		g.Expect(readerTree).ToNot(BeNil())
		g.Expect(readerTree.GetGlobalReader()).ToNot(BeNil())
	})
}

func TestTagReaders_CreateGroup(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("Creates all tag combinations and assigns one channel for all of them", func(t *testing.T) {
		reader := NewTagReaderTree()
		channels := reader.CreateGroup([]string{"a", "b", "c", "d", "e"})

		// Note 3 here for (global, strict, new reader for all tag combos)
		g.Expect(len(channels)).To(Equal(3))
		g.Expect(channels[0]).ToNot(BeNil())
		g.Expect(channels[1]).ToNot(BeNil())
		g.Expect(channels[2]).ToNot(BeNil())
	})

	t.Run("when creating subseet of the same tags, a new 'strict' channel is created", func(t *testing.T) {
		reader := NewTagReaderTree()

		channels := reader.CreateGroup([]string{"a", "b", "c", "d", "e"})
		g.Expect(len(channels)).To(Equal(3))

		channels2 := reader.CreateGroup([]string{"a", "b", "c", "d"})
		g.Expect(len(channels2)).To(Equal(3))

		// global + general are the same
		// different strict readers
		numberOfDifferences(g, 1, channels, channels2)
	})

	t.Run("when creating new tags a new channel is used in addition to the common channel for exists pairs", func(t *testing.T) {
		reader := NewTagReaderTree()

		channels := reader.CreateGroup([]string{"a", "b", "c", "d", "e"})
		g.Expect(len(channels)).To(Equal(3))

		channels2 := reader.CreateGroup([]string{"a", "b", "c", "d", "f"})
		g.Expect(len(channels2)).To(Equal(4))

		// global + 1 general are the same
		// different strict readers + 1 general reader for the 'f' tag
		numberOfDifferences(g, 2, channels2, channels)
	})
}

func TestTagReader_GetStrictReader(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("Getting a tag group that already exists returns the proper strict reader", func(t *testing.T) {
		reader := NewTagReaderTree()
		channels := reader.CreateGroup([]string{"a", "b", "c", "d", "e"})
		g.Expect(len(channels)).To(Equal(3))

		strictChan := reader.GetStrictReader([]string{"a", "b", "c", "d", "e"})

		// ensure the strict chan is a proper reader from the CreateGroup writers
		for index, channel := range channels {
			go func(i int, ch chan<- Tag) {
				select {
				case ch <- func() *v1.DequeueItemResponse { return &v1.DequeueItemResponse{ID: uint64(i + 1)} }:
					// trigger the message
				case <-time.Tick(time.Second):
					//nothing to do here
				}
			}(index, channel)
		}

		readValue := <-strictChan
		dequeueMessage := readValue()

		g.Expect(dequeueMessage.ID).To(BeNumerically(">=", 1))
		g.Expect(dequeueMessage.ID).To(BeNumerically("<=", 3))
	})

	t.Run("Getting a tag group that does not exists returns a new strict reader", func(t *testing.T) {
		reader := NewTagReaderTree()
		strictChan := reader.GetStrictReader([]string{"a", "b", "c", "d", "e"})
		g.Expect(strictChan).ToNot(BeNil())
	})
}

func TestTagReader_GetSubsetReader(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("Getting a tag group that already exists returns the proper subset reader", func(t *testing.T) {
		reader := NewTagReaderTree()
		channels := reader.CreateGroup([]string{"a", "b", "c", "d", "e"})
		g.Expect(len(channels)).To(Equal(3))

		subsetChan := reader.GetSubsetReader([]string{"a", "b", "c", "d"})

		// ensure the subset chan is a proper reader from the CreateGroup writers
		for index, channel := range channels {
			go func(i int, ch chan<- Tag) {
				select {
				case ch <- func() *v1.DequeueItemResponse { return &v1.DequeueItemResponse{ID: uint64(i + 1)} }:
					// trigger the message
				case <-time.Tick(time.Second):
					//nothing to do here
				}
			}(index, channel)
		}

		readValue := <-subsetChan
		dequeueMessage := readValue()

		g.Expect(dequeueMessage.ID).To(BeNumerically(">=", 1))
		g.Expect(dequeueMessage.ID).To(BeNumerically("<=", 3))
	})

	t.Run("Getting a tag group that does not exists returns a new subset reader", func(t *testing.T) {
		reader := NewTagReaderTree()
		strictChan := reader.GetSubsetReader([]string{"a", "b", "c", "d"})
		g.Expect(strictChan).ToNot(BeNil())
	})
}

func TestTagReader_GetAnyReaders(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("Getting all readers for each tag", func(t *testing.T) {
		reader := NewTagReaderTree()
		channels := reader.CreateGroup([]string{"a", "b", "c", "d", "e"})
		g.Expect(len(channels)).To(Equal(3))

		anyChan := reader.GetAnyReaders([]string{"a", "b", "c", "d", "e"})
		g.Expect(len(anyChan)).To(Equal(1))

		// ensure the subset chan is a proper reader from the CreateGroup writers
		for index, channel := range channels {
			go func(i int, ch chan<- Tag) {
				select {
				case ch <- func() *v1.DequeueItemResponse { return &v1.DequeueItemResponse{ID: uint64(i + 1)} }:
					// trigger the message
				case <-time.Tick(time.Second):
					//nothing to do here
				}
			}(index, channel)
		}

		readValue := <-anyChan[0]
		dequeueMessage := readValue()

		g.Expect(dequeueMessage.ID).To(BeNumerically(">=", 1))
		g.Expect(dequeueMessage.ID).To(BeNumerically("<=", 3))
	})

	t.Run("Getting all readers returns multiple readers if they were made on differnt create requests", func(t *testing.T) {
		reader := NewTagReaderTree()
		channels := reader.CreateGroup([]string{"a", "b", "c"})
		g.Expect(len(channels)).To(Equal(3))

		channels = reader.CreateGroup([]string{"d", "e"})
		g.Expect(len(channels)).To(Equal(3))

		anyChan := reader.GetAnyReaders([]string{"a", "e"})
		g.Expect(len(anyChan)).To(Equal(2))
	})

	t.Run("Getting a number of tags that does not exists returns a new reader", func(t *testing.T) {
		reader := NewTagReaderTree()
		anyChans := reader.GetAnyReaders([]string{"a", "b", "c", "d"})
		g.Expect(len(anyChans)).To(Equal(1))
	})
}
