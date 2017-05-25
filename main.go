package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
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
	apiHeaders := map[string]string{"User-Agent": "Nginx-Proxy-Lego/" + version}
	apiVersion := "1.29"

	client, err := client.NewClient(conf.dEndpoint, apiVersion, nil, apiHeaders)
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Panic("Error")
	}

	containers, err := client.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Panic("Error")
	}

	inspectedContainers := make([]types.ContainerJSON, len(containers))

	for i, container := range containers {
		inspectedContainer, err := client.ContainerInspect(context.Background(), container.ID)

		if err != nil {
			log.WithFields(log.Fields{"err": err}).Panic("Error")
		}

		inspectedContainers[i] = inspectedContainer
	}

	for _, inspectedContainer := range inspectedContainers {
		cID := fmt.Sprintf("%.12s", inspectedContainer.ID)
		envMap := make(map[string]string)

		for _, envVar := range inspectedContainer.Config.Env {
			envSplit := strings.Split(envVar, "=")
			envMap[envSplit[0]] = envSplit[1]
		}

		if hosts, ok := envMap["LETSENCRYPT_HOST"]; ok {
			hostsArray := strings.Split(hosts, ",")
			testBool, err := strconv.ParseBool(envMap["LETSENCRYPT_TEST"])

			if err != nil {
				testBool = false
			}

			cert := Certificate{hostsArray, GetAccount(envMap["LETSENCRYPT_EMAIL"], conf), testBool}
			log.WithFields(log.Fields{"CID": cID, "Cert": cert}).Debug("Found LE container")

			checkWg.Add(1)
			cert.generateCertificate()
		}
	}

	if err := client.Close(); err != nil {
		log.WithFields(log.Fields{"err": err}).Panic("Error")
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
