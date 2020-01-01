package imagemonkeydb

import (
	"context"
	"github.com/getsentry/raven-go"
	"github.com/jackc/pgx/v4/pgxpool"
)

type ImageMonkeyDatabase struct {
	db *pgxpool.Pool
}

func NewImageMonkeyDatabase() *ImageMonkeyDatabase {
	return &ImageMonkeyDatabase{}
}

func (p *ImageMonkeyDatabase) Open(connectionString string) error {
	var err error
	p.db, err = pgxpool.Connect(context.Background(), connectionString)
	if err != nil {
		return err
	}

	/*err = p.db.Ping(context.Background())
	if err != nil {
		return err
	}*/

	return nil
}

func (p *ImageMonkeyDatabase) InitializeSentry(sentryDSN string, environment string) {
	raven.SetDSN(sentryDSN)
	raven.SetEnvironment(environment)
}

func (p *ImageMonkeyDatabase) Close() {
	p.db.Close()
}
