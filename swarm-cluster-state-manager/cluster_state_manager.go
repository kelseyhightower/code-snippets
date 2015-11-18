// Copyright 2015 Google, Inc All rights reserved.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
//
// You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package main

import (
	"crypto/tls"
	"encoding/json"
	"log"
	"math"
	"sync"
	"time"

	"github.com/samalba/dockerclient"
)

type ClusterState struct {
	Count int
	Image string
	Name  string
}

type ClusterStateManager struct {
	dockerClient *dockerclient.DockerClient
	Store        map[string]*ClusterState
	mu           *sync.RWMutex
}

func (cm *ClusterStateManager) Submit(cs *ClusterState) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.Store[cs.Name] = cs
}

func newClusterStateManager(daemonUrl string, tlsConfig *tls.Config) (*ClusterStateManager, error) {
	store := make(map[string]*ClusterState)
	client, err := dockerclient.NewDockerClient(daemonUrl, tlsConfig)
	if err != nil {
		return nil, err
	}
	return &ClusterStateManager{client, store, &sync.RWMutex{}}, nil
}

func (cm *ClusterStateManager) Sync() {
	for {
		cm.mu.RLock()
		for _, state := range cm.Store {
			m := make(map[string][]string)
			m["label"] = []string{"com.swarm.app=nginx"}
			data, err := json.Marshal(m)
			if err != nil {
				log.Println(err)
				continue
			}
			containers, err := cm.dockerClient.ListContainers(false, false, string(data))
			if err != nil {
				log.Println(err)
				continue
			}
			log.Println(len(containers))
			difference := state.Count - len(containers)
			if difference == 0 {
				continue
			}
			if difference > 0 {
				log.Printf("need to create %d containers", difference)
				labels := make(map[string]string)
				labels["com.swarm.app"] = state.Name
				for i := 0; i < difference; i++ {
					id, err := cm.dockerClient.CreateContainer(&dockerclient.ContainerConfig{Image: state.Image, Labels: labels}, "")
					if err != nil {
						log.Print(err)
						continue
					}
					err = cm.dockerClient.StartContainer(id, nil)
					if err != nil {
						log.Print(err)
						continue
					}
					log.Printf("created container: %s", id)
				}
				continue
			}
			log.Printf("need to delete %d containers", math.Abs(float64(difference)))
			for _, container := range containers[state.Count:] {
				err = cm.dockerClient.RemoveContainer(container.Id, true, true)
				if err != nil {
					log.Print(err)
					continue
				}
				log.Printf("deleted container: %s", container.Id)
			}
		}
		cm.mu.RUnlock()
		time.Sleep(10 * time.Second)
	}
}
