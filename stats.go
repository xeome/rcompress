package main

import (
	"fmt"
	"os"
)

type Stats struct {
	processedFiles      int64
	totalSize           int64
	totalCompressedSize int64
	totalFiles          int64
}

func (s *Stats) print() {
	fmt.Printf("Compressed %d files\n", s.processedFiles)
	fmt.Printf("Total size: %d KB\n", s.totalSize)
	fmt.Printf("Total compressed size: %d KB\n", s.totalCompressedSize)
	fmt.Printf("Saved %d KB\n", s.totalSize-s.totalCompressedSize)
	fmt.Printf("Reduced size by %.2f%%\n", float64(s.totalSize-s.totalCompressedSize)/float64(s.totalSize)*100)
}

func (s *Stats) update(path string, info os.FileInfo) {
	s.totalSize += info.Size() / 1024

	infoAfter, err := os.Stat(path)
	if err != nil {
		log.Errorf("Error getting file info: %q", err)
		return
	}

	s.totalCompressedSize += infoAfter.Size() / 1024
	s.processedFiles++
}
