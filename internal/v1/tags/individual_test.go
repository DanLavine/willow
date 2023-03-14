package tags_test

import (
	"testing"

	"github.com/DanLavine/willow/internal/v1/tags"
	. "github.com/onsi/gomega"
)

func TestIndividual_GetGlobalReader(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("returns a global reader even if no tags have been  defined yet", func(t *testing.T) {
		reader := tags.NewReaderTree()
		g.Expect(reader).ToNot(BeNil())
		g.Expect(reader.GetGlobalReader()).ToNot(BeNil())
	})
}

func TestIndividual_CreateTagsGroup(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("Creates all tag combinations and assigns one channel for all of them", func(t *testing.T) {
		reader := tags.NewReaderTree()
		channels := reader.CreateTagsGroup([]string{"a", "b", "c", "d", "e"})

		g.Expect(len(channels)).To(Equal(2))
		g.Expect(channels[0]).ToNot(BeNil())
		g.Expect(channels[1]).ToNot(BeNil())
	})

	t.Run("when creating all the same tags, no new channels are created", func(t *testing.T) {
		reader := tags.NewReaderTree()

		channels := reader.CreateTagsGroup([]string{"a", "b", "c", "d", "e"})
		g.Expect(len(channels)).To(Equal(2))
		g.Expect(channels[0]).ToNot(BeNil())
		g.Expect(channels[1]).ToNot(BeNil())

		channels2 := reader.CreateTagsGroup([]string{"a", "b", "c", "d"})
		g.Expect(len(channels2)).To(Equal(2))

		for _, channel := range channels {
			g.Expect(channels2).To(ContainElement(channel))
		}
	})

	t.Run("when creating new tags a new channel is used in addition to the common channel for exists pairs", func(t *testing.T) {
		reader := tags.NewReaderTree()

		channels := reader.CreateTagsGroup([]string{"a", "b", "c", "d", "e"})
		g.Expect(len(channels)).To(Equal(2))
		g.Expect(channels[0]).ToNot(BeNil())
		g.Expect(channels[1]).ToNot(BeNil())

		channels2 := reader.CreateTagsGroup([]string{"a", "b", "c", "d", "f"})
		g.Expect(len(channels2)).To(Equal(3))

		for _, channel := range channels {
			g.Expect(channels2).To(ContainElement(channel))
		}
	})
}

func TestIndividual_GetReader(t *testing.T) {
	g := NewGomegaWithT(t)

	reader := tags.NewReaderTree()
	channels := reader.CreateTagsGroup([]string{"a", "b", "c", "d", "e"})

	t.Run("Getting a tag that already exists returns the proper channel", func(t *testing.T) {
		channel := reader.GetReader("abc")
		g.Expect(channel).ToNot(BeNil())
		g.Expect(channels).To(ContainElement(channel))
	})

	t.Run("Getting a new tag creates a new channel", func(t *testing.T) {
		channel := reader.GetReader("big boom")
		g.Expect(channel).ToNot(BeNil())
		g.Expect(channels).ToNot(ContainElement(channel))
	})
}

func TestIndividual_GetReaders(t *testing.T) {
	g := NewGomegaWithT(t)

	reader := tags.NewReaderTree()
	channels := reader.CreateTagsGroup([]string{"a", "b", "c", "d", "e"})

	t.Run("Getting tags that already exists returns the proper channel", func(t *testing.T) {
		chans := reader.GetReaders([]string{"abc", "e"})
		g.Expect(len(chans)).To(Equal(1))
		g.Expect(channels).To(ContainElement(chans[0]))
	})

	t.Run("Getting a new tag creates a new channel", func(t *testing.T) {
		chans := reader.GetReaders([]string{"gtha", "f"})
		g.Expect(len(chans)).To(Equal(2))

		for _, channel := range chans {
			g.Expect(channels).ToNot(ContainElement(channel))
		}
	})
}
