package main

import (
	"os"
	"os/signal"
	"syscall"
)

func handleSignals(stats *Stats) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		log.Infof("Received signal %s, exiting...", sig)
		stats.print()
		os.Exit(0)
	}()
}
