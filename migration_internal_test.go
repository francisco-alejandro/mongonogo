package mongonogo

import (
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"testing"
)

func TestMigration(t *testing.T) {
	m := migrationMap{
		1: Migration{
			ID:      primitive.NewObjectID(),
			Version: 1,
			Up:      func(db *mongo.Database) error { return nil },
			Down:    func(db *mongo.Database) error { return nil },
		},
		2: Migration{
			ID:      primitive.NewObjectID(),
			Version: 2,
			Up:      func(db *mongo.Database) error { return nil },
			Down:    func(db *mongo.Database) error { return nil },
		},
	}

	testCases := []struct {
		name     string
		reverse  bool
		expected []int
	}{
		{
			name:     "getting migrations asc order",
			reverse:  false,
			expected: []int{1, 2},
		},
		{
			name:     "getting migrations desc order",
			reverse:  true,
			expected: []int{2, 1},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			v := m.Versions(tc.reverse)

			assert.Equal(t, tc.expected, v)
		})
	}

}
