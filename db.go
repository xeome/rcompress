package main

import "database/sql"

func initDB() (*sql.DB, error) {
	DB, err := sql.Open("sqlite3", *dbPath)
	if err != nil {
		log.Fatalf("Error opening database: %q", err)
		return nil, err
	}
	_, err = DB.Exec("CREATE TABLE IF NOT EXISTS files (hash TEXT PRIMARY KEY, path TEXT)")
	if err != nil {
		return nil, err
	}

	_, err = DB.Exec("CREATE INDEX IF NOT EXISTS hash_index ON files (hash)")
	if err != nil {
		return nil, err
	}

	return DB, nil
}
