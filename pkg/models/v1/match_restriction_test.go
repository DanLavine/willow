package v1_test

import (
	"testing"

	v1 "github.com/DanLavine/willow/pkg/models/v1"
	. "github.com/onsi/gomega"
)

func Test_MatchQuery_MatchRestrictions(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("an uknonw BrokerTagsMatch always returns false", func(t *testing.T) {
		matchQuery := &v1.MatchQuery{MatchRestriction: v1.MatchRestriction(-1), BrokerTags: []string{"a", "b"}}
		g.Expect(matchQuery.MatchTags([]string{"a"})).To(BeFalse())
	})

	t.Run("when BrokerTagsMatch is STRICT", func(t *testing.T) {
		t.Run("it only matches exact slices", func(t *testing.T) {
			matchQuery := &v1.MatchQuery{MatchRestriction: v1.STRICT, BrokerTags: []string{"a", "b"}}
			g.Expect(matchQuery.MatchTags([]string{"a", "b"})).To(BeTrue())
			g.Expect(matchQuery.MatchTags([]string{"a"})).To(BeFalse())
		})
	})

	t.Run("when BrokerTagsMatch is SUBSET", func(t *testing.T) {
		t.Run("it matches when the BrokerTags are a subset of the passed in tags", func(t *testing.T) {
			matchQuery := &v1.MatchQuery{MatchRestriction: v1.SUBSET, BrokerTags: []string{"a", "b"}}
			g.Expect(matchQuery.MatchTags([]string{"a", "b", "c"})).To(BeTrue())
			g.Expect(matchQuery.MatchTags([]string{"a", "b", "d"})).To(BeTrue())
			g.Expect(matchQuery.MatchTags([]string{"a", "d"})).To(BeFalse())
		})
	})

	t.Run("when BrokerTagsMatch is ANY", func(t *testing.T) {
		t.Run("it matches when any BrokerTags are in the passed in tags", func(t *testing.T) {
			matchQuery := &v1.MatchQuery{MatchRestriction: v1.ANY, BrokerTags: []string{"a", "b"}}
			g.Expect(matchQuery.MatchTags([]string{"a"})).To(BeTrue())
			g.Expect(matchQuery.MatchTags([]string{"b", "f", "2"})).To(BeTrue())
			g.Expect(matchQuery.MatchTags([]string{"c", "d"})).To(BeFalse())
		})
	})

	t.Run("when BrokerTagsMatch is ALL", func(t *testing.T) {
		t.Run("it always returns true", func(t *testing.T) {
			matchQuery := &v1.MatchQuery{MatchRestriction: v1.ALL, BrokerTags: []string{"a", "b"}}
			g.Expect(matchQuery.MatchTags([]string{"c"})).To(BeTrue())
			g.Expect(matchQuery.MatchTags([]string{})).To(BeTrue())
			g.Expect(matchQuery.MatchTags(nil)).To(BeTrue())
		})
	})
}
