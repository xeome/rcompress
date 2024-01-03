package main

import (
	"crypto/sha256"
	"database/sql"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

var (
	dbPath      string
	compressdir string
	maxconcur   int
)

func initFlags() {
	SetupLogger()
	flag.StringVar(&dbPath, "db", "./oxipng.db", "Path to database")
	flag.StringVar(&compressdir, "dir", ".", "Path to directory to compress")
	flag.IntVar(&maxconcur, "maxconcur", 4, "Maximum number of concurrent compressions")
	flag.Parse()
}

func main() {
	initFlags()

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("Error opening database: %q", err)
	}
	defer db.Close()

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS files (hash TEXT PRIMARY KEY, path TEXT)")
	if err != nil {
		panic(err)
	}

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, maxconcur)

	totalFiles := 0
	processedFiles := 0

	err = filepath.Walk(compressdir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if filepath.Ext(path) == ".png" {
			totalFiles++
			absPath, err := filepath.Abs(path)
			if err != nil {
				return err
			}

			wg.Add(1)
			semaphore <- struct{}{} // Acquire semaphore
			go func() {
				defer func() {
					<-semaphore // Release semaphore
					processedFiles++
					log.Infof("Finished compressing %s (%d/%d)", path, processedFiles, totalFiles)
					wg.Done()
				}()
				data, err := os.ReadFile(absPath)
				if err != nil {
					fmt.Println(err)
					return
				}

				hash := fmt.Sprintf("%x", sha256.Sum256(data))

				row := db.QueryRow("SELECT * FROM files WHERE hash = ?", hash)
				var dummy string
				if row.Scan(&dummy, &dummy) != sql.ErrNoRows {
					return
				}

				cmd := exec.Command("oxipng", "-t", "4", "--fast", "-o", "max", "--strip", "safe", "--preserve", "-q", absPath)
				err = cmd.Run()
				if err != nil {
					fmt.Println(err)
					return
				}

				hash = fmt.Sprintf("%x", sha256.Sum256(data))

				_, err = db.Exec("INSERT INTO files (hash, path) VALUES (?, ?)", hash, absPath)
				if err != nil {
					fmt.Println(err)
				}
			}()
		}

		return nil
	})
	if err != nil {
		panic(err)
	}

	wg.Wait()
	close(semaphore)
}
