package main

import (
	"os"
	"os/signal"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
)

var version = "dev"
var globalWg, checkWg sync.WaitGroup

func checkFolder(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.MkdirAll(path, 0700)
	}
	return nil
}

func checkContainers() {
	conf := NewConfiguration()

	for _, container := range GetContainers(conf) {
		cert := Certificate{container, GetAccount(container.email, conf)}
		log.WithFields(log.Fields{"CID": container.ID, "Cert": cert}).Debug("Found LE container")

		checkWg.Add(1)
		cert.generateCertificate()
	}

	// We're done here
	log.Info("Done, sleeping for 1 hour")
	globalWg.Done()
}

func signalHandler(ticker *time.Ticker) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

signalLoop:
	for {
		sig := <-c
		switch sig {
		case os.Interrupt:
			log.WithField("signal", "SIGTERM").Info("Terminating...")
			ticker.Stop()
			break signalLoop
		}
	}

	globalWg.Done()
}

func tickerHandler(ticker *time.Ticker) {
	select {
	case <-ticker.C:
		globalWg.Add(1)
		go checkContainers()
	default:
		{
		}
	}

	globalWg.Done()
}

func main() {
	ticker := time.NewTicker(time.Hour)

	// Immediately run a check on start
	globalWg.Add(1)
	go checkContainers()

	// Set up our ticker
	globalWg.Add(1)
	go tickerHandler(ticker)

	// Set up our signal handler
	globalWg.Add(1)
	go signalHandler(ticker)

	// Wait for all goroutines
	globalWg.Wait()
}
