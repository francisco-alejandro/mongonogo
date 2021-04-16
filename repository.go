package mongonogo

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type versioner interface {
	ReadVersion(context.Context) (int, error)
	WriteVersion(context.Context, Migration) error
	DeleteVersion(context.Context, Migration) error
}

type version struct {
	Collection collectionManager
}

func (v *version) ReadVersion(ctx context.Context) (int, error) {
	var migration Migration

	options := options.FindOne()
	options.SetSort(bson.D{{"version", -1}})

	err := v.Collection.FindOne(ctx, bson.M{}, options).Decode(&migration)
	switch err {
	case nil, mongo.ErrNoDocuments:
		return migration.Version, nil
	default:
		return -1, err
	}
}

func (v *version) WriteVersion(ctx context.Context, migration Migration) error {
	_, err := v.Collection.InsertOne(ctx, bson.D{
		{Key: "version", Value: migration.Version},
		{Key: "timestamp", Value: time.Now()},
	})

	return err
}

func (v *version) DeleteVersion(ctx context.Context, migration Migration) error {
	_, err := v.Collection.DeleteOne(ctx, bson.M{"version": migration.Version})

	return err
}
