package imagemonkeydb

import (
	datastructures "github.com/bbernhard/imagemonkey-core/datastructures"
	log "github.com/sirupsen/logrus"
	"github.com/getsentry/raven-go"
	"context"
)

func (p *ImageMonkeyDatabase) GetPgStatStatements() ([]datastructures.PgStatStatementResult, error) {
	res := []datastructures.PgStatStatementResult{}

	rows, err := p.db.Query(context.TODO(),
							`SELECT 
  								(total_exec_time / 1000 / 60) as total, 
 	 							(total_exec_time/calls) as avg, 
  								query 
							 FROM pg_stat_statements 
							 ORDER BY 1 DESC 
							 LIMIT 100`)
	if err != nil {
		log.Error("[PostgreSQL Statistics] Couldn't get pg statistics: ", err.Error())
		raven.CaptureError(err, nil)
		return res, err
	}

	defer rows.Close()

	for rows.Next() {
		var r datastructures.PgStatStatementResult
		err = rows.Scan(&r.Total, &r.Avg, &r.Query)
		if err != nil {
			log.Error("[PostgreSQL Statistics] Couldn't scan pg statistics: ", err.Error())
			raven.CaptureError(err, nil)
			return res, err
		}

		res = append(res, r)
	}

	return res, nil
}
