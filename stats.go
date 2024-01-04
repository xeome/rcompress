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
	fmt.Printf("%-22s %d\n", "Compressed files:", s.processedFiles)
	fmt.Printf("%-22s %d KiB\n", "Total size:", s.totalSize)
	fmt.Printf("%-22s %d KiB\n", "Total compressed size:", s.totalCompressedSize)
	fmt.Printf("%-22s %d KiB\n", "Saved:", s.totalSize-s.totalCompressedSize)
	fmt.Printf("%-22s %.2f%%\n", "Reduced size by:", float64(s.totalSize-s.totalCompressedSize)/float64(s.totalSize)*100)
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
