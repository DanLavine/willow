package tags

import (
	"testing"
	"time"

	"github.com/DanLavine/willow/pkg/models/datatypes"
	v1 "github.com/DanLavine/willow/pkg/models/v1"
	. "github.com/onsi/gomega"
)

func numberOfDifferences(g *GomegaWithT, diffCount int, expected, compare []chan<- Tag) {
	for _, expectedVal := range expected {
		for index, compareVal := range compare {
			if expectedVal == compareVal {
				break
			}

			if index == len(compareVal)-1 {
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

	t.Run("Creates all tag combinations and assigns one reader per combination", func(t *testing.T) {
		reader := NewTagReaderTree()
		channels := reader.CreateGroup(datatypes.Strings{"a", "b", "c", "d", "e"})

		// Note 3 here for (global, strict, and 31 channels for all tag combinations)
		g.Expect(len(channels)).To(Equal(33))
		for index, channel := range channels {
			g.Expect(channel).ToNot(BeNil(), index)
		}
	})

	t.Run("when creating subset of the same tags, a new 'strict' channel is created", func(t *testing.T) {
		reader := NewTagReaderTree()

		channels := reader.CreateGroup(datatypes.Strings{"a", "b", "c", "d", "e"})
		g.Expect(len(channels)).To(Equal(33))

		channels2 := reader.CreateGroup(datatypes.Strings{"a", "b", "c", "d"})
		g.Expect(len(channels2)).To(Equal(17))

		// global + general are the same
		// different strict readers
		numberOfDifferences(g, 1, channels2, channels)
	})

	t.Run("when creating new tags a new channel is used in addition to the common channel for exists pairs", func(t *testing.T) {
		reader := NewTagReaderTree()

		channels := reader.CreateGroup(datatypes.Strings{"a", "b", "c", "d", "e"})
		g.Expect(len(channels)).To(Equal(33))

		channels2 := reader.CreateGroup(datatypes.Strings{"a", "b", "c", "d", "f"})
		g.Expect(len(channels2)).To(Equal(33))

		// global + 1 general are the same
		// different strict readers + 1 general reader for the 'f' tag
		numberOfDifferences(g, 8, channels2, channels)
	})
}

func TestTagReader_GetStrictReader(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("Getting a tag group that already exists returns the proper strict reader", func(t *testing.T) {
		reader := NewTagReaderTree()
		channels := reader.CreateGroup(datatypes.Strings{"a", "b", "c", "d", "e"})
		g.Expect(len(channels)).To(Equal(33))

		strictChan := reader.GetStrictReader(datatypes.Strings{"a", "b", "c", "d", "e"})

		// ensure the strict chan is a proper reader from the CreateGroup writers
		for index, channel := range channels {
			go func(i int, ch chan<- Tag) {
				select {
				case ch <- func() *v1.DequeueItemResponse { return &v1.DequeueItemResponse{ID: 1} }:
					// trigger the message
				case <-time.Tick(time.Second):
					//nothing to do here
				}
			}(index, channel)
		}

		g.Eventually(strictChan).Should(Receive())
	})

	t.Run("Getting a tag group that does not exists returns a new strict reader", func(t *testing.T) {
		reader := NewTagReaderTree()
		strictChan := reader.GetStrictReader(datatypes.Strings{"a", "b", "c", "d", "e"})
		g.Expect(strictChan).ToNot(BeNil())
	})
}

func TestTagReader_GetSubsetReader(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("Getting a tag group that already exists returns the proper subset reader", func(t *testing.T) {
		reader := NewTagReaderTree()
		channels := reader.CreateGroup(datatypes.Strings{"a", "b", "c", "d", "e"})
		g.Expect(len(channels)).To(Equal(33))

		subsetChan := reader.GetSubsetReader(datatypes.Strings{"a", "b", "c", "d"})

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

		g.Eventually(subsetChan).Should(Receive())
	})

	t.Run("Getting a tag group that does not exists returns a new subset reader", func(t *testing.T) {
		reader := NewTagReaderTree()
		subsetChan := reader.GetSubsetReader(datatypes.Strings{"a", "b", "c", "d"})
		g.Expect(subsetChan).ToNot(BeNil())
	})
}

func TestTagReader_GetAnyReaders(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("Getting a reader for each tag", func(t *testing.T) {
		reader := NewTagReaderTree()
		channels := reader.CreateGroup(datatypes.Strings{"a", "b", "c", "d", "e"})
		g.Expect(len(channels)).To(Equal(33))

		anyChan := reader.GetAnyReaders(datatypes.Strings{"a", "b", "c", "d", "e"})
		g.Expect(len(anyChan)).To(Equal(5))

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

		for _, channel := range anyChan {
			g.Eventually(channel).Should(Receive())
		}
	})

	t.Run("Getting all readers returns multiple readers if they were made on differnt create requests", func(t *testing.T) {
		reader := NewTagReaderTree()
		channels := reader.CreateGroup(datatypes.Strings{"a", "b", "c"})
		g.Expect(len(channels)).To(Equal(9))

		channels = reader.CreateGroup(datatypes.Strings{"d", "e"})
		g.Expect(len(channels)).To(Equal(5))

		anyChan := reader.GetAnyReaders(datatypes.Strings{"a", "e"})
		g.Expect(len(anyChan)).To(Equal(2))
	})

	t.Run("Getting a number of tags that does not exists returns a new reader", func(t *testing.T) {
		reader := NewTagReaderTree()
		anyChans := reader.GetAnyReaders(datatypes.Strings{"a", "b", "c", "d"})
		g.Expect(len(anyChans)).To(Equal(4))
	})
}
