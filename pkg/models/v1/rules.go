package v1

import "github.com/DanLavine/willow/pkg/models/datatypes"

type UniqueKeys map[string]struct{}

// All rules are stored in a BTree bassed on their name? How do we get the Order
// for the possible select cases to know what to run?
//
// Also, how do we find the rules to execute a tags group against?
type Rule struct {
	Name datatypes.String

	// This somehow needs to be checked against the arbitrary tags
	// This isn't saved in the composite tree, but is saved in the Int value of the
	// composite tree. So that needs to be broken out.
	//
	// not true. It gets "close", but then would need to double check all the rules... good enough?
	// maybe not the end of the world if its just so slow to start with?
	//
	// if thats the case, we could just iterate through a complete ComositeTree table
	// ... for all possible tag combinations. That might not be great
	//
	// This requires more reasearch. The disjoint set probably isn't gret eather, need a way of
	// figuring out an optimal way to do this. What I am planning could work, but isn't great
	GroupBy UniqueKeys

	Selection []Select
}

type Select struct {
	// The order on which this group should be acted on. Higher orders are checked first
	Order uint

	Where *[]WhereClause

	// Limits for whan a rule is found
	Limit *Limit
}

// This is what is actually stored in the CompositeTree table. The Limits that
// can be acted on
type Limit struct {
	Max uint

	active uint
}

func (l *Limit) Max() bool {
	return l.active >= l.Max
}

func (l *Limit) Increment() {
	l.active++
}

func (l *Limit) Decriment() {
	l.active--
}
