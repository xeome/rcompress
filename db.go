package main

import (
	"crypto/sha256"
	"database/sql"
	"fmt"
	"os"
)

func initDB() (*sql.DB, error) {
	DB, err := sql.Open("sqlite3", config.Dbpath)
	if err != nil {
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

	_, err = DB.Exec("PRAGMA cache_size = -8192")
	if err != nil {
		return nil, err
	}

	_, err = DB.Exec("PRAGMA journal_mode = WAL")
	if err != nil {
		return nil, err
	}

	return DB, nil
}

func addFileToDB(db *sql.DB, path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Errorf("Error reading file: %q", err)
		return
	}

	hash := fmt.Sprintf("%x", sha256.Sum256(data))

	_, err = db.Exec("INSERT INTO files (hash, path) VALUES (?, ?)", hash, path)
	if err != nil {
		log.Warnf("Error inserting file %s into database: %q", path, err)
	}
}

func isAlreadyCompressed(db *sql.DB, path string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Errorf("Error reading file: %q", err)
		return false
	}

	hash := fmt.Sprintf("%x", sha256.Sum256(data))

	row := db.QueryRow("SELECT hash FROM files WHERE hash = ?", hash)
	var retrievedHash string
	err = row.Scan(&retrievedHash)
	if err != nil {
		if err == sql.ErrNoRows {
			return false
		}
		log.Errorf("Error scanning row: %q", err)
		return false
	}
	log.Tracef("Skipping %s", path)
	return true
}
