package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/fsouza/go-dockerclient"
)

type letsEncryptCertificate struct {
	hosts   []string
	account *Account
	test    bool
}

func checkFolder(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.MkdirAll(path, 0700)
	}
	return nil
}

func main() {
	conf := NewConfiguration()
	endpoint := "tcp://127.0.0.1:32768"
	client, err := docker.NewClient(endpoint)
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Panic("Error")
	}
	containers, err := client.ListContainers(docker.ListContainersOptions{All: true})
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Panic("Error")
	}
	inspectedContainers := make([]*docker.Container, len(containers))
	for i, container := range containers {
		inspectedContainer, err := client.InspectContainer(container.ID)
		if err != nil {
			log.WithFields(log.Fields{"err": err}).Panic("Error")
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
			leMap[cID] = letsEncryptCertificate{hostsArray, NewAccount(envMap["LETSENCRYPT_EMAIL"], conf), testBool}
			log.WithFields(log.Fields{"CID": cID, "LE": leMap[cID]}).Info("Found LE container")
		}
	}
}
