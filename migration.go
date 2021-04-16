package mongonogo

import (
	"sort"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// Migration defines operations and how migration is store into db
type Migration struct {
	ID      primitive.ObjectID `bson:"_id,omitempty"`
	Version int                `bson:"version,omitempty"`
	Up      func(db *mongo.Database) error
	Down    func(db *mongo.Database) error
}

type migrationMap map[int]Migration

func (m *migrationMap) Versions(reverse bool) []int {
	keys := make([]int, len(*m))

	i := 0
	for version := range *m {
		keys[i] = version
		i++
	}

	s := sort.IntSlice(keys)

	if reverse {
		sort.Sort(sort.Reverse(s))
	} else {
		sort.Sort(s)
	}

	return keys
}
