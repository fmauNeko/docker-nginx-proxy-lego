package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/fsouza/go-dockerclient"
)

type letsEncryptCertificate struct {
	host  []string
	email string
	test  bool
}

func main() {
	endpoint := "tcp://127.0.0.1:32768"
	client, err := docker.NewClient(endpoint)
	if err != nil {
		panic(err)
	}
	containers, err := client.ListContainers(docker.ListContainersOptions{All: true})
	if err != nil {
		panic(err)
	}
	inspectedContainers := make([]*docker.Container, len(containers))
	for i, container := range containers {
		inspectedContainer, err := client.InspectContainer(container.ID)
		if err != nil {
			panic(err)
		}
		inspectedContainers[i] = inspectedContainer
	}
	leMap := make(map[string]letsEncryptCertificate)
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
			leMap[cID] = letsEncryptCertificate{hostsArray, envMap["LETSENCRYPT_EMAIL"], testBool}
			fmt.Println("CID:", cID, "- LE:", leMap[cID])
		}
	}
}
