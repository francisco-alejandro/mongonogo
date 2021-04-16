package mongonogo

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"testing"
)

type collectionMock struct {
	mock.Mock
}

type singleResult struct {
	mock.Mock
}

func (s *singleResult) Decode(v interface{}) error {
	args := s.Called(v)

	return args.Error(0)
}

type dbMock struct {
	C collectionMock
}

func (m *collectionMock) FindOne(ctx context.Context, q interface{}, o ...*options.FindOneOptions) querySingleResult {
	args := m.Called(ctx, q, o)

	return args.Get(0).(querySingleResult)

}

func (m *collectionMock) InsertOne(ctx context.Context, d interface{}, o ...*options.InsertOneOptions) (*mongo.InsertOneResult, error) {
	args := m.Called(ctx, d)

	return args.Get(0).(*mongo.InsertOneResult), args.Error(1)

}

func (m *collectionMock) DeleteOne(ctx context.Context, q interface{}, o ...*options.DeleteOptions) (*mongo.DeleteResult, error) {
	args := m.Called(ctx, q)

	return args.Get(0).(*mongo.DeleteResult), args.Error(1)
}

func (m *dbMock) Collection(string) collectionManager {
	return &collectionMock{}
}

func TestReadVersion(t *testing.T) {
	testCases := []struct {
		name            string
		run             func(mock.Arguments)
		decodeError     error
		expectedVersion int
		expectedError   error
	}{
		{
			name:            "with unexpected error",
			run:             func(mock.Arguments) {},
			decodeError:     mongo.ErrClientDisconnected,
			expectedVersion: -1,
			expectedError:   mongo.ErrClientDisconnected,
		},
		{
			name:            "version not found",
			run:             func(mock.Arguments) {},
			decodeError:     mongo.ErrNoDocuments,
			expectedVersion: 0,
			expectedError:   nil,
		},
		{
			name: "without error",
			run: func(args mock.Arguments) {
				migration := args.Get(0).(*Migration)

				migration.Version = 2
			},
			decodeError:     nil,
			expectedVersion: 2,
			expectedError:   nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c := collectionMock{}
			q := singleResult{}

			q.On("Decode", mock.Anything).Return(tc.decodeError).Run(tc.run)
			c.On("FindOne", mock.Anything, mock.Anything, mock.Anything).Return(&q)

			r := version{Collection: &c}
			version, err := r.ReadVersion(context.Background())
			assert.Equal(t, tc.expectedVersion, version)
			assert.ErrorIs(t, err, tc.expectedError)
		})
	}
}

func TestWriteVersion(t *testing.T) {
	testCases := []struct {
		name          string
		insertError   error
		expectedError error
	}{
		{
			name:          "with error",
			insertError:   mongo.ErrClientDisconnected,
			expectedError: mongo.ErrClientDisconnected,
		},
		{
			name:          "without error",
			insertError:   nil,
			expectedError: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c := collectionMock{}
			c.On("InsertOne", mock.Anything, mock.Anything).Return(&mongo.InsertOneResult{}, tc.insertError)

			r := version{Collection: &c}
			err := r.WriteVersion(context.Background(), Migration{})
			assert.ErrorIs(t, err, tc.expectedError)
		})
	}
}

func TestDeleteVersion(t *testing.T) {
	testCases := []struct {
		name          string
		deleteError   error
		expectedError error
	}{
		{
			name:          "with error",
			deleteError:   mongo.ErrClientDisconnected,
			expectedError: mongo.ErrClientDisconnected,
		},
		{
			name:          "without error",
			deleteError:   nil,
			expectedError: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c := collectionMock{}
			c.On("DeleteOne", mock.Anything, mock.Anything).Return(&mongo.DeleteResult{}, tc.deleteError)

			r := version{Collection: &c}
			err := r.DeleteVersion(context.Background(), Migration{})
			assert.ErrorIs(t, err, tc.expectedError)
		})
	}
}
