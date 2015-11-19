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
	"crypto/x509"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

var (
	clusterStateManager *ClusterStateManager

	addr               string
	swarmManager       string
	tlscacert          string
	tlscert            string
	tlskey             string
	insecureSkipVerify bool
)

func init() {
	flag.StringVar(&addr, "addr", "127.0.0.1:2476", "HTTP listen address")
	flag.StringVar(&swarmManager, "swarm-manager", "tcp://127.0.0.1:2376", "Docker Swarm manager address")
	flag.BoolVar(&insecureSkipVerify, "insecure-skip-verify", false, "Skip server certificate verification")
	flag.StringVar(&tlscacert, "tlscacert", "~/.docker/ca.pem", "Trust certs signed only by this CA")
	flag.StringVar(&tlscert, "tlscert", "~/.docker/cert.pem", "Path to TLS certificate file")
	flag.StringVar(&tlskey, "tlskey", "~/.docker/key.pem", "Path to TLS key file")
}

func GetClusterState(w http.ResponseWriter, r *http.Request) {
	clusterStatus, err := clusterStateManager.ClusterStatus()
	if err != nil {
		log.Print(err)
		w.WriteHeader(500)
	}
	data, err := json.MarshalIndent(clusterStatus, "", "  ")
	if err != nil {
		log.Print(err)
		w.WriteHeader(500)
	}
	w.Write(data)
}

func SubmitClusterState(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var cs ClusterState
	err := decoder.Decode(&cs)
	if err != nil {
		log.Print(err)
		w.WriteHeader(500)
	}
	if err := clusterStateManager.Submit(&cs); err != nil {
		log.Print(err)
		w.WriteHeader(500)
	}
}

func RemoveClusterState(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	if name == "" {
		log.Println("error removing cluster state: missing name parameter")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	clusterStateManager.Remove(name)
}

func main() {
	flag.Parse()

	clientCert, err := tls.LoadX509KeyPair(tlscert, tlskey)
	if err != nil {
		log.Fatal(err)
	}
	certPool := x509.NewCertPool()
	caCert, err := ioutil.ReadFile(tlscacert)
	if err != nil {
		log.Fatal(err)
	}
	certPool.AppendCertsFromPEM(caCert)

	// TLS and TLS client authentication is required.
	tlsConfig := &tls.Config{
		ClientAuth:         tls.RequireAndVerifyClientCert,
		ClientCAs:          certPool,
		Certificates:       []tls.Certificate{clientCert},
		RootCAs:            certPool,
		InsecureSkipVerify: insecureSkipVerify,
	}

	clusterStateManager, err = newClusterStateManager(swarmManager, tlsConfig)
	if err != nil {
		log.Fatal(err)
	}

	// Start the background job to sync desired state with Docker Swarm.
	go clusterStateManager.Sync()

	server := &http.Server{
		Addr:      addr,
		TLSConfig: tlsConfig,
	}
	http.HandleFunc("/submit", SubmitClusterState)
	http.HandleFunc("/status", GetClusterState)
	http.HandleFunc("/remove", RemoveClusterState)

	fmt.Println("Starting Swarm cluster state manager...")
	log.Fatal(server.ListenAndServeTLS("", ""))
}
