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
	"fmt"
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

type ClusterStateStatus struct {
	Name         string
	Image        string
	CurrentCount int
	DesiredCount int
	Labels       []string
	Containers   []dockerclient.Container
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

func (cm *ClusterStateManager) Status() ([]*ClusterStateStatus, error) {
	cs := make([]*ClusterStateStatus, 0)

	cm.mu.RLock()
	defer cm.mu.RUnlock()

	for _, state := range cm.Store {
		m := make(map[string][]string)
		m["label"] = []string{fmt.Sprintf("com.swarm.app=%s", state.Name)}
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
		cs = append(cs, &ClusterStateStatus{
			Name:         state.Name,
			Image:        state.Image,
			CurrentCount: len(containers),
			DesiredCount: state.Count,
			Labels:       m["label"],
			Containers:   containers,
		})
	}

	return cs, nil
}

func (cm *ClusterStateManager) Sync() {
	for {
		clusterStatus, err := cm.Status()
		if err != nil {
			log.Print(err)
			time.Sleep(5 * time.Second)
			continue
		}

		for _, status := range clusterStatus {
			delta := status.DesiredCount - status.CurrentCount
			if delta == 0 {
				continue
			}

			if delta > 0 {
				log.Printf("need to create %d containers", delta)
				labels := make(map[string]string)
				labels["com.swarm.app"] = status.Name
				for i := 0; i < delta; i++ {
					id, err := cm.dockerClient.CreateContainer(&dockerclient.ContainerConfig{Image: status.Image, Labels: labels}, "")
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

			log.Printf("need to delete %d containers", int(math.Abs(float64(delta))))
			for _, container := range status.Containers[status.DesiredCount:] {
				err = cm.dockerClient.RemoveContainer(container.Id, true, true)
				if err != nil {
					log.Print(err)
					continue
				}
				log.Printf("deleted container: %s", container.Id)
			}
		}

		time.Sleep(10 * time.Second)
	}
}
