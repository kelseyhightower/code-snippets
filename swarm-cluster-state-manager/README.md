# Docker Swarm Cluster State Manager

This is not an official Google product.

## Overview

The Swarm cluster state manager is an example application on how cluster state could be managed in Swarm. The ideas for this come directly from Kubernetes.

## Usage

### Start the swarm cluster state manager

```
$ swarm-cluster-state-manager \
  --addr="tcp://127.0.0.1:2376" \
  --tlscacert="~/.docker/machine/certs/ca.pem" \
  --tlscert="~/.docker/machine/certs/cert.pem" \
  --tlskey="~/.docker/machine/certs/key.pem"
```

### Sumbit a new cluster state object

```
$ curl http://<swarm-cluster-state-manager:IP>:8080/submit -d '{"Name": "nginx", "Image": "nginx:1.9.6", "Count": 3}'
```

### Get the current status

```
$ curl http://<swarm-cluster-state-manager:IP>:8080/status
```
