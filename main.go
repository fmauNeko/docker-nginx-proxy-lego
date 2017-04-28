package main

import (
	"fmt"

	"github.com/fsouza/go-dockerclient"
)

func main() {
	endpoint := "unix:///tmp/docker.sock"
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
	for _, inspectedContainer := range inspectedContainers {
		fmt.Println("Env : ", inspectedContainer.Config.Env)
	}
}
