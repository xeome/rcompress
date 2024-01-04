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
	// Human readable output, CLI flag -human
	if config.human {
		fmt.Printf("%-22s %d\n", "Compressed files:", s.processedFiles)
		fmt.Printf("%-22s %s\n", "Total size:", humanSize(s.totalSize))
		fmt.Printf("%-22s %s\n", "Total compressed size:", humanSize(s.totalCompressedSize))
		fmt.Printf("%-22s %s\n", "Saved:", humanSize(s.totalSize-s.totalCompressedSize))
		fmt.Printf("%-22s %.2f%%\n", "Reduced size by:", float64(s.totalSize-s.totalCompressedSize)/float64(s.totalSize)*100)
		return
	}

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

func humanSize(size int64) string {
	units := []string{"KiB", "MiB", "GiB", "TiB", "PiB"}
	unitIndex := 0

	sizeFloat := float64(size)
	for sizeFloat >= 1024 && unitIndex < len(units)-1 {
		sizeFloat /= 1024
		unitIndex++
	}

	return fmt.Sprintf("%.2f %s", sizeFloat, units[unitIndex])
}
