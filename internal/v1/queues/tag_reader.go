package queues

import (
	"github.com/DanLavine/willow/internal/v1/models"
)

// TODO: this needs a rework to be fast. Right now the behavior is important, but
//			 this is a pain to solve. The problem is "SUBSET" query for queues. If we want to
//       create a new queue, but already have a client listening, they should be able to
//       receive a message from the new queue. That means that the channel already needs
//       exist.

type filteredTags struct {
	allTags []filteredTag
}

type filteredTag struct {
	// how many queues are using this channel
	queueCount int

	// tags associated with this channel
	tags []string

	// reader for the associated tags
	reader chan *models.Location
}

func NewFilteredTags() *filteredTags {
	return &filteredTags{
		allTags: []filteredTag{{queueCount: 1, tags: nil, reader: make(chan *models.Location)}}, // create with "global" chan
	}
}

func (f *filteredTags) createFilteredTags(tags []string) []chan *models.Location {
	allReaders := []chan *models.Location{f.allTags[0].reader}
	allReaders = append(allReaders, f.addIndividual(tags)...)
	allReaders = append(allReaders, f.addCombination(tags, 0, 1)...)

	return allReaders
}

func (f *filteredTags) addIndividual(tags []string) []chan *models.Location {
	individualReaders := []chan *models.Location{}

	for _, tag := range tags {
		foundTag := false

		// see if the tag already exists
		for index, filteredTag := range f.allTags {
			if tagEqual(filteredTag.tags, tag) {
				foundTag = true
				f.allTags[index].queueCount++
				individualReaders = append(individualReaders, filteredTag.reader)
				break
			}
		}

		if !foundTag {
			newTag := filteredTag{queueCount: 1, tags: []string{tag}, reader: make(chan *models.Location)}
			f.allTags = append(f.allTags, newTag)
			individualReaders = append(individualReaders, newTag.reader)
		}
	}

	return individualReaders
}

func (f *filteredTags) addCombination(tags []string, startIndex, nextIndex int) []chan *models.Location {
	comboReaders := []chan *models.Location{}
	tagsLen := len(tags)

	if nextIndex >= tagsLen-1 {
		return comboReaders
	}

	tagsCombo := []string{tags[startIndex], tags[nextIndex]}

	advanceStart := false
	if nextIndex == tagsLen-1 && startIndex < tagsLen-2 {
		advanceStart = true
	}

	for nextIndex < tagsLen-1 {
		foundTags := false

		for index, filteredTag := range f.allTags {
			if tagsEqual(filteredTag.tags, tagsCombo) {
				foundTags = true
				f.allTags[index].queueCount++
				comboReaders = append(comboReaders, filteredTag.reader)
				break
			}
		}

		// not found so create
		if !foundTags {
			newTag := filteredTag{queueCount: 1, tags: tagsCombo, reader: make(chan *models.Location)}
			f.allTags = append(f.allTags, newTag)
			comboReaders = append(comboReaders, newTag.reader)
		}

		nextIndex++

		if nextIndex < tagsLen-1 {
			tagsCombo = append(tagsCombo, tags[nextIndex])
		}
	}

	if nextIndex <= tagsLen-1 {
		return append(comboReaders, f.addCombination(tags, startIndex, nextIndex+1)...)
	} else if advanceStart {
		return append(comboReaders, f.addCombination(tags, startIndex+1, startIndex+2)...)
	}

	return comboReaders
}

func (f *filteredTags) findOrCreateSubset(tags []string) chan *models.Location {
	// found tag
	for _, filteredTag := range f.allTags {
		if tagsEqual(filteredTag.tags, tags) {
			return filteredTag.reader
		}
	}

	// create new tag
	reader := make(chan *models.Location)
	f.allTags = append(f.allTags, filteredTag{queueCount: 1, tags: tags, reader: reader})

	return reader
}

func (f *filteredTags) findAny(tags []string) []chan *models.Location {
	chans := []chan *models.Location{}

	for _, filteredTag := range f.allTags {
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
