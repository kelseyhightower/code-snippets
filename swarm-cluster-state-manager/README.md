# Docker Swarm Cluster State Manager

This is not an official Google product.

## Overview

The Swarm cluster state manager is a simple prototype that demonstrates how cluster state could be managed for Docker Swarm. The ideas for this prototype are based on the Kubernetes pod replication controller.

## Usage

```
$ swarm-cluster-state-manager -h
Usage of swarm-cluster-state-manager:
  -addr string
    	HTTP listen address (default "127.0.0.1:2476")
  -insecure-skip-verify
    	Skip server certificate verification
  -swarm-manager string
    	Docker Swarm manager address (default "tcp://127.0.0.1:2376")
  -tlscacert string
    	Trust certs signed only by this CA
  -tlscert string
    	Path to TLS certificate file
  -tlskey string
    	Path to TLS key file
```

### Start the swarm cluster state manager

By default `swarm-cluster-state-manager` will listen for remote connections on https://127.0.0.1:2476.
TLS client authentication is required, see next section.

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

On OS X you'll need to create a pkcs12 bundle of the PEM encoded client certs in order
to use them with cURL.

```
$ openssl pkcs12 -export \
  -in ~/.docker/machine/certs/cert.pem \
  -inkey ~/.docker/machine/certs/key.pem \
  -out ~/.docker/machine/certs/cert.pfx 
```

At the follow prompt set the export password to protect the cert and key:

```
Enter Export Password:
Verifying - Enter Export Password:
```

Later examples assume the word `swarm` was used for the password.

### Submit a new cluster state object

The following command submits a cluster state object named nginx and will ensure 5
containers are started from the `nginx:1.9.6` Docker image.

```
$ curl -k https://127.0.0.1:2476/submit \
  -d '{"Name": "nginx", "Image": "nginx:1.9.6", "Count": 5}' \
  --cert ~/.docker/machine/certs/cert.pfx
  --pass swarm
```

### Get the current status

The following command retrieves the cluster status from the Swarm cluster state manager.

```
$ curl -k https://127.0.0.1:2476/status \
  --cert ~/.docker/machine/certs/cert.pfx
  --pass swarm
```
