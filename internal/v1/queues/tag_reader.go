package queues

import (
	"github.com/DanLavine/willow/internal/v1/models"
)

// TODO: this needs a rework to be fast. Right now the behavior is important, but
//			 this is a pain to solve. The problem is "SUBSET" query for queues. If we want to
//       create a new queue, but already have a client listening, they should be able to
//       receive a message from the new queue. That means that the channel already needs
//       exist.

// create filtered tag with [0] index being global
// TODO this needs to be on the disk queue managere object. breaks tests since they don't clean up
// the state between runs
var filteredTags = []filteredTag{{queueCount: 1, tags: nil, reader: make(chan *models.Location)}}

type filteredTag struct {
	// how many queues are using this channel
	queueCount int

	// tags associated with this channel
	tags []string

	// reader for the associated tags
	reader chan *models.Location
}

func createFilteredTags(tags []string) []filteredTag {
	allTags := []filteredTag{filteredTags[0]}
	allTags = append(allTags, addIndividual(tags)...)
	allTags = append(allTags, addCombination(tags, 0, 1)...)

	return allTags
}

func addIndividual(tags []string) []filteredTag {
	individualTags := []filteredTag{}

	for _, tag := range tags {
		foundTag := false

		// see if the tag already exists
		for _, filteredTag := range filteredTags {
			if tagEqual(filteredTag.tags, tag) {
				foundTag = true
				filteredTag.queueCount++
				individualTags = append(individualTags, filteredTag)
				break
			}
		}

		// didn't find the 1 tag, so create the rest of them. Since tags are sorted this is fine
		if !foundTag {
			newTag := filteredTag{queueCount: 1, tags: []string{tag}, reader: make(chan *models.Location)}
			filteredTags = append(filteredTags, newTag)
			individualTags = append(individualTags, newTag)
		}
	}

	return individualTags
}

func addCombination(tags []string, startIndex, nextIndex int) []filteredTag {
	comboTags := []filteredTag{}
	tagsLen := len(tags)

	if nextIndex >= tagsLen-1 {
		return comboTags
	}

	tagsCombo := []string{tags[startIndex], tags[nextIndex]}

	advanceStart := false
	if nextIndex == tagsLen-1 && startIndex < tagsLen-2 {
		advanceStart = true
	}

	for nextIndex < tagsLen-1 {
		foundTags := false

		for _, filteredTag := range filteredTags {
			if tagsEqual(filteredTag.tags, tagsCombo) {
				foundTags = true
				filteredTag.queueCount++
				comboTags = append(comboTags, filteredTag)
				break
			}
		}

		// not found so create
		if !foundTags {
			newTag := filteredTag{queueCount: 1, tags: tagsCombo, reader: make(chan *models.Location)}
			filteredTags = append(filteredTags, newTag)
			comboTags = append(comboTags, newTag)
		}

		nextIndex++

		if nextIndex < tagsLen-1 {
			tagsCombo = append(tagsCombo, tags[nextIndex])
		}
	}

	if nextIndex <= tagsLen-1 {
		return append(comboTags, addCombination(tags, startIndex, nextIndex+1)...)
	} else if advanceStart {
		return append(comboTags, addCombination(tags, startIndex+1, startIndex+2)...)
	}

	return comboTags
}

func findOrCreateSubset(tags []string) chan *models.Location {
	// found tag
	for _, filteredTag := range filteredTags {
		if tagsEqual(filteredTag.tags, tags) {
			return filteredTag.reader
		}
	}

	// create new tag
	reader := make(chan *models.Location)
	filteredTags = append(filteredTags, filteredTag{queueCount: 1, tags: tags, reader: reader})

	return reader
}

func findAny(tags []string) []chan *models.Location {
	chans := []chan *models.Location{}

	for _, filteredTag := range filteredTags {
		if tagsContaing(filteredTag.tags, tags) {
			chans = append(chans, filteredTag.reader)
			continue
		}
	}

	return chans
}

func tagsContaing(tags, compareTags []string) bool {
	for _, tag := range tags {
		for _, compare := range compareTags {
			if compare == tag {
				return true
			}
		}
	}

	return false
}

func tagEqual(tags []string, matchTag string) bool {
	if len(tags) != 1 {
		return false
	}

	return tags[0] == matchTag
}

func tagsEqual(tags1, tags2 []string) bool {
	if len(tags1) != len(tags2) {
		return false
	}

	for index, tag := range tags1 {
		if tag != tags2[index] {
			return false
		}
	}

	return true
}
