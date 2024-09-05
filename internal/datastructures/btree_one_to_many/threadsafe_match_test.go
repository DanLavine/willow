package btreeonetomany

import (
	"testing"

	v1 "github.com/DanLavine/willow/pkg/models/api/common/v1"
	querymatchaction "github.com/DanLavine/willow/pkg/models/api/common/v1/query_match_action"
	"github.com/DanLavine/willow/pkg/models/datatypes"
	"github.com/DanLavine/willow/testhelpers/testmodels"

	. "github.com/onsi/gomega"
)

func TestOneToManyTree_MatchPermutations(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("Context parameters", func(t *testing.T) {
		t.Run("It returns an error if the oneID is empty", func(t *testing.T) {
			tree := NewThreadSafe()

			err := tree.MatchAction("", &querymatchaction.MatchActionQuery{}, nil)
			g.Expect(err).To(Equal(ErrorOneIDEmpty))
		})

		t.Run("It returns an error if the KeyValues are empty", func(t *testing.T) {
			tree := NewThreadSafe()

			err := tree.MatchAction("oneID", &querymatchaction.MatchActionQuery{}, nil)
			g.Expect(err).To(HaveOccurred())
			g.Expect(err.Error()).To(ContainSubstring("KeyValues: requires a length of at least 1, but recevied 0"))
		})

		t.Run("It returns an error if the KeyValues are invalid", func(t *testing.T) {
			tree := NewThreadSafe()

			err := tree.MatchAction("oneID", &querymatchaction.MatchActionQuery{
				KeyValues: querymatchaction.MatchKeyValues{
					"bad key": querymatchaction.MatchValue{
						Value: datatypes.EncapsulatedValue{Type: datatypes.T_int, Data: "nope"},
						TypeRestrictions: v1.TypeRestrictions{
							MinDataType: datatypes.T_any,
							MaxDataType: datatypes.T_any,
						},
					},
				},
			}, nil)
			g.Expect(err.Error()).To(ContainSubstring("KeyValues.[bad key].Value.Type: 'int' has Data of kind: string"))
		})

		t.Run("It returns an error if onPagination is nil", func(t *testing.T) {
			tree := NewThreadSafe()

			err := tree.MatchAction("oneID", &querymatchaction.MatchActionQuery{
				KeyValues: querymatchaction.MatchKeyValues{
					"bad key": querymatchaction.MatchValue{
						Value: datatypes.EncapsulatedValue{Type: datatypes.T_int, Data: 3},
						TypeRestrictions: v1.TypeRestrictions{
							MinDataType: datatypes.T_any,
							MaxDataType: datatypes.T_any,
						},
					},
				},
			}, nil)
			g.Expect(err).To(Equal(ErrorOnIterateNil))
		})
	})

	t.Run("It can match all items in the One Relation", func(t *testing.T) {
		tree := NewThreadSafe()

		// create 5 items all under the same one tree
		tree.CreateOrFind("one name", datatypes.KeyValues{"key1": datatypes.Int(1)}, func() any { return "1" }, func(oneToManyItem OneToManyItem) { panic("no") })
		tree.CreateOrFind("one name", datatypes.KeyValues{"key2": datatypes.Int(2)}, func() any { return "2" }, func(oneToManyItem OneToManyItem) { panic("no") })
		tree.CreateOrFind("one name", datatypes.KeyValues{"key3": datatypes.Int(3)}, func() any { return "3" }, func(oneToManyItem OneToManyItem) { panic("no") })
		tree.CreateOrFind("one name", datatypes.KeyValues{"key1": datatypes.Int(1), "key2": datatypes.Int(2)}, func() any { return "4" }, func(oneToManyItem OneToManyItem) { panic("no") })
		tree.CreateOrFind("one name", datatypes.KeyValues{"key4": datatypes.Int(4)}, func() any { return "5" }, func(oneToManyItem OneToManyItem) { panic("no") })

		foundPagination := []string{}
		onPaginationQuery := func(paginationItem OneToManyItem) bool {
			foundPagination = append(foundPagination, paginationItem.Value().(string))
			return true
		}

		matchKeys := &querymatchaction.MatchActionQuery{
			KeyValues: querymatchaction.MatchKeyValues{
				"key1": {Value: datatypes.Int(1), TypeRestrictions: testmodels.NoTypeRestrictions(g)},
				"key2": {Value: datatypes.Int(2), TypeRestrictions: testmodels.NoTypeRestrictions(g)},
				"key3": {Value: datatypes.Int(3), TypeRestrictions: testmodels.NoTypeRestrictions(g)},
			},
		}

		err := tree.MatchAction("one name", matchKeys, onPaginationQuery)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(len(foundPagination)).To(Equal(4))
		g.Expect(foundPagination).To(ContainElements([]string{"1", "2", "3", "4"}))
	})

	// TODO: error handling checks, but thats going to be reworked so ignore for now
}
