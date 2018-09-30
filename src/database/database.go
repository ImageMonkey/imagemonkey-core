package imagemonkeydb

import (
	"database/sql"
	"github.com/getsentry/raven-go"
)

type ImageMonkeyDatabase struct {
	db *sql.DB
}

func NewImageMonkeyDatabase() *ImageMonkeyDatabase {
    return &ImageMonkeyDatabase{} 
}

func (p *ImageMonkeyDatabase) Open(connectionString string) error {
	var err error
    p.db, err = sql.Open("postgres", connectionString)
	if err != nil {
		return err
	}

	err = p.db.Ping()
	if err != nil {
		return err
	}

	return nil
}

func (p *ImageMonkeyDatabase) InitializeSentry(sentryDSN string, environment string) {
	raven.SetDSN(sentryDSN)
	raven.SetEnvironment(environment)
}

func (p *ImageMonkeyDatabase) Close() {
	p.db.Close()
}