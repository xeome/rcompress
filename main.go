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

	handleSignals(&processedFiles, &totalSize, &totalCompressedSize)
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

func handleSignals(processedFiles, totalSize, totalCompressedSize *int64) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		log.Infof("Received signal %s, exiting...", sig)
		printStats(*processedFiles, *totalSize, *totalCompressedSize)
		os.Exit(0)
	}()
}

func processFile(absPath string, db *sql.DB, totalSize, totalCompressedSize, processedFiles *int64, info os.FileInfo) {
	data, err := os.ReadFile(absPath)
	if err != nil {
		fmt.Println(err)
		return
	}

	hashBefore := fmt.Sprintf("%x", sha256.Sum256(data))

	if isAlreadyCompressed(db, hashBefore, absPath) {
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

	*totalSize += info.Size() / 1024

	info, err = os.Stat(absPath)
	if err != nil {
		log.Errorf("Error getting file info: %q", err)
		return
	}

	*totalCompressedSize += info.Size() / 1024
	*processedFiles++
}

func isAlreadyCompressed(db *sql.DB, hashBefore string, path string) bool {
	row := db.QueryRow("SELECT * FROM files WHERE hash = ?", hashBefore)
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
