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
	dbPath      = flag.String("db", "./oxipng.db", "Path to database")
	compressdir = flag.String("dir", "./", "Path to directory to compress")
	maxconcur   = flag.Int("maxconcur", 8, "Maximum number of concurrent compressions")
)

func main() {
	flag.Parse()
	SetupLogger()
	db, err := initDB()
	if err != nil {
		log.Fatalf("Error opening database: %q", err)
		return
	}
	defer db.Close()

	var totalFiles, processedFiles, totalSize, totalCompressedSize int64
	walkRecursive(processedFiles, totalSize, totalCompressedSize, totalFiles, db)
	printStats(processedFiles, totalSize, totalCompressedSize)
}

func walkRecursive(processedFiles int64, totalSize int64, totalCompressedSize int64, totalFiles int64, db *sql.DB) {
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, *maxconcur)
	defer close(semaphore)

	handleSignals(&processedFiles, &totalSize, &totalCompressedSize) // non-blocking
	err := filepath.Walk(*compressdir, func(path string, info os.FileInfo, err error) error {
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
		semaphore <- struct{}{}
		go func() {
			defer func() {
				<-semaphore
				log.Infof("Finished compressing %s (%d/%d)", path, processedFiles, totalFiles)
				wg.Done()
			}()
			processFile(absPath, db, &totalSize, &totalCompressedSize, &processedFiles, info)
		}()

		return nil
	})

	if err != nil {
		panic(err)
	}

	wg.Wait()
}

func processFile(absPath string, db *sql.DB, totalSize, totalCompressedSize, processedFiles *int64, info os.FileInfo) {
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
	updateStats(absPath, totalSize, totalCompressedSize, info)
	*processedFiles++
}

func updateStats(path string, totalSize, totalCompressedSize *int64, infoBefore os.FileInfo) {
	*totalSize += infoBefore.Size() / 1024

	infoAfter, err := os.Stat(path)
	if err != nil {
		log.Errorf("Error getting file info: %q", err)
		return
	}

	*totalCompressedSize += infoAfter.Size() / 1024
}

func isAlreadyCompressed(db *sql.DB, path string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Errorf("Error reading file: %q", err)
		return false
	}

	hash := fmt.Sprintf("%x", sha256.Sum256(data))

	row := db.QueryRow("SELECT * FROM files WHERE hash = ?", hash)
	var dummy string
	if row.Scan(&dummy, &dummy) != sql.ErrNoRows {
		log.Infof("Skipping %s", path)
		return true
	}
	return false
}

func printStats(processedFiles int64, totalSize int64, totalCompressedSize int64) {
	fmt.Printf("Compressed %d files\n", processedFiles)
	fmt.Printf("Total size: %d KB\n", totalSize)
	fmt.Printf("Total compressed size: %d KB\n", totalCompressedSize)
	fmt.Printf("Saved %d KB\n", totalSize-totalCompressedSize)
	fmt.Printf("Reduced size by %.2f%%\n", float64(totalSize-totalCompressedSize)/float64(totalSize)*100)
}
