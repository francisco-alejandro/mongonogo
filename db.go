package mongonogo

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type databaseManager interface {
	Collection(string) collectionManager
}

type database struct {
	db *mongo.Database
}

type querySingleResult interface {
	Decode(v interface{}) error
}

type collectionManager interface {
	FindOne(context.Context, interface{}, ...*options.FindOneOptions) querySingleResult
	InsertOne(context.Context, interface{}, ...*options.InsertOneOptions) (*mongo.InsertOneResult, error)
	DeleteOne(context.Context, interface{}, ...*options.DeleteOptions) (*mongo.DeleteResult, error)
}

type collection struct {
	c *mongo.Collection
}

func (r *database) Collection(c string) collectionManager {
	return &collection{
		c: r.db.Collection(c),
	}
}

func (m *collection) FindOne(ctx context.Context, q interface{}, o ...*options.FindOneOptions) querySingleResult {
	return m.c.FindOne(ctx, q, o...)
}

func (m *collection) InsertOne(ctx context.Context, d interface{}, o ...*options.InsertOneOptions) (*mongo.InsertOneResult, error) {
	return m.c.InsertOne(ctx, d, o...)
}

func (m *collection) DeleteOne(ctx context.Context, q interface{}, o ...*options.DeleteOptions) (*mongo.DeleteResult, error) {
	return m.c.DeleteOne(ctx, q, o...)
}
