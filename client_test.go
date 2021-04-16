package mongonogo_test

import (
	"context"
	"github.com/francisco-alejandro/mongonogo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"testing"
)

type versioner struct {
	mock.Mock
}

func (v *versioner) ReadVersion(c context.Context) (int, error) {
	args := v.Called(c)

	return args.Int(0), args.Error(1)
}

func (v *versioner) WriteVersion(c context.Context, m mongonogo.Migration) error {
	args := v.Called(c, m)

	return args.Error(0)
}

func (v *versioner) DeleteVersion(c context.Context, m mongonogo.Migration) error {
	args := v.Called(c, m)

	return args.Error(0)
}

func TestNew(t *testing.T) {
	ctx := context.Background()
	client, _ := mongo.Connect(ctx, options.Client())
	db := client.Database("test")

	testCases := []struct {
		name               string
		opt                mongonogo.Options
		expectedCollection string
		failed             bool
	}{
		{
			name:               "with default values",
			opt:                mongonogo.Options{Database: db},
			expectedCollection: "migrations",
			failed:             false,
		},
		{
			name: "with collection set",
			opt: mongonogo.Options{
				Schema:   "changes",
				Database: db,
				Timeout:  "10s",
			},
			expectedCollection: "changes",
			failed:             false,
		},
		{
			name: "with invalid timeout",
			opt: mongonogo.Options{
				Database: db,
				Timeout:  "invalid",
			},
			expectedCollection: "migrations",
			failed:             true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client, err := mongonogo.New(tc.opt)
			if err == nil {
				assert.Equal(t, tc.expectedCollection, client.Schema)
			}

			assert.Equal(t, tc.failed, err != nil)
		})
	}
}

func TestRegister(t *testing.T) {
	version := 1
	t.Cleanup(func() {
		mongonogo.Unregister(version)
	})
	testCases := []struct {
		name          string
		expectedError error
	}{
		{
			name:          "no duplicated migrations",
			expectedError: nil,
		},
		{
			name:          "duplicated migration",
			expectedError: mongonogo.ErrDuplicatedMigration,
		},
	}

	f := func(db *mongo.Database) error { return nil }

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := mongonogo.Register(version, f, f)

			assert.ErrorIs(t, err, tc.expectedError)
		})
	}
}

func TestVersion(t *testing.T) {
	version := 1
	m := versioner{}
	m.On("ReadVersion", mock.Anything).Return(version, nil)

	c := mongonogo.Client{
		Versioner: &m,
	}
	v, err := c.Version()

	assert.Equal(t, version, v)
	assert.Nil(t, err)
}

func TestUp_ErrVersion(t *testing.T) {
	m := versioner{}
	m.On("ReadVersion", mock.Anything).Return(-1, mongo.ErrClientDisconnected)

	c := mongonogo.Client{
		Versioner: &m,
	}
	err := c.Up()

	assert.ErrorIs(t, err, mongo.ErrClientDisconnected)
}

func TestUp_ErrMigrationUp(t *testing.T) {
	version := 1
	t.Cleanup(func() {
		mongonogo.Unregister(version)
	})

	f := func(db *mongo.Database) error {
		return mongo.ErrClientDisconnected
	}

	m := versioner{}
	m.On("ReadVersion", mock.Anything).Return(0, nil)

	c := mongonogo.Client{
		Versioner: &m,
	}
	mongonogo.Register(version, f, f)
	err := c.Up()

	assert.ErrorIs(t, err, mongo.ErrClientDisconnected)
}

func TestUp_Success(t *testing.T) {
	versions := []int{1, 2}
	t.Cleanup(func() {
		for _, v := range versions {
			mongonogo.Unregister(v)
		}
	})

	f := func(db *mongo.Database) error {
		return nil
	}

	m := versioner{}
	m.On("ReadVersion", mock.Anything).Return(1, nil)
	m.On("WriteVersion", mock.Anything, mock.Anything).Return(nil)

	c := mongonogo.Client{
		Versioner: &m,
	}

	for _, v := range versions {
		mongonogo.Register(v, f, f)
	}

	err := c.Up()

	assert.Nil(t, err)
}

func TestDown_ErrVersion(t *testing.T) {
	m := versioner{}
	m.On("ReadVersion", mock.Anything).Return(-1, mongo.ErrClientDisconnected)

	c := mongonogo.Client{
		Versioner: &m,
	}
	err := c.Down()

	assert.ErrorIs(t, err, mongo.ErrClientDisconnected)
}

func TestDown_ErrMigrationDown(t *testing.T) {
	version := 1
	t.Cleanup(func() {
		mongonogo.Unregister(version)
	})

	f := func(db *mongo.Database) error {
		return mongo.ErrClientDisconnected
	}

	m := versioner{}
	m.On("ReadVersion", mock.Anything).Return(1, nil)

	c := mongonogo.Client{
		Versioner: &m,
	}
	mongonogo.Register(version, f, f)
	err := c.Down()

	assert.ErrorIs(t, err, mongo.ErrClientDisconnected)
}

func TestDown_Success(t *testing.T) {
	versions := []int{1, 2}
	t.Cleanup(func() {
		for _, v := range versions {
			mongonogo.Unregister(v)
		}
	})

	f := func(db *mongo.Database) error {
		return nil
	}

	m := versioner{}
	m.On("ReadVersion", mock.Anything).Return(1, nil)
	m.On("DeleteVersion", mock.Anything, mock.Anything).Return(nil)

	c := mongonogo.Client{
		Versioner: &m,
	}

	for _, v := range versions {
		mongonogo.Register(v, f, f)
	}

	err := c.Down()

	assert.Nil(t, err)
}
