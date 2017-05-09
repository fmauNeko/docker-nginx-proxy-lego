package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

var version = "dev"

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
			log.WithFields(log.Fields{"CID": cID, "LE": leMap[cID]}).Debug("Found LE container")
			log.WithFields(log.Fields{"Hosts": strings.Join(leMap[cID].hosts, " ")}).Info("Generating new certificate")
		}
	}
}
