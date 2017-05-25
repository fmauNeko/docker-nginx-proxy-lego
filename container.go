package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

// Container represents a Docker container to be secured
type Container struct {
	ID    string
	hosts []string
	email string
	test  bool
}

// GetContainers gets the list of containers to be secured from the Docker server
func GetContainers(conf *Configuration) []*Container {
	leContainers := make([]*Container, 0)
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

			leContainer := Container{cID, hostsArray, envMap["LETSENCRYPT_EMAIL"], testBool}
			leContainers = append(leContainers, &leContainer)
			log.WithField("Container", leContainer).Debug("Found LE container")
		}
	}

	if err := client.Close(); err != nil {
		log.WithFields(log.Fields{"err": err}).Panic("Error")
	}

	return leContainers
}
