package main

import (
	"os"
	"os/signal"
	"syscall"
)

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
