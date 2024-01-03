package main

import (
	"crypto/sha256"
	"database/sql"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"

	_ "github.com/mattn/go-sqlite3"
)

var (
	dbPath      string
	compressdir string
	maxconcur   int
	db          *sql.DB
)

func main() {
	initFlags()
	initDB()
	defer db.Close()

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, maxconcur)
	defer close(semaphore)

	var totalFiles, processedFiles, totalSize, totalCompressedSize int64

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigs
		fmt.Printf("Compressed %d files\n", totalFiles)
		fmt.Printf("Total size: %d KB\n", totalSize)
		fmt.Printf("Total compressed size: %d KB\n", totalCompressedSize)
		fmt.Printf("Saved %d KB\n", totalSize-totalCompressedSize)
		fmt.Printf("Reduced size by %.2f%%\n", float64(totalSize-totalCompressedSize)/float64(totalSize)*100)
		os.Exit(0)
	}()

	err := filepath.Walk(compressdir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if filepath.Ext(path) != ".png" {
			return nil
		}

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

			hashBefore := fmt.Sprintf("%x", sha256.Sum256(data))

			log.Infof("Querying for %s (%s)", path, hashBefore)
			row := db.QueryRow("SELECT * FROM files WHERE hash = ?", hashBefore)
			var dummy string
			if row.Scan(&dummy, &dummy) != sql.ErrNoRows {
				log.Infof("Skipping %s", path)
				return
			}

			cmd := exec.Command("oxipng", "-t", "4", "--fast", "-o", "max", "--strip", "safe", "--preserve", "-q", absPath)
			err = cmd.Run()
			if err != nil {
				fmt.Println(err)
				return
			}

			dataAfter, err := os.ReadFile(absPath)
			if err != nil {
				fmt.Println(err)
				return
			}
			hashAfter := fmt.Sprintf("%x", sha256.Sum256(dataAfter))

			_, err = db.Exec("INSERT INTO files (hash, path) VALUES (?, ?)", hashAfter, absPath)
			if err != nil {
				log.Warnf("Error inserting into database: %q", err)
			}

			totalSize += info.Size() / 1024

			info, err = os.Stat(absPath)
			if err != nil {
				log.Errorf("Error getting file info: %q", err)
				return
			}

			totalCompressedSize += info.Size() / 1024
		}()

		return nil
	})

	if err != nil {
		panic(err)
	}

	wg.Wait()
	fmt.Printf("Compressed %d files\n", totalFiles)
	fmt.Printf("Total size: %d KB\n", totalSize)
	fmt.Printf("Total compressed size: %d KB\n", totalCompressedSize)
	fmt.Printf("Saved %d KB\n", totalSize-totalCompressedSize)
	fmt.Printf("Reduced size by %.2f%%\n", float64(totalSize-totalCompressedSize)/float64(totalSize)*100)
}

func initDB() {
	var err error
	db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("Error opening database: %q", err)
	}

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS files (hash TEXT PRIMARY KEY, path TEXT)")
	if err != nil {
		panic(err)
	}

	_, err = db.Exec("CREATE INDEX IF NOT EXISTS hash_index ON files (hash)")
	if err != nil {
		panic(err)
	}
}

func initFlags() {
	SetupLogger()
	flag.StringVar(&dbPath, "db", "./oxipng.db", "Path to database")
	flag.StringVar(&compressdir, "dir", ".", "Path to directory to compress")
	flag.IntVar(&maxconcur, "maxconcur", 4, "Maximum number of concurrent compressions")
	flag.Parse()
}
