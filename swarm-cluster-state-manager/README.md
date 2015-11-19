# Docker Swarm Cluster State Manager

This is not an official Google product.

## Overview

The Swarm cluster state manager is an example application on how cluster state could be managed in Swarm. The ideas for this come directly from Kubernetes.

## Usage

### Start the swarm cluster state manager

```
$ swarm-cluster-state-manager \
  --addr 127.0.0.1:2476 \
  --swarm-manager "tcp://104.197.107.13:2376" \
  --tlscacert ~/.docker/machine/certs/ca.pem \
  --tlscert ~/.docker/machine/certs/cert.pem \
  --tlskey ~/.docker/machine/certs/key.pem
```
```
Starting Swarm cluster state manager...
```

### Reusing the Docker machine client certs

#### OS X

On OS X you'll need to create a pkcs12 bundle of the PEM encoded client certs.

```
$ openssl pkcs12 -export \
  -in ~/.docker/machine/certs/cert.pem \
  -inkey ~/.docker/machine/certs/key.pem \
  -out ~/.docker/machine/certs/cert.pfx 
```

### Sumbit a new cluster state object

The following command submits a cluster state object named nginx and will ensure 5
containers are started from the `nginx:1.9.6` Docker image.

```
$ curl -k https://127.0.0.1:2476/submit \
  -d '{"Name": "nginx", "Image": "nginx:1.9.6", "Count": 5}' \
  --cert ~/.docker/machine/certs/cert.pfx:linux
```

### Get the current status

The following command retrieves the cluster status from the Swarm cluster state manager.

```
$ curl -k https://127.0.0.1:2476/status \
  --cert ~/.docker/machine/certs/cert.pfx:linux
```
