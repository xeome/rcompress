package main

import (
	"database/sql"
	"flag"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

var config Config

func main() {
	SetupLogger()
	config.setDefaults()
	config.Parse()
	flag.Parse()

	db, err := initDB()
	if err != nil {
		log.Fatalf("Error opening database: %q", err)
		return
	}
	defer db.Close()

	var stats Stats
	walkRecursive(db, &stats)
	stats.print()
}

func walkRecursive(db *sql.DB, stats *Stats) {
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, config.maxconcur)
	defer close(semaphore)

	handleSignals(stats) // non-blocking
	err := filepath.Walk(config.compressdir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if filepath.Ext(path) != ".png" {
			return nil
		}

		stats.totalFiles++
		absPath, err := filepath.Abs(path)
		if err != nil {
			return err
		}

		wg.Add(1)
		semaphore <- struct{}{}
		go func() {
			defer func() {
				<-semaphore
				wg.Done()
			}()
			processFile(absPath, db, stats, info)
		}()

		return nil
	})

	if err != nil {
		panic(err)
	}

	wg.Wait() // wait for all goroutines to finish
}

func processFile(absPath string, db *sql.DB, stats *Stats, info os.FileInfo) {
	if isAlreadyCompressed(db, absPath) {
		return
	}

	cmd := exec.Command("oxipng", "-t", "4", "--fast", "-o", "max", "--strip", "safe", "--preserve", "-q", absPath)
	err := cmd.Run()
	if err != nil {
		log.Errorf("Error running oxipng: %q", err)
		return
	}

	addFileToDB(db, absPath)
	stats.update(absPath, info)
	log.Infof("Compressed %s (%d/%d)", absPath, stats.processedFiles, stats.totalFiles)
}
