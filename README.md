# Mongonogo

Tiny migration tool for MongoDB. Update your structure and data of MongoDB using the migration pattern.

Migrations data is saved into a collection for versioning, by default is called migrations.

## How to use

1. Register your migration
Create a package that includes all migrations files.

```go
package migrations

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"github.com/francisco-alejandro/mongonogo"
)

func init() {
	mongonogo.Register(func(db *mongo.Database) error {
		opt := options.Index().SetName("index")
		keys := bson.D{{"key", 1}}
		model := mongo.IndexModel{Keys: keys, Options: opt}
		_, err := db.Collection("collection").Indexes().CreateOne(context.TODO(), model)

		return err
	}, func(db *mongo.Database) error {
		_, err := db.Collection("collection").Indexes().DropOne(context.TODO(), "index")

		return err
	})
}
```
2. Run your migration script
First of all import your migrations package. Using your `*mongo.Database` pointer, instantiate the mongonogo API client and run your migrations.

```go
package main

import (
	...
	"github.com/francisco-alejandro/mongonogo"
	_ "path/to/migrations_package" // database migrations package
	...
)

func main() {
	...
	opt := mongonogo.Options{
		Schema: "migrations",
		Timeout: "10s",
		Database: database, // *mongo.Databse, handle connection before
	}

	client, err := mongonogo.New(opt)
	if err != nil {
		panic(err)
	}

	err = client.Up()
	if err != nil {
		panic(err)
	}
	...
}
```
