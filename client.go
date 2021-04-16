package mongonogo

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/mongo"
	"reflect"
	"sync"
	"time"
)

var migrations = migrationMap{}
var m sync.Mutex

// Client is mongo.Databse wrapper to run the migrations, it defines the mongonogo API
type Client struct {
	Versioner versioner
	Schema    string
	Timeout   time.Duration
	db        *mongo.Database
}

// Options defines the mongonogo API configuration
type Options struct {
	Schema   string `default:"migrations"`
	Timeout  string `default:"2s"`
	Database *mongo.Database
}

// New uses the provided options and returns a mongonog API client.
// Default options values are used if they are not provided
func New(opt Options) (*Client, error) {
	t := reflect.TypeOf(opt)
	if opt.Schema == "" {
		f, _ := t.FieldByName("Schema")
		opt.Schema = f.Tag.Get("default")
	}

	if opt.Timeout == "" {
		f, _ := t.FieldByName("Timeout")
		opt.Timeout = f.Tag.Get("default")
	}

	timeout, err := time.ParseDuration(opt.Timeout)
	if err != nil {
		return nil, err
	}

	d := database{db: opt.Database}
	v := version{
		Collection: d.Collection(opt.Schema),
	}

	return &Client{
		Versioner: &v,
		Schema:    opt.Schema,
		Timeout:   timeout,
		db:        opt.Database,
	}, nil
}

// ErrDuplicatedMigration is an error that ocurring during the migration registration process
var ErrDuplicatedMigration = errors.New("duplicated migration version")

// Register performs the migration logic registration.
// Up function is responsible for applying changes into DB
// Down function is responsible to unapplying changes into DB
func Register(v int, up func(*mongo.Database) error, down func(*mongo.Database) error) error {
	m.Lock()
	defer m.Unlock()

	if _, ok := migrations[v]; ok {
		return ErrDuplicatedMigration
	}

	migrations[v] = Migration{
		Version: v,
		Up:      up,
		Down:    down,
	}

	return nil
}

// Unregister performs the migration logic deregistration
func Unregister(v int) {
	_, ok := migrations[v]
	if ok {
		delete(migrations, v)
	}
}

// Version returns the last migration applied version
func (c *Client) Version() (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.Timeout)
	defer cancel()

	return c.Versioner.ReadVersion(ctx)
}

//Up applies all pending migrations
func (c *Client) Up() error {
	v, err := c.Version()
	if err != nil {
		return err
	}

	versions := migrations.Versions(false)

	for _, version := range versions {
		if version <= v {
			continue
		}

		err := func() error {
			migration := migrations[version]
			err := migration.Up(c.db)
			if err != nil {
				return err
			}

			ctx, cancel := context.WithTimeout(context.Background(), c.Timeout)
			defer cancel()

			return c.Versioner.WriteVersion(ctx, migration)
		}()

		if err != nil {
			return err
		}
	}

	return nil
}

// Down unapplies all migrations
func (c *Client) Down() error {
	v, err := c.Version()
	if err != nil {
		return err
	}

	versions := migrations.Versions(false)

	for _, version := range versions {
		if version > v {
			continue
		}

		err := func() error {
			migration := migrations[version]
			err := migration.Down(c.db)
			if err != nil {
				return err
			}

			ctx, cancel := context.WithTimeout(context.Background(), c.Timeout)
			defer cancel()

			return c.Versioner.DeleteVersion(ctx, migration)
		}()

		if err != nil {
			return err
		}
	}

	return nil
}
