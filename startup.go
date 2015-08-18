package main

import (
	"encoding/json"
	"github.com/fsouza/go-dockerclient"
	"io/ioutil"
	"log"
	"strconv"
	"time"
)

type ConfigFile struct {
	Start   StartConfig `json:"start"`
	Configs HostTypes   `json:"configs"`
}

// ========= Start section =====================
type StartConfig struct {
	Images     []string      `json:"images"`
	Containers []string      `json:"containers"`
	Settings   StartSettings `json:"settings"`
}

type StartSettings struct {
	StopContainers bool `json:"stop-containers"`
}

// ========= Configs section =====================
type HostTypes struct {
	Images     []ImageConfig      `json:"images"`
	Containers []ContainerConfigs `json:"containers"`
}

// ========= Images ==============================
type ImageConfig struct {
	Image   string                        `json:"image"`
	Options docker.CreateContainerOptions `json:"options"`
}

// ========= Containers ==========================
type ContainerConfigs struct {
	Id         string             `json:"id"`
	Hostconfig *docker.HostConfig `json:"hostconfig"`
}

func LoadConfigFile() (config ConfigFile) {
	f, err := ioutil.ReadFile("config.json")
	if err != nil {
		log.Panic(err)
	}

	err = json.Unmarshal(f, &config)
	if err != nil {
		log.Printf("Error loading config file. %v", err)
	}

	return
}

func Start(config ConfigFile) {
	// Start up parameters
	if config.Start.Settings.StopContainers {
		StopAllContainers()
		DeleteAllContainers()
	}
	// Load containers
	// We will iterate through the specified containers and look for a matching
	// docker.HostConfig if it is specified in the config.

	// Iterate over containers to start up
	for i := range config.Start.Containers {
		var hostConfig *docker.HostConfig
		for j := range config.Configs.Containers {
			// Iterate over specified HostConfigs and look for a match
			if config.Configs.Containers[j].Id == config.Start.Containers[i] {
				hostConfig = config.Configs.Containers[j].Hostconfig
				break
			}
		}
		// Start the container
		err := startContainer(config.Start.Containers[i], hostConfig)
		if err != nil {
			log.Printf("Error starting container %v. Error: %v", config.Configs.Containers[i].Id, err)
		} else {
			log.Printf("Container %v started successfully.", config.Configs.Containers[i].Id)
		}
	}

	for i := range config.Start.Images {
		var opts docker.CreateContainerOptions
		var hostConfig *docker.HostConfig

		for j := range config.Configs.Images {
			hostConfig = &docker.HostConfig{}
			// Iterate over specified HostConfigs and look for a match
			if config.Configs.Images[j].Image == config.Start.Images[i] {
				opts = config.Configs.Images[j].Options
				if config.Configs.Images[j].Options.Name == "" {
					opts.Name = config.Configs.Images[j].Image + strconv.FormatInt(time.Now().UnixNano(), 10)
				}

				hostConfig = config.Configs.Images[j].Options.HostConfig

				break
			}
		}
		// Create a container
		container, err := createContainer(opts)
		if err != nil {
			log.Panic(err)
		}

		err = startContainer(container.ID, hostConfig)
		if err != nil {
			log.Printf("Error starting image %v. Error: %v", container.ID, err)
		} else {
			log.Printf("Image %v started successfully as %v.", container.ID, container.Name)
		}
	}
}

func StopAllContainers() {
	containers, err := listContainers(docker.ListContainersOptions{All: false})
	if err != nil {
		log.Panic(err)
	}

	if len(containers) == 0 {
		return
	}

	// Iterate over the containers and stop them.
	for i := range containers {
		err = stopContainer(containers[i].ID)
		if err != nil {
			log.Panic(err)
		}
		log.Printf("%v stopped.", containers[i].Names)
	}

	return
}

func DeleteAllContainers() {
	containers, err := listContainers(docker.ListContainersOptions{All: true})
	if err != nil {
		log.Panic(err)
	}

	if len(containers) == 0 {
		return
	}

	// Iterate over the containers and stop them.
	for i := range containers {
		var opts docker.RemoveContainerOptions
		opts.ID = containers[i].ID
		err = removeContainer(opts)
		if err != nil {
			log.Panic(err)
		}
		log.Printf("Container %v removed.", containers[i].Names)
	}

	return
}
