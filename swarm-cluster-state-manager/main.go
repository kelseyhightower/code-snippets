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
	csm *ClusterStateManager

	addr               string
	tlscacert          string
	tlscert            string
	tlskey             string
	insecureSkipVerify bool
)

func init() {
	flag.StringVar(&addr, "addr", "tcp://127.0.0.1:2376", "Docker Swarm manager address")
	flag.BoolVar(&insecureSkipVerify, "insecure-skip-verify", false, "Skip server certificate verification")
	flag.StringVar(&tlscacert, "tlscacert", "~/.docker/ca.pem", "Trust certs signed only by this CA")
	flag.StringVar(&tlscert, "tlscert", "~/.docker/cert.pem", "Path to TLS certificate file")
	flag.StringVar(&tlskey, "tlskey", "~/.docker/key.pem", "Path to TLS key file")
}

func SubmitClusterState(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var cs ClusterState
	err := decoder.Decode(&cs)
	if err != nil {
		log.Print(err)
		w.WriteHeader(500)
	}
	csm.Submit(&cs)
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
	tlsConfig := &tls.Config{
		Certificates:       []tls.Certificate{clientCert},
		RootCAs:            certPool,
		InsecureSkipVerify: insecureSkipVerify,
	}
	csm, err = newClusterStateManager(addr, tlsConfig)
	if err != nil {
		log.Fatal(err)
	}
	go csm.Sync()

	http.HandleFunc("/submit", SubmitClusterState)
	fmt.Println("Starting Swarm cluster state manager...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
