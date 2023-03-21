package tags

import (
	"testing"

	"github.com/DanLavine/willow/internal/datastructures"
	"github.com/DanLavine/willow/internal/datastructures/datastructuresfakes"
	. "github.com/onsi/gomega"
)

func callback() datastructures.TreeItem {
	return &datastructuresfakes.FakeTreeItem{}
}

func TestTagsGoup_FindOrCreateTagsGroup(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("returns an error if the tags are nil", func(t *testing.T) {
		tagsGroup := NewTagsGroup()

		val, err := tagsGroup.FindOrCreateTagsGroup(nil, callback)
		g.Expect(val).To(BeNil())
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("received an invalid tags length"))
	})

	t.Run("returns an error if the tags are empty", func(t *testing.T) {
		tagsGroup := NewTagsGroup()

		val, err := tagsGroup.FindOrCreateTagsGroup([]string{}, callback)
		g.Expect(val).To(BeNil())
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("received an invalid tags length"))
	})

	t.Run("returns the passed in value, if the tag group has not yet been created", func(t *testing.T) {
		tagsGroup := NewTagsGroup()

		val, err := tagsGroup.FindOrCreateTagsGroup([]string{"a"}, callback)
		g.Expect(val).ToNot(BeNil())
		g.Expect(err).ToNot(HaveOccurred())
	})

	t.Run("returns the original value, if the tag group has already been created", func(t *testing.T) {
		tagsGroup := NewTagsGroup()

		val, err := tagsGroup.FindOrCreateTagsGroup([]string{"a"}, callback)
		g.Expect(err).ToNot(HaveOccurred())

		val2, err := tagsGroup.FindOrCreateTagsGroup([]string{"a"}, callback)
		g.Expect(val2).To(Equal(val))
		g.Expect(err).ToNot(HaveOccurred())
	})

	t.Run("creates a proper tree item when using multiple tags", func(t *testing.T) {
		tagsGroupTree := NewTagsGroup()
		slice := []string{"a", "b", "c", "d", "e"}

		val, err := tagsGroupTree.FindOrCreateTagsGroup(slice, callback)
		g.Expect(val).ToNot(BeNil())
		g.Expect(err).ToNot(HaveOccurred())

		var privateTagsGroup *tagsGroup
		for index, val := range slice {
			if privateTagsGroup == nil {
				privateTagsGroup = tagsGroupTree.tree.Find(datastructures.NewStringTreeKey(val)).(*tagsGroup)
			} else {
				privateTagsGroup = privateTagsGroup.children.tree.Find(datastructures.NewStringTreeKey(val)).(*tagsGroup)
			}

			// only the last index (in this case "e") should have a value set
			if index != len(slice)-1 {
				g.Expect(privateTagsGroup.value).To(BeNil())
			} else {
				g.Expect(privateTagsGroup.value).ToNot(BeNil())
			}
		}
	})

	t.Run("can insert on a subtree if it is free", func(t *testing.T) {
		tagsGroupTree := NewTagsGroup()
		slice := []string{"a", "b", "c", "d", "e"}

		_, err := tagsGroupTree.FindOrCreateTagsGroup(slice, callback)
		g.Expect(err).ToNot(HaveOccurred())

		_, err = tagsGroupTree.FindOrCreateTagsGroup([]string{"a", "b"}, callback)
		g.Expect(err).ToNot(HaveOccurred())

		var privateTagsGroup *tagsGroup
		for _, val := range slice {
			if privateTagsGroup == nil {
				privateTagsGroup = tagsGroupTree.tree.Find(datastructures.NewStringTreeKey(val)).(*tagsGroup)
			} else {
				privateTagsGroup = privateTagsGroup.children.tree.Find(datastructures.NewStringTreeKey(val)).(*tagsGroup)
			}

			// only the last index (in this case "e") should have a value set
			if val == "b" || val == "e" {
				g.Expect(privateTagsGroup.value).ToNot(BeNil(), val)
			} else {
				g.Expect(privateTagsGroup.value).To(BeNil())
			}
		}
	})
}
