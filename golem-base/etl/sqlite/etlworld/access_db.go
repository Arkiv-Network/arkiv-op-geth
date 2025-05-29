package etlworld

import (
	"context"
	"database/sql"
	"fmt"
)

func (e *ETLWorld) WithDB(ctx context.Context, dbfunc func(db *sql.DB) error) error {
	db, err := sql.Open("sqlite3", fmt.Sprintf("file:%s?mode=ro&_journal_mode=WAL", e.etlProcess.dbPath))
	if err != nil {
		return err
	}
	defer db.Close()
	return dbfunc(db)
}
